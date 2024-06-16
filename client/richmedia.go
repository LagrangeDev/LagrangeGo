package client

import (
	"bytes"
	"encoding/hex"
	"errors"

	highway2 "github.com/LagrangeDev/LagrangeGo/client/internal/highway"
	"github.com/LagrangeDev/LagrangeGo/client/packets/oidb"
	message2 "github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/highway"
	oidb2 "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

func oidbIPv4ToNTHighwayIPv4(ipv4s []*oidb2.IPv4) []*highway.NTHighwayIPv4 {
	hwipv4s := make([]*highway.NTHighwayIPv4, len(ipv4s))
	for i, ip := range ipv4s {
		hwipv4s[i] = &highway.NTHighwayIPv4{
			Domain: &highway.NTHighwayDomain{
				IsEnable: true,
				IP:       binary.UInt32ToIPV4Address(ip.OutIP),
			},
			Port: ip.OutPort,
		}
	}
	return hwipv4s
}

func (c *QQClient) ImageUploadPrivate(targetUid string, element message.IMessageElement) (*message.ImageElement, error) {
	image, ok := element.(*message.ImageElement)
	if !ok {
		return nil, errors.New("element type is not friend image")
	}
	req, err := oidb.BuildPrivateImageUploadReq(targetUid, image)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	uploadResp, err := oidb.ParsePrivateImageUploadResp(resp)
	if err != nil {
		return nil, err
	}
	ukey := uploadResp.Upload.UKey.Unwrap()
	c.debugln("private image upload ukey:", ukey)
	if ukey != "" {
		index := uploadResp.Upload.MsgInfo.MsgInfoBody[0].Index
		sha1hash, err := hex.DecodeString(index.Info.FileSha1)
		if err != nil {
			return nil, err
		}
		extend := &highway.NTV2RichMediaHighwayExt{
			FileUuid: index.FileUuid,
			UKey:     ukey,
			Network: &highway.NTHighwayNetwork{
				IPv4S: oidbIPv4ToNTHighwayIPv4(uploadResp.Upload.IPv4S),
			},
			MsgInfoBody: uploadResp.Upload.MsgInfo.MsgInfoBody,
			BlockSize:   uint32(highway2.BlockSize),
			Hash: &highway.NTHighwayHash{
				FileSha1: [][]byte{sha1hash},
			},
		}
		extStream, err := proto.Marshal(extend)
		if err != nil {
			return nil, err
		}
		md5hash, err := hex.DecodeString(index.Info.FileHash)
		if err != nil {
			return nil, err
		}
		err = c.highwayUpload(1003,
			bytes.NewReader(image.Stream), uint64(len(image.Stream)),
			md5hash, extStream,
		)
		if err != nil {
			return nil, err
		}
	}
	image.MsgInfo = uploadResp.Upload.MsgInfo
	compatImage := &message2.NotOnlineImage{}
	err = proto.Unmarshal(uploadResp.Upload.CompatQMsg, compatImage)
	if err != nil {
		return nil, err
	}
	image.CompatImage = compatImage
	return image, nil
}

func (c *QQClient) ImageUploadGroup(groupUin uint32, element message.IMessageElement) (*message.ImageElement, error) {
	image, ok := element.(*message.ImageElement)
	if !ok {
		return nil, errors.New("element type is not group image")
	}
	req, err := oidb.BuildGroupImageUploadReq(groupUin, image)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	uploadResp, err := oidb.ParseGroupImageUploadResp(resp)
	if err != nil {
		return nil, err
	}
	ukey := uploadResp.Upload.UKey.Unwrap()
	c.debugln("group image upload ukey:", ukey)
	if ukey != "" {
		index := uploadResp.Upload.MsgInfo.MsgInfoBody[0].Index
		sha1hash, err := hex.DecodeString(index.Info.FileSha1)
		if err != nil {
			return nil, err
		}
		extend := &highway.NTV2RichMediaHighwayExt{
			FileUuid: index.FileUuid,
			UKey:     ukey,
			Network: &highway.NTHighwayNetwork{
				IPv4S: oidbIPv4ToNTHighwayIPv4(uploadResp.Upload.IPv4S),
			},
			MsgInfoBody: uploadResp.Upload.MsgInfo.MsgInfoBody,
			BlockSize:   uint32(highway2.BlockSize),
			Hash: &highway.NTHighwayHash{
				FileSha1: [][]byte{sha1hash},
			},
		}
		extStream, err := proto.Marshal(extend)
		if err != nil {
			return nil, err
		}
		md5hash, err := hex.DecodeString(index.Info.FileHash)
		if err != nil {
			return nil, err
		}
		err = c.highwayUpload(1004,
			bytes.NewReader(image.Stream), uint64(len(image.Stream)),
			md5hash, extStream,
		)
		if err != nil {
			return nil, err
		}
	}
	image.MsgInfo = uploadResp.Upload.MsgInfo
	_ = proto.Unmarshal(uploadResp.Upload.CompatQMsg, image.CompatFace)
	return image, nil
}

func (c *QQClient) RecordUploadPrivate(targetUid string, element message.IMessageElement) (*message.VoiceElement, error) {
	record, ok := element.(*message.VoiceElement)
	if !ok {
		return nil, errors.New("element type is not friend record")
	}
	req, err := oidb.BuildPrivateRecordUploadReq(targetUid, record)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	uploadResp, err := oidb.ParsePrivateRecordUploadResp(resp)
	if err != nil {
		return nil, err
	}
	ukey := uploadResp.Upload.UKey.Unwrap()
	c.debugln("private record upload ukey:", ukey)
	if ukey != "" {
		index := uploadResp.Upload.MsgInfo.MsgInfoBody[0].Index
		sha1hash, err := hex.DecodeString(index.Info.FileSha1)
		if err != nil {
			return nil, err
		}
		extend := &highway.NTV2RichMediaHighwayExt{
			FileUuid: index.FileUuid,
			UKey:     ukey,
			Network: &highway.NTHighwayNetwork{
				IPv4S: oidbIPv4ToNTHighwayIPv4(uploadResp.Upload.IPv4S),
			},
			MsgInfoBody: uploadResp.Upload.MsgInfo.MsgInfoBody,
			BlockSize:   uint32(highway2.BlockSize),
			Hash: &highway.NTHighwayHash{
				FileSha1: [][]byte{sha1hash},
			},
		}
		extStream, err := proto.Marshal(extend)
		if err != nil {
			return nil, err
		}
		md5hash, err := hex.DecodeString(index.Info.FileHash)
		if err != nil {
			return nil, err
		}
		err = c.highwayUpload(1007,
			bytes.NewReader(record.Data), uint64(len(record.Data)),
			md5hash, extStream,
		)
		if err != nil {
			return nil, err
		}
	}
	record.MsgInfo = uploadResp.Upload.MsgInfo
	record.Compat = uploadResp.Upload.CompatQMsg
	return record, nil
}

func (c *QQClient) RecordUploadGroup(groupUin uint32, element message.IMessageElement) (*message.VoiceElement, error) {
	record, ok := element.(*message.VoiceElement)
	if !ok {
		return nil, errors.New("element type is not voice record")
	}
	req, err := oidb.BuildGroupRecordUploadReq(groupUin, record)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	uploadResp, err := oidb.ParseGroupRecordUploadResp(resp)
	if err != nil {
		return nil, err
	}
	ukey := uploadResp.Upload.UKey.Unwrap()
	c.debugln("group record upload ukey:", ukey)
	if ukey != "" {
		index := uploadResp.Upload.MsgInfo.MsgInfoBody[0].Index
		sha1hash, err := hex.DecodeString(index.Info.FileSha1)
		if err != nil {
			return nil, err
		}
		extend := &highway.NTV2RichMediaHighwayExt{
			FileUuid: index.FileUuid,
			UKey:     ukey,
			Network: &highway.NTHighwayNetwork{
				IPv4S: oidbIPv4ToNTHighwayIPv4(uploadResp.Upload.IPv4S),
			},
			MsgInfoBody: uploadResp.Upload.MsgInfo.MsgInfoBody,
			BlockSize:   uint32(highway2.BlockSize),
			Hash: &highway.NTHighwayHash{
				FileSha1: [][]byte{sha1hash},
			},
		}
		extStream, err := proto.Marshal(extend)
		if err != nil {
			return nil, err
		}
		md5hash, err := hex.DecodeString(index.Info.FileHash)
		if err != nil {
			return nil, err
		}
		err = c.highwayUpload(1008,
			bytes.NewReader(record.Data), uint64(len(record.Data)),
			md5hash, extStream,
		)
		if err != nil {
			return nil, err
		}
	}
	record.MsgInfo = uploadResp.Upload.MsgInfo
	record.Compat = uploadResp.Upload.CompatQMsg
	return record, nil
}
