package client

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net/netip"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/packets/oidb"
	message2 "github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/highway"
	oidb2 "github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

func le32toipstr(raw uint32) string {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], raw)
	return netip.AddrFrom4(buf).String()
}

func ConvertNTHighwayNetWork(ipv4s []*oidb2.IPv4) []*highway.NTHighwayIPv4 {
	hwipv4s := make([]*highway.NTHighwayIPv4, len(ipv4s))
	for i, ip := range ipv4s {
		hwipv4s[i] = &highway.NTHighwayIPv4{
			Domain: &highway.NTHighwayDomain{
				IsEnable: true,
				IP:       le32toipstr(ip.OutIP),
			},
			Port: ip.OutPort,
		}
	}
	return hwipv4s
}

func (c *QQClient) ImageUploadPrivate(targetUid string, element message.IMessageElement) (*message.FriendImageElement, error) {
	image, ok := element.(*message.FriendImageElement)
	if !ok {
		return nil, errors.New("element type is not friend image")
	}
	req, err := oidb.BuildImageUploadReq(targetUid, image.Stream)
	if err != nil {
		return nil, err
	}
	resp, err := c.SendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	uploadResp, err := oidb.ParseImageUploadResp(resp.Data)
	if err != nil {
		return nil, err
	}
	ukey := uploadResp.Upload.UKey
	if ukey.Unwrap() != "" {
		index := uploadResp.Upload.MsgInfo.MsgInfoBody[0].Index
		sha1hash, err := hex.DecodeString(index.Info.FileSha1)
		if err != nil {
			return nil, err
		}
		extend := &highway.NTV2RichMediaHighwayExt{
			FileUuid: index.FileUuid,
			UKey:     ukey.Unwrap(),
			Network: &highway.NTHighwayNetwork{
				IPv4S: ConvertNTHighwayNetWork(uploadResp.Upload.IPv4S),
			},
			MsgInfoBody: uploadResp.Upload.MsgInfo.MsgInfoBody,
			BlockSize:   1024 * 1024,
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
		err = c.UploadSrcByStream(1003,
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

func (c *QQClient) ImageUploadGroup(groupUin uint32, element message.IMessageElement) (*message.GroupImageElement, error) {
	image, ok := element.(*message.GroupImageElement)
	if !ok {
		return nil, errors.New("element type is not group image")
	}
	req, err := oidb.BuildGroupImageUploadReq(groupUin, image.Stream)
	if err != nil {
		return nil, err
	}
	resp, err := c.SendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	uploadResp, err := oidb.ParseGroupImageUploadResp(resp.Data)
	if err != nil {
		return nil, err
	}
	ukey := uploadResp.Upload.UKey
	if ukey.Unwrap() != "" {
		index := uploadResp.Upload.MsgInfo.MsgInfoBody[0].Index
		sha1hash, err := hex.DecodeString(index.Info.FileSha1)
		if err != nil {
			return nil, err
		}
		extend := &highway.NTV2RichMediaHighwayExt{
			FileUuid: index.FileUuid,
			UKey:     ukey.Unwrap(),
			Network: &highway.NTHighwayNetwork{
				IPv4S: ConvertNTHighwayNetWork(uploadResp.Upload.IPv4S),
			},
			MsgInfoBody: uploadResp.Upload.MsgInfo.MsgInfoBody,
			BlockSize:   1024 * 1024,
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
		err = c.UploadSrcByStream(1004,
			bytes.NewReader(image.Stream), uint64(len(image.Stream)),
			md5hash, extStream,
		)
		if err != nil {
			return nil, err
		}
	}
	image.MsgInfo = uploadResp.Upload.MsgInfo
	image.CompatFace = uploadResp.Upload.CompatQMsg
	return image, nil
}
