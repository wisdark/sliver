package rpc

import (
	"time"

	"github.com/bishopfox/sliver/server/core"

	clientpb "github.com/bishopfox/sliver/protobuf/client"
	implantpb "github.com/bishopfox/sliver/protobuf/implant"

	"github.com/golang/protobuf/proto"
)

const (
	tunDefaultTimeout = 30 * time.Second
)

func tunnelCreate(client *core.Client, req []byte, resp RPCResponse) {
	tunCreateReq := &clientpb.TunnelCreateReq{}
	proto.Unmarshal(req, tunCreateReq)

	tunnel := core.Tunnels.CreateTunnel(client, tunCreateReq.SliverID)

	data, err := proto.Marshal(&clientpb.TunnelCreate{
		SliverID: tunnel.Sliver.ID,
		TunnelID: tunnel.ID,
	})

	resp(data, err)
}

func tunnelData(client *core.Client, req []byte, _ RPCResponse) {
	tunnelData := &implantpb.TunnelData{}
	proto.Unmarshal(req, tunnelData)
	tunnel := core.Tunnels.Tunnel(tunnelData.TunnelID)
	if tunnel != nil && client.ID == tunnel.Client.ID {
		tunnel.Sliver.Request(implantpb.MsgTunnelData, tunDefaultTimeout, req)
	} else {
		rpcLog.Warnf("Data sent on nil tunnel %d", tunnelData.TunnelID)
	}
}

func tunnelClose(client *core.Client, req []byte, resp RPCResponse) {
	tunCloseReq := &clientpb.TunnelCloseReq{}
	proto.Unmarshal(req, tunCloseReq)

	tunnel := core.Tunnels.Tunnel(tunCloseReq.TunnelID)

	if tunnel != nil && client.ID == tunnel.Client.ID {
		closed := core.Tunnels.CloseTunnel(tunCloseReq.TunnelID, "Client exit")
		closeResp := &implantpb.TunnelClose{
			TunnelID: tunCloseReq.TunnelID,
		}
		if !closed {
			closeResp.Err = "Failed to close tunnel"
		}
		data, err := proto.Marshal(closeResp)
		resp(data, err)
	}
}
