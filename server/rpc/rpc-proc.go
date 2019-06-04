package rpc

import (
	"time"

	"github.com/bishopfox/sliver/server/core"

	implantpb "github.com/bishopfox/sliver/protobuf/implant"

	"github.com/golang/protobuf/proto"
)

func rpcPs(req []byte, timeout time.Duration, resp RPCResponse) {
	psReq := &implantpb.PsReq{}
	err := proto.Unmarshal(req, psReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := (*core.Hive.Slivers)[psReq.SliverID]
	if sliver == nil {
		resp([]byte{}, err)
		return
	}

	data, _ := proto.Marshal(&implantpb.PsReq{})
	data, err = sliver.Request(implantpb.MsgPsReq, timeout, data)
	resp(data, err)
}

func rpcProcdump(req []byte, timeout time.Duration, resp RPCResponse) {
	procdumpReq := &implantpb.ProcessDumpReq{}
	err := proto.Unmarshal(req, procdumpReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := (*core.Hive.Slivers)[procdumpReq.SliverID]
	if sliver == nil {
		resp([]byte{}, err)
		return
	}
	data, _ := proto.Marshal(&implantpb.ProcessDumpReq{
		Pid: procdumpReq.Pid,
	})

	data, err = sliver.Request(implantpb.MsgProcessDumpReq, timeout, data)
	resp(data, err)
}
