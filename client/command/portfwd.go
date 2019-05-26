package command

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/bishopfox/sliver/client/core"

	"github.com/desertbit/grumble"
)

func portfwd(ctx *grumble.Context, server *core.SliverServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	remoteHost := ctx.Flags.String("remote-host")
	if remoteHost == "" {
		fmt.Printf(Warn + "You must provide a remote host with `-r`\n")
		return
	}
	remotePort := ctx.Flags.Uint("remote-port")
	if remotePort == 0 {
		fmt.Printf(Warn + "You must provide a remote port with `-p`\n")
		return
	}
	localPort := ctx.Flags.Uint("local-port")
	if localPort == 0 {
		fmt.Printf(Warn + "You must provide a local port with `-l`\n")
		return
	}
	localHost := ctx.Flags.String("local-host")
	// Start local listener
	err := server.StartForwardListener(localHost, localPort, remoteHost, remotePort, ActiveSliver.Sliver.ID)
	if err != nil {
		fmt.Printf(Warn+"%s\n", err.Error())
		return
	}
	fmt.Printf(Info+"Added port forwarding rule: %s:%d -> %s:%d\n", localHost, localPort, remoteHost, remotePort)
}

func portfwdList(ctx *grumble.Context, server *core.SliverServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	if len(server.Forwarders.Forwarders) > 0 {
		table := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		fmt.Fprintf(table, "ID\tLocal Host\tLocal Port\tRemote Host\tRemote Port\t\n")
		fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t\n",
			strings.Repeat("=", len("ID")),
			strings.Repeat("=", len("Local Host")),
			strings.Repeat("=", len("Local Port")),
			strings.Repeat("=", len("Remote Host")),
			strings.Repeat("=", len("Remote Port")))

		for fid, fwd := range server.Forwarders.Forwarders {
			fmt.Fprintf(table, "%d\t%s\t%d\t%s\t%d\t\n", fid, fwd.LHost, fwd.LPort, fwd.RHost, fwd.RPort)
		}
		table.Flush()
	}
}

func portfwdDelete(ctx *grumble.Context, server *core.SliverServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	if len(ctx.Args) != 1 {
		fmt.Printf(Warn + "You must provide a rule identifier\n")
		return
	}
	arg := ctx.Args[0]
	fid, err := strconv.ParseUint(arg, 10, 32)
	if err != nil {
		fmt.Printf(Warn+"%s", err.Error())
		return
	}
	err = server.DeleteForwarder(uint32(fid))
	if err != nil {
		fmt.Printf(Warn+"%s", err.Error())
		return
	}
	fmt.Printf(Info+"Successfully removed forward rule %d\n", fid)
}
