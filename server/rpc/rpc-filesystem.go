package rpc

import (
	"time"

	implantpb "github.com/bishopfox/sliver/protobuf/implant"
	"github.com/bishopfox/sliver/server/core"

	"github.com/golang/protobuf/proto"
)

func rpcLs(req []byte, timeout time.Duration, resp RPCResponse) {
	dirList := &implantpb.LsReq{}
	err := proto.Unmarshal(req, dirList)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := core.Hive.Sliver(dirList.SliverID)

	data, _ := proto.Marshal(&implantpb.LsReq{
		Path: dirList.Path,
	})
	data, err = sliver.Request(implantpb.MsgLsReq, timeout, data)
	resp(data, err)
}

func rpcRm(req []byte, timeout time.Duration, resp RPCResponse) {
	rmReq := &implantpb.RmReq{}
	err := proto.Unmarshal(req, rmReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := core.Hive.Sliver(rmReq.SliverID)

	data, _ := proto.Marshal(&implantpb.RmReq{
		Path: rmReq.Path,
	})
	data, err = sliver.Request(implantpb.MsgRmReq, timeout, data)
	resp(data, err)
}

func rpcMkdir(req []byte, timeout time.Duration, resp RPCResponse) {
	mkdirReq := &implantpb.MkdirReq{}
	err := proto.Unmarshal(req, mkdirReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := core.Hive.Sliver(mkdirReq.SliverID)

	data, _ := proto.Marshal(&implantpb.MkdirReq{
		Path: mkdirReq.Path,
	})
	data, err = sliver.Request(implantpb.MsgMkdirReq, timeout, data)
	resp(data, err)
}

func rpcCd(req []byte, timeout time.Duration, resp RPCResponse) {
	cdReq := &implantpb.CdReq{}
	err := proto.Unmarshal(req, cdReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := core.Hive.Sliver(cdReq.SliverID)

	data, _ := proto.Marshal(&implantpb.CdReq{
		Path: cdReq.Path,
	})
	data, err = sliver.Request(implantpb.MsgCdReq, timeout, data)
	resp(data, err)
}

func rpcPwd(req []byte, timeout time.Duration, resp RPCResponse) {
	pwdReq := &implantpb.PwdReq{}
	err := proto.Unmarshal(req, pwdReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := (*core.Hive.Slivers)[pwdReq.SliverID]

	data, _ := proto.Marshal(&implantpb.PwdReq{})
	data, err = sliver.Request(implantpb.MsgPwdReq, timeout, data)
	resp(data, err)
}

func rpcDownload(req []byte, timeout time.Duration, resp RPCResponse) {
	downloadReq := &implantpb.DownloadReq{}
	err := proto.Unmarshal(req, downloadReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := core.Hive.Sliver(downloadReq.SliverID)

	data, _ := proto.Marshal(&implantpb.DownloadReq{
		Path: downloadReq.Path,
	})
	data, err = sliver.Request(implantpb.MsgDownloadReq, timeout, data)
	resp(data, err)
}

func rpcUpload(req []byte, timeout time.Duration, resp RPCResponse) {
	uploadReq := &implantpb.UploadReq{}
	err := proto.Unmarshal(req, uploadReq)
	if err != nil {
		resp([]byte{}, err)
		return
	}
	sliver := core.Hive.Sliver(uploadReq.SliverID)

	data, _ := proto.Marshal(&implantpb.UploadReq{
		Encoder: uploadReq.Encoder,
		Path:    uploadReq.Path,
		Data:    uploadReq.Data,
	})
	data, err = sliver.Request(implantpb.MsgUploadReq, timeout, data)
	resp(data, err)
}
