package rpc

import (
	"time"

	"github.com/bishopfox/sliver/server/core"
	"github.com/bishopfox/sliver/server/log"

	clientpb "github.com/bishopfox/sliver/protobuf/client"
	implantpb "github.com/bishopfox/sliver/protobuf/implant"
)

var (
	rpcLog = log.NamedLogger("rpc", "server")
)

// RPCResponse - Called with response data, mapped back to reqID
type RPCResponse func([]byte, error)

// RPCHandler - RPC handlers accept bytes and return bytes
type RPCHandler func([]byte, time.Duration, RPCResponse)
type TunnelHandler func(*core.Client, []byte, RPCResponse)

var (
	rpcHandlers = &map[uint32]RPCHandler{
		clientpb.MsgJobs:    rpcJobs,
		clientpb.MsgJobKill: rpcJobKill,
		clientpb.MsgMtls:    rpcStartMTLSListener,
		clientpb.MsgDns:     rpcStartDNSListener,
		clientpb.MsgHttp:    rpcStartHTTPListener,
		clientpb.MsgHttps:   rpcStartHTTPSListener,

		clientpb.MsgWebsiteList:          rpcWebsiteList,
		clientpb.MsgWebsiteAddContent:    rpcWebsiteAddContent,
		clientpb.MsgWebsiteRemoveContent: rpcWebsiteRemoveContent,

		clientpb.MsgSessions:         rpcSessions,
		clientpb.MsgGenerate:         rpcGenerate,
		clientpb.MsgRegenerate:       rpcRegenerate,
		clientpb.MsgListSliverBuilds: rpcListSliverBuilds,
		clientpb.MsgListCanaries:     rpcListCanaries,
		clientpb.MsgProfiles:         rpcProfiles,
		clientpb.MsgNewProfile:       rpcNewProfile,
		clientpb.MsgPlayers:          rpcPlayers,

		clientpb.MsgMsf:       rpcMsf,
		clientpb.MsgMsfInject: rpcMsfInject,

		clientpb.MsgGetSystemReq: rpcGetSystem,

		// "Req"s directly map to responses
		implantpb.MsgPsReq:          rpcPs,
		implantpb.MsgKill:           rpcKill,
		implantpb.MsgProcessDumpReq: rpcProcdump,

		implantpb.MsgElevate:         rpcElevate,
		implantpb.MsgImpersonate:     rpcImpersonate,
		implantpb.MsgExecuteAssembly: rpcExecuteAssembly,

		implantpb.MsgLsReq:       rpcLs,
		implantpb.MsgRmReq:       rpcRm,
		implantpb.MsgMkdirReq:    rpcMkdir,
		implantpb.MsgCdReq:       rpcCd,
		implantpb.MsgPwdReq:      rpcPwd,
		implantpb.MsgDownloadReq: rpcDownload,
		implantpb.MsgUploadReq:   rpcUpload,

		implantpb.MsgShellReq: rpcShell,

		clientpb.MsgTask:    rpcLocalTask,
		clientpb.MsgMigrate: rpcMigrate,
	}

	tunHandlers = &map[uint32]TunnelHandler{
		clientpb.MsgTunnelCreate: tunnelCreate,
		implantpb.MsgTunnelData:  tunnelData,
		implantpb.MsgTunnelClose: tunnelClose,
	}
)

// GetRPCHandlers - Returns a map of server-side msg handlers
func GetRPCHandlers() *map[uint32]RPCHandler {
	return rpcHandlers
}

// GetTunnelHandlers - Returns a map of tunnel handlers
func GetTunnelHandlers() *map[uint32]TunnelHandler {
	return tunHandlers
}
