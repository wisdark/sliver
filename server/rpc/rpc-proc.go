package rpc

/*
	Sliver Implant Framework
	Copyright (C) 2019  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"time"

	"github.com/bishopfox/sliver/server/core"

	sliverpb "github.com/bishopfox/sliver/protobuf/sliver"

	"github.com/golang/protobuf/proto"
)

func rpcPs(req []byte, timeout time.Duration, resp RPCResponse) {
	psReq := &sliverpb.PsReq{}
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

	data, _ := proto.Marshal(&sliverpb.PsReq{})
	data, err = sliver.Request(sliverpb.MsgPsReq, timeout, data)
	resp(data, err)
}

func rpcProcdump(req []byte, timeout time.Duration, resp RPCResponse) {
	procdumpReq := &sliverpb.ProcessDumpReq{}
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
	data, _ := proto.Marshal(&sliverpb.ProcessDumpReq{
		Pid: procdumpReq.Pid,
	})

	data, err = sliver.Request(sliverpb.MsgProcessDumpReq, timeout, data)
	resp(data, err)
}

func rpcTerminate(req []byte, timeout time.Duration, resp RPCResponse) {
	terminateReq := &sliverpb.TerminateReq{}
	err := proto.Unmarshal(req, terminateReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := (*core.Hive.Slivers)[terminateReq.SliverID]
	if sliver == nil {
		resp([]byte{}, err)
		return
	}
	data, _ := proto.Marshal(&sliverpb.TerminateReq{Pid: terminateReq.GetPid()})
	data, err = sliver.Request(sliverpb.MsgTerminate, timeout, data)
	resp(data, err)
}
