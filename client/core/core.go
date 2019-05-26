package core

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"net"

	"github.com/bishopfox/sliver/client/assets"
	clientpb "github.com/bishopfox/sliver/protobuf/client"
	sliverpb "github.com/bishopfox/sliver/protobuf/sliver"

	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	randomIDSize   = 16 // 64bits
	forwardTimeout = 5 * time.Second
)

var fwdID uint32

type tunnels struct {
	server  *SliverServer
	tunnels *map[uint64]*tunnel
	mutex   *sync.RWMutex
}

func (t *tunnels) bindTunnel(SliverID uint32, TunnelID uint64) *tunnel {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	(*t.tunnels)[TunnelID] = &tunnel{
		server:   t.server,
		SliverID: SliverID,
		ID:       TunnelID,
		Recv:     make(chan []byte),
	}

	return (*t.tunnels)[TunnelID]
}

// RecvTunnelData - Routes a TunnelData protobuf msg to the correct tunnel object
func (t *tunnels) RecvTunnelData(tunnelData *sliverpb.TunnelData) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	tunnel := (*t.tunnels)[tunnelData.TunnelID]
	if tunnel != nil {
		(*tunnel).Recv <- tunnelData.Data
	} else {
		log.Printf("No client tunnel with ID %d", tunnelData.TunnelID)
	}
}

func (t *tunnels) Close(ID uint64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	close((*t.tunnels)[ID].Recv)
	delete(*t.tunnels, ID)
}

// tunnel - Duplex data tunnel
type tunnel struct {
	server   *SliverServer
	SliverID uint32
	ID       uint64
	Recv     chan []byte
}

func (t *tunnel) Send(data []byte) {
	log.Printf("Sending %d bytes on tunnel %d (sliver %d)", len(data), t.ID, t.SliverID)
	tunnelData := &sliverpb.TunnelData{
		SliverID: t.SliverID,
		TunnelID: t.ID,
		Data:     data,
	}
	rawTunnelData, _ := proto.Marshal(tunnelData)
	t.server.Send <- &sliverpb.Envelope{
		Type: sliverpb.MsgTunnelData,
		Data: rawTunnelData,
	}
}

// Forwarder stores the configuration for a local port forwarder
type Forwarder struct {
	listener *net.Listener
	SliverID uint32
	ID       uint32
	server   *SliverServer
	LHost    string
	LPort    uint
	RHost    string
	RPort    uint
}

func (f *Forwarder) listenAndForward() {
	for {
		conn, _ := (*f.listener).Accept()
		// Create new tunnel
		t, err := f.server.CreateTunnel(f.SliverID, forwardTimeout)
		if err != nil {
			return
		}
		data, _ := proto.Marshal(&sliverpb.PortFwdReq{
			SliverID: f.SliverID,
			Host:     f.RHost,
			Port:     uint32(f.RPort),
			TunnelID: t.ID,
		})
		resp := <-f.server.RPC(&sliverpb.Envelope{
			Type: sliverpb.MsgPortfwdReq,
			Data: data,
		}, forwardTimeout)
		if resp.Err != "" {
			break
		}
		cleanup := func() {
			tunnelClose, _ := proto.Marshal(&sliverpb.PortFwdReq{
				TunnelID: t.ID,
			})
			f.server.RPC(&sliverpb.Envelope{
				Type: sliverpb.MsgTunnelClose,
				Data: tunnelClose,
			}, forwardTimeout)
			conn.Close()
		}
		// Copy the incoming data from the tunnel into the net.Conn
		go func() {
			defer cleanup()
			for data := range t.Recv {
				conn.Write(data)
			}
		}()
		// Copy incoming data into the tunnel Read chan
		readBuf := make([]byte, 128)
		for {
			// In case the conn has been shutdown
			if conn != nil {
				n, err := conn.Read(readBuf)
				if err == io.EOF {
					break
				}
				if err == nil && 0 < n {
					t.Send(readBuf[:n])
				}
			}
		}
	}
}

// Forwarders stores the map of all port forwarders
type Forwarders struct {
	Forwarders map[uint32]*Forwarder
	server     *SliverServer
}

// SliverServer - Server info
type SliverServer struct {
	Send       chan *sliverpb.Envelope
	recv       chan *sliverpb.Envelope
	responses  *map[uint64]chan *sliverpb.Envelope
	mutex      *sync.RWMutex
	Config     *assets.ClientConfig
	Events     chan *clientpb.Event
	Tunnels    *tunnels
	Forwarders *Forwarders
}

// CreateTunnel - Create a new tunnel on the server, returns tunnel metadata
func (ss *SliverServer) CreateTunnel(sliverID uint32, defaultTimeout time.Duration) (*tunnel, error) {
	tunReq := &clientpb.TunnelCreateReq{SliverID: sliverID}
	tunReqData, _ := proto.Marshal(tunReq)

	tunResp := <-ss.RPC(&sliverpb.Envelope{
		Type: clientpb.MsgTunnelCreate,
		Data: tunReqData,
	}, defaultTimeout)
	if tunResp.Err != "" {
		return nil, fmt.Errorf("Error: %s", tunResp.Err)
	}

	tunnelCreated := &clientpb.TunnelCreate{}
	proto.Unmarshal(tunResp.Data, tunnelCreated)

	tunnel := ss.Tunnels.bindTunnel(tunnelCreated.SliverID, tunnelCreated.TunnelID)

	log.Printf("Created new tunnel with ID %d", tunnel.ID)

	return tunnel, nil
}

// ResponseMapper - Maps recv'd envelopes to response channels
func (ss *SliverServer) ResponseMapper() {
	for envelope := range ss.recv {
		if envelope.ID != 0 {
			ss.mutex.Lock()
			if resp, ok := (*ss.responses)[envelope.ID]; ok {
				resp <- envelope
			}
			ss.mutex.Unlock()
		} else {
			// If the message does not have an envelope ID then we route it based on type
			switch envelope.Type {

			case clientpb.MsgEvent:
				event := &clientpb.Event{}
				err := proto.Unmarshal(envelope.Data, event)
				if err != nil {
					log.Printf("Failed to decode event envelope")
					continue
				}
				// log.Printf("[client] Routing event message")
				ss.Events <- event

			case sliverpb.MsgTunnelData:
				tunnelData := &sliverpb.TunnelData{}
				err := proto.Unmarshal(envelope.Data, tunnelData)
				if err != nil {
					log.Printf("Failed to decode tunnel data envelope")
					continue
				}
				// log.Printf("[client] Routing tunnel data with id %d", tunnelData.TunnelID)
				ss.Tunnels.RecvTunnelData(tunnelData)

			case sliverpb.MsgTunnelClose:
				tunnelClose := &sliverpb.TunnelClose{}
				err := proto.Unmarshal(envelope.Data, tunnelClose)
				if err != nil {
					log.Printf("Failed to decode tunnel data envelope")
					continue
				}
				ss.Tunnels.Close(tunnelClose.TunnelID)

			}
		}
	}
}

// RPC - Send a request envelope and wait for a response (blocking)
func (ss *SliverServer) RPC(envelope *sliverpb.Envelope, timeout time.Duration) chan *sliverpb.Envelope {
	reqID := EnvelopeID()
	envelope.ID = reqID
	envelope.Timeout = timeout.Nanoseconds()
	resp := make(chan *sliverpb.Envelope)
	ss.AddRespListener(reqID, resp)
	ss.Send <- envelope
	respCh := make(chan *sliverpb.Envelope)
	go func() {
		defer ss.RemoveRespListener(reqID)
		select {
		case respEnvelope := <-resp:
			respCh <- respEnvelope
		case <-time.After(timeout + time.Second):
			respCh <- &sliverpb.Envelope{Err: "Timeout"}
		}
	}()
	return respCh
}

// AddRespListener - Add a response listener
func (ss *SliverServer) AddRespListener(envelopeID uint64, resp chan *sliverpb.Envelope) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	(*ss.responses)[envelopeID] = resp
}

// RemoveRespListener - Remove a listener
func (ss *SliverServer) RemoveRespListener(envelopeID uint64) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	close((*ss.responses)[envelopeID])
	delete((*ss.responses), envelopeID)
}

// StartForwardListener - Starts a local TCP listener to be used by the port forwarder
func (ss *SliverServer) StartForwardListener(localhost string, localport uint, remotehost string, remoteport uint, sliverID uint32) error {
	lst, err := net.Listen("tcp", fmt.Sprintf("%s:%d", localhost, localport))
	if err != nil {
		return err
	}
	fwd := &Forwarder{
		server:   ss,
		SliverID: sliverID,
		listener: &lst,
		ID:       nextForwarderID(),
		LHost:    localhost,
		LPort:    localport,
		RHost:    remotehost,
		RPort:    remoteport,
	}
	ss.addForwarder(fwd)
	go fwd.listenAndForward()
	return nil
}

// GetForwarder retrieves a port forwarder based on its identifier
func (ss *SliverServer) GetForwarder(fID uint32) (*Forwarder, error) {
	fwd, ok := ss.Forwarders.Forwarders[fID]
	if ok {
		return fwd, nil
	}
	return nil, fmt.Errorf("no forwarder found for id %d", fID)
}

func (ss *SliverServer) addForwarder(fwd *Forwarder) {
	ss.Forwarders.Forwarders[fwdID] = fwd
	fwdID++
}

// DeleteForwarder stops and remove a forwarder from the list
func (ss *SliverServer) DeleteForwarder(fid uint32) error {
	if fwd, ok := ss.Forwarders.Forwarders[fid]; ok {
		(*fwd.listener).Close()
		delete(ss.Forwarders.Forwarders, fid)
		log.Printf("killed port forwarder %d\n", fid)
		return nil
	}
	return fmt.Errorf("no forwarder found with id %d", fid)
}

// BindSliverServer - Bind send/recv channels to a server
func BindSliverServer(send, recv chan *sliverpb.Envelope) *SliverServer {
	server := &SliverServer{
		Send:      send,
		recv:      recv,
		responses: &map[uint64]chan *sliverpb.Envelope{},
		mutex:     &sync.RWMutex{},
		Events:    make(chan *clientpb.Event, 1),
	}
	server.Tunnels = &tunnels{
		server:  server,
		tunnels: &map[uint64]*tunnel{},
		mutex:   &sync.RWMutex{},
	}
	server.Forwarders = &Forwarders{
		Forwarders: map[uint32]*Forwarder{},
		server:     server,
	}
	return server
}

// EnvelopeID - Generate random ID
func EnvelopeID() uint64 {
	randBuf := make([]byte, 8) // 64 bits of randomness
	rand.Read(randBuf)
	return binary.LittleEndian.Uint64(randBuf)
}

func nextForwarderID() uint32 {
	return uint32(fwdID + 1)
}
