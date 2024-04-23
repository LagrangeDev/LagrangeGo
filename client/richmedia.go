package client

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/packets/oidb"
	message2 "github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/highway"
	oidb2 "github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

func ConvertIP(raw uint32) string {
	ip := make(net.IP, 4)
	ip[0] = byte(raw & 0xFF)
	ip[1] = byte((raw >> 8) & 0xFF)
	ip[2] = byte((raw >> 16) & 0xFF)
	ip[3] = byte((raw >> 24) & 0xFF)
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func ConvertNTHighwayNetWork(ipv4s []*oidb2.IPv4) []*highway.NTHighwayIPv4 {
	var IPv4s []*highway.NTHighwayIPv4
	for _, ipv4 := range ipv4s {
		IPv4s = append(IPv4s, &highway.NTHighwayIPv4{
			Domain: &highway.NTHighwayDomain{
				IsEnable: true,
				IP:       ConvertIP(ipv4.OutIP),
			},
			Port: ipv4.OutPort,
		})
	}
	return IPv4s
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
		sigSession, _ := c.GetServiceServer()
		if !c.UploadSrcByStreamAsync(1003, io.ReadSeeker(bytes.NewReader(image.Stream)), sigSession, md5hash, extStream) {
			return nil, errors.New("upload failed")
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
		sigSession, _ := c.GetServiceServer()
		if !c.UploadSrcByStreamAsync(1004, io.ReadSeeker(bytes.NewReader(image.Stream)), sigSession, md5hash, extStream) {
			return nil, errors.New("upload failed")
		}
	}
	image.MsgInfo = uploadResp.Upload.MsgInfo
	image.CompatFace = uploadResp.Upload.CompatQMsg
	return image, nil
}
