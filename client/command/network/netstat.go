package network

/*
	Sliver Implant Framework
	Copyright (C) 2021  Bishop Fox

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
	"context"
	"fmt"
	"net"

	"github.com/bishopfox/sliver/client/command/settings"
	"github.com/bishopfox/sliver/client/console"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
	"github.com/desertbit/grumble"
	"github.com/jedib0t/go-pretty/v6/table"
	"google.golang.org/protobuf/proto"
)

// NetstatCmd - Display active network connections on the remote system
func NetstatCmd(ctx *grumble.Context, con *console.SliverConsoleClient) {
	session, beacon := con.ActiveTarget.GetInteractive()
	if session == nil && beacon == nil {
		return
	}

	listening := ctx.Flags.Bool("listen")
	ip4 := ctx.Flags.Bool("ip4")
	ip6 := ctx.Flags.Bool("ip6")
	tcp := ctx.Flags.Bool("tcp")
	udp := ctx.Flags.Bool("udp")

	implantPID := getPID(session, beacon)
	activeC2 := getActiveC2(session, beacon)

	netstat, err := con.Rpc.Netstat(context.Background(), &sliverpb.NetstatReq{
		Request:   con.ActiveTarget.Request(ctx),
		TCP:       tcp,
		UDP:       udp,
		Listening: listening,
		IP4:       ip4,
		IP6:       ip6,
	})
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}
	if netstat.Response != nil && netstat.Response.Async {
		con.AddBeaconCallback(netstat.Response.TaskID, func(task *clientpb.BeaconTask) {
			err = proto.Unmarshal(task.Response, netstat)
			if err != nil {
				con.PrintErrorf("Failed to decode response %s\n", err)
				return
			}
			PrintNetstat(netstat, implantPID, activeC2, con)
		})
		con.PrintAsyncResponse(netstat.Response)
	} else {
		PrintNetstat(netstat, implantPID, activeC2, con)
	}
}

func PrintNetstat(netstat *sliverpb.Netstat, implantPID int32, activeC2 string, con *console.SliverConsoleClient) {
	lookup := func(skaddr *sliverpb.SockTabEntry_SockAddr) string {
		const IPv4Strlen = 17
		addr := skaddr.Ip
		names, err := net.LookupAddr(addr)
		if err == nil && len(names) > 0 {
			addr = names[0]
		}
		if len(addr) > IPv4Strlen {
			addr = addr[:IPv4Strlen]
		}
		return fmt.Sprintf("%s:%d", addr, skaddr.Port)
	}

	tw := table.NewWriter()
	tw.SetStyle(settings.GetTableStyle(con))
	tw.AppendHeader(table.Row{"Protocol", "Local Address", "Foreign Address", "State", "PID/Program name"})

	for _, entry := range netstat.Entries {
		pid := ""
		if entry.Process != nil {
			pid = fmt.Sprintf("%d/%s", entry.Process.Pid, entry.Process.Executable)
		}
		srcAddr := lookup(entry.LocalAddr)
		dstAddr := lookup(entry.RemoteAddr)
		if entry.Process != nil && entry.Process.Pid == implantPID {
			tw.AppendRow(table.Row{
				fmt.Sprintf(console.Green+"%s"+console.Normal, entry.Protocol),
				fmt.Sprintf(console.Green+"%s"+console.Normal, srcAddr),
				fmt.Sprintf(console.Green+"%s"+console.Normal, dstAddr),
				fmt.Sprintf(console.Green+"%s"+console.Normal, entry.SkState),
				fmt.Sprintf(console.Green+"%s"+console.Normal, pid),
			})
		} else {
			tw.AppendRow(table.Row{entry.Protocol, srcAddr, dstAddr, entry.SkState, pid})
		}
	}
	con.Printf("%s\n", tw.Render())
}

func getActiveC2(session *clientpb.Session, beacon *clientpb.Beacon) string {
	if session != nil {
		return session.ActiveC2
	}
	if beacon != nil {
		return beacon.ActiveC2
	}
	return ""
}

func getPID(session *clientpb.Session, beacon *clientpb.Beacon) int32 {
	if session != nil {
		return session.PID
	}
	if beacon != nil {
		return beacon.PID
	}
	return -1
}
