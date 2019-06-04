package handlers

import (
	"github.com/bishopfox/sliver/implant/transports"
	pb "github.com/bishopfox/sliver/protobuf/implant"
)

type RPCResponse func([]byte, error)
type RPCHandler func([]byte, RPCResponse)
type SpecialHandler func([]byte, *transports.Connection) error
type TunnelHandler func(*pb.Envelope, *transports.Connection)
