package rpc

import (
	"fmt"
	"time"

	"github.com/bishopfox/sliver/server/core"

	sliverpb "github.com/bishopfox/sliver/protobuf/sliver"

	"github.com/golang/protobuf/proto"
)

func rpcPortfwd(req []byte, timeout time.Duration, resp RPCResponse) {
	pfwdReq := &sliverpb.PortFwdReq{}
	proto.Unmarshal(req, pfwdReq)

	sliver := core.Hive.Sliver(pfwdReq.SliverID)
	tunnel := core.Tunnels.Tunnel(pfwdReq.TunnelID)

	startPortFwdReq, err := proto.Marshal(&sliverpb.PortFwdReq{
		Host:     pfwdReq.Host,
		Port:     pfwdReq.Port,
		SliverID: sliver.ID,
		TunnelID: tunnel.ID,
	})
	if err != nil {
		resp([]byte{}, err)
		return
	}
	rpcLog.Info(fmt.Sprintf("Requesting Sliver %d to start a forward rule to %s:%d", sliver.ID, pfwdReq.Host, pfwdReq.Port))
	data, err := sliver.Request(sliverpb.MsgPortfwdReq, timeout, startPortFwdReq)
	resp(data, err)
}
