package rpc

/*
	Sliver Implant Framework
	Copyright (C) 2022  Bishop Fox

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
	"regexp"

	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
	"github.com/bishopfox/sliver/server/core"
	"github.com/bishopfox/sliver/server/db"
)

const maxNameLength = 32

// Reconfigure - Reconfigure a beacon/session
func (rpc *Server) Reconfigure(ctx context.Context, req *sliverpb.ReconfigureReq) (*sliverpb.Reconfigure, error) {
	resp := &sliverpb.Reconfigure{Response: &commonpb.Response{}}
	err := rpc.GenericHandler(req, resp)
	if err != nil {
		return nil, err
	}

	// Successfully execute command, update server's info on reconnect interval
	if req.Request.SessionID != "" {
		session := core.Sessions.Get(req.Request.SessionID)
		if req.ReconnectInterval != 0 {
			session.ReconnectInterval = req.ReconnectInterval
		}
	}
	return resp, nil
}

// Rename - Rename a beacon/session
func (rpc *Server) Rename(ctx context.Context, req *clientpb.RenameReq) (*commonpb.Empty, error) {
	resp := &commonpb.Empty{}

	if len(req.Name) < 1 || maxNameLength < len(req.Name) {
		return resp, ErrInvalidName
	}
	if !regexp.MustCompile(`^[[:alnum:]]+$`).MatchString(req.Name) {
		return resp, ErrInvalidName
	}

	if req.SessionID != "" {
		session := core.Sessions.Get(req.SessionID)
		if session == nil {
			return nil, ErrInvalidSessionID
		}
		session.Name = req.Name
	} else if req.BeaconID != "" {
		beacon, err := db.BeaconByID(req.BeaconID)
		if err != nil || beacon == nil {
			return nil, ErrInvalidBeaconID
		}
		err = db.RenameBeacon(beacon.ID.String(), req.Name)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, ErrMissingRequestField
	}
	return resp, nil
}
