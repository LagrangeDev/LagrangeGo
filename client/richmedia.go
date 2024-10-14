package client

import (
	"encoding/hex"
	"errors"

	highway2 "github.com/LagrangeDev/LagrangeGo/client/internal/highway"
	"github.com/LagrangeDev/LagrangeGo/client/packets/oidb"
	message2 "github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/highway"
	oidb2 "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
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

func (c *QQClient) UploadImage(target message.Source, image *message.ImageElement) (*message.ImageElement, error) {
	switch target.SourceType {
	case message.SourceGroup:
		return c.ImageUploadGroup(uint32(target.PrimaryID), image)
	case message.SourcePrivate:
		return c.ImageUploadPrivate(c.GetUid(uint32(target.PrimaryID)), image)
	}
	return nil, errors.New("unknown target type")
}

func (c *QQClient) UploadRecord(target message.Source, voice *message.VoiceElement) (*message.VoiceElement, error) {
	switch target.SourceType {
	case message.SourceGroup:
		return c.RecordUploadGroup(uint32(target.PrimaryID), voice)
	case message.SourcePrivate:
		return c.RecordUploadPrivate(c.GetUid(uint32(target.PrimaryID)), voice)
	}
	return nil, errors.New("unknown target type")
}
func (c *QQClient) UploadShortVideo(target message.Source, video *message.ShortVideoElement) (*message.ShortVideoElement, error) {
	switch target.SourceType {
	case message.SourceGroup:
		return c.VideoUploadGroup(uint32(target.PrimaryID), video)
	case message.SourcePrivate:
		return c.VideoUploadPrivate(c.GetUid(uint32(target.PrimaryID)), video)
	}
	return nil, errors.New("unknown target type")
}

func (c *QQClient) ImageUploadPrivate(targetUid string, image *message.ImageElement) (*message.ImageElement, error) {
	if image == nil || image.Stream == nil {
		return nil, errors.New("image is nil")
	}
	defer utils.CloseIO(image.Stream)
	image.IsGroup = false
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
			image.Stream, uint64(image.Size),
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

func (c *QQClient) ImageUploadGroup(groupUin uint32, image *message.ImageElement) (*message.ImageElement, error) {
	if image == nil || image.Stream == nil {
		return nil, errors.New("element type is not group image")
	}
	defer utils.CloseIO(image.Stream)
	image.IsGroup = true
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
			image.Stream, uint64(image.Size),
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

func (c *QQClient) RecordUploadPrivate(targetUid string, record *message.VoiceElement) (*message.VoiceElement, error) {
	if record == nil || record.Stream == nil {
		return nil, errors.New("element type is not friend record")
	}
	defer utils.CloseIO(record.Stream)
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
		sha1, err := hex.DecodeString(index.Info.FileSha1)
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
				FileSha1: [][]byte{sha1},
			},
		}
		extStream, err := proto.Marshal(extend)
		if err != nil {
			return nil, err
		}
		md5, err := hex.DecodeString(index.Info.FileHash)
		if err != nil {
			return nil, err
		}
		err = c.highwayUpload(1007, record.Stream, uint64(record.Size), md5, extStream)
		if err != nil {
			return nil, err
		}
	}
	record.MsgInfo = uploadResp.Upload.MsgInfo
	record.Compat = uploadResp.Upload.CompatQMsg
	return record, nil
}

func (c *QQClient) RecordUploadGroup(groupUin uint32, record *message.VoiceElement) (*message.VoiceElement, error) {
	if record == nil || record.Stream == nil {
		return nil, errors.New("element type is not voice record")
	}
	defer utils.CloseIO(record.Stream)
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
		sha1, err := hex.DecodeString(index.Info.FileSha1)
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
				FileSha1: [][]byte{sha1},
			},
		}
		extStream, err := proto.Marshal(extend)
		if err != nil {
			return nil, err
		}
		hash, err := hex.DecodeString(index.Info.FileHash)
		if err != nil {
			return nil, err
		}
		err = c.highwayUpload(1008, record.Stream, uint64(record.Size), hash, extStream)
		if err != nil {
			return nil, err
		}
	}
	record.MsgInfo = uploadResp.Upload.MsgInfo
	record.Compat = uploadResp.Upload.CompatQMsg
	return record, nil
}

func (c *QQClient) VideoUploadPrivate(targetUid string, video *message.ShortVideoElement) (*message.ShortVideoElement, error) {
	if video == nil || video.Stream == nil {
		return nil, errors.New("video is nil")
	}
	if video.Thumb == nil || video.Thumb.Stream == nil {
		return nil, errors.New("video thumb is nil")
	}
	defer utils.CloseIO(video.Stream)
	defer utils.CloseIO(video.Thumb.Stream)
	req, err := oidb.BuildPrivateVideoUploadReq(targetUid, video)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	uploadResp, err := oidb.ParsePrivateVideoUploadResp(resp)
	if err != nil {
		return nil, err
	}
	ukey := uploadResp.Upload.UKey.Unwrap()
	// video
	if ukey != "" {
		index := uploadResp.Upload.MsgInfo.MsgInfoBody[0].Index
		extend := &highway.NTV2RichMediaHighwayExt{
			FileUuid: index.FileUuid,
			UKey:     ukey,
			Network: &highway.NTHighwayNetwork{
				IPv4S: oidbIPv4ToNTHighwayIPv4(uploadResp.Upload.IPv4S),
			},
			MsgInfoBody: uploadResp.Upload.MsgInfo.MsgInfoBody,
			BlockSize:   uint32(highway2.BlockSize),
			Hash: &highway.NTHighwayHash{
				FileSha1: crypto.ComputeBlockSha1(video.Stream, highway2.BlockSize),
			},
		}
		extStream, err := proto.Marshal(extend)
		if err != nil {
			return nil, err
		}
		md5, err := hex.DecodeString(index.Info.FileHash)
		if err != nil {
			return nil, err
		}
		err = c.highwayUpload(1001, video.Stream, uint64(video.Size), md5, extStream)
		if err != nil {
			return nil, err

		}
	}
	// thumb
	subFile := uploadResp.Upload.SubFileInfos[0]
	if subFile.UKey != "" {
		index := uploadResp.Upload.MsgInfo.MsgInfoBody[1].Index
		sha1, err := hex.DecodeString(index.Info.FileSha1)
		if err != nil {
			return nil, err
		}
		extend := &highway.NTV2RichMediaHighwayExt{
			FileUuid: index.FileUuid,
			UKey:     subFile.UKey,
			Network: &highway.NTHighwayNetwork{
				IPv4S: oidbIPv4ToNTHighwayIPv4(subFile.IPv4S),
			},
			MsgInfoBody: uploadResp.Upload.MsgInfo.MsgInfoBody,
			BlockSize:   uint32(highway2.BlockSize),
			Hash: &highway.NTHighwayHash{
				FileSha1: [][]byte{sha1},
			},
		}
		extStream, err := proto.Marshal(extend)
		if err != nil {
			return nil, err
		}
		md5, err := hex.DecodeString(index.Info.FileHash)
		if err != nil {
			return nil, err
		}
		err = c.highwayUpload(1002, video.Thumb.Stream, uint64(video.Thumb.Size), md5, extStream)
		if err != nil {
			return nil, err
		}
	}
	video.MsgInfo = uploadResp.Upload.MsgInfo
	err = proto.Unmarshal(uploadResp.Upload.CompatQMsg, video.Compat)
	if err != nil {
		return nil, err
	}
	return video, nil
}

func (c *QQClient) VideoUploadGroup(groupUin uint32, video *message.ShortVideoElement) (*message.ShortVideoElement, error) {
	if video == nil || video.Stream == nil {
		return nil, errors.New("video is nil")
	}
	if video.Thumb == nil || video.Thumb.Stream == nil {
		return nil, errors.New("video thumb is nil")
	}
	defer utils.CloseIO(video.Stream)
	defer utils.CloseIO(video.Thumb.Stream)
	req, err := oidb.BuildGroupVideoUploadReq(groupUin, video)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	uploadResp, err := oidb.ParseGroupVideoUploadResp(resp)
	if err != nil {
		return nil, err
	}
	ukey := uploadResp.Upload.UKey.Unwrap()
	// video
	if ukey != "" {
		index := uploadResp.Upload.MsgInfo.MsgInfoBody[0].Index
		extend := &highway.NTV2RichMediaHighwayExt{
			FileUuid: index.FileUuid,
			UKey:     ukey,
			Network: &highway.NTHighwayNetwork{
				IPv4S: oidbIPv4ToNTHighwayIPv4(uploadResp.Upload.IPv4S),
			},
			MsgInfoBody: uploadResp.Upload.MsgInfo.MsgInfoBody,
			BlockSize:   uint32(highway2.BlockSize),
			Hash: &highway.NTHighwayHash{
				FileSha1: crypto.ComputeBlockSha1(video.Stream, highway2.BlockSize),
			},
		}
		extStream, err := proto.Marshal(extend)
		if err != nil {
			return nil, err
		}
		md5, err := hex.DecodeString(index.Info.FileHash)
		if err != nil {
			return nil, err
		}
		err = c.highwayUpload(1005, video.Stream, uint64(video.Size), md5, extStream)
		if err != nil {
			return nil, err

		}
	}
	// thumb
	subFile := uploadResp.Upload.SubFileInfos[0]
	if subFile.UKey != "" {
		index := uploadResp.Upload.MsgInfo.MsgInfoBody[1].Index
		sha1, err := hex.DecodeString(index.Info.FileSha1)
		if err != nil {
			return nil, err
		}
		extend := &highway.NTV2RichMediaHighwayExt{
			FileUuid: index.FileUuid,
			UKey:     subFile.UKey,
			Network: &highway.NTHighwayNetwork{
				IPv4S: oidbIPv4ToNTHighwayIPv4(subFile.IPv4S),
			},
			MsgInfoBody: uploadResp.Upload.MsgInfo.MsgInfoBody,
			BlockSize:   uint32(highway2.BlockSize),
			Hash: &highway.NTHighwayHash{
				FileSha1: [][]byte{sha1},
			},
		}
		extStream, err := proto.Marshal(extend)
		if err != nil {
			return nil, err
		}
		md5, err := hex.DecodeString(index.Info.FileHash)
		if err != nil {
			return nil, err
		}
		err = c.highwayUpload(1006, video.Thumb.Stream, uint64(video.Thumb.Size), md5, extStream)
		if err != nil {
			return nil, err
		}
	}
	video.MsgInfo = uploadResp.Upload.MsgInfo
	err = proto.Unmarshal(uploadResp.Upload.CompatQMsg, video.Compat)
	if err != nil {
		return nil, err
	}
	return video, nil
}

func (c *QQClient) FileUploadPrivate(targetUid string, file *message.FileElement) (*message.FileElement, error) {
	if file == nil || file.FileStream == nil {
		return nil, errors.New("element type is not file")
	}
	req, err := oidb.BuildPrivateFileUploadReq(c.GetUid(c.Uin), targetUid, file)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	uploadResp, err := oidb.ParsePrivateFileUploadResp(resp)
	if err != nil {
		return nil, err
	}
	if !uploadResp.Upload.BoolFileExist {
		ext := &highway.FileUploadExt{
			Unknown1: 100,
			Unknown2: 1,
			Entry: &highway.FileUploadEntry{
				BusiBuff: &highway.ExcitingBusiInfo{
					SenderUin: uint64(c.Uin),
				},
				FileEntry: &highway.ExcitingFileEntry{
					FileSize:  file.FileSize,
					Md5:       file.FileMd5,
					CheckKey:  file.FileSha1,
					Md5S2:     file.FileMd5,
					FileId:    uploadResp.Upload.Uuid,
					UploadKey: uploadResp.Upload.MediaPlatformUploadKey,
				},
				ClientInfo: &highway.ExcitingClientInfo{
					ClientType:   3,
					AppId:        "100",
					TerminalType: 3,
					ClientVer:    "1.1.1",
					Unknown:      4,
				},
				FileNameInfo: &highway.ExcitingFileNameInfo{
					FileName: file.FileName,
				},
				Host: &highway.ExcitingHostConfig{
					Hosts: []*highway.ExcitingHostInfo{
						{
							Url: &highway.ExcitingUrlInfo{
								Unknown: 1,
								Host:    uploadResp.Upload.UploadIp,
							},
							Port: uploadResp.Upload.UploadPort,
						},
					},
				},
			},
			Unknown200: 1,
			Unknown3:   0,
		}
		extStream, err := proto.Marshal(ext)
		if err != nil {
			return nil, err
		}
		if err = c.highwayUpload(95, file.FileStream, file.FileSize, file.FileMd5, extStream); err != nil {
			return nil, err
		}
	}
	file.FileHash = uploadResp.Upload.FileAddon
	file.FileUUID = uploadResp.Upload.Uuid
	return file, nil
}

func (c *QQClient) FileUploadGroup(groupUin uint32, file *message.FileElement, targetDirectory string) (*message.FileElement, error) {
	if file == nil || file.FileStream == nil {
		return nil, errors.New("element type is not group file")
	}
	req, err := oidb.BuildGroupFileUploadReq(groupUin, file, targetDirectory)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	uploadResp, err := oidb.ParseGroupFileUploadResp(resp)
	if err != nil {
		return nil, err
	}
	if !uploadResp.Upload.BoolFileExist {
		ext := &highway.FileUploadExt{
			Unknown1: 100,
			Unknown2: 1,
			Entry: &highway.FileUploadEntry{
				BusiBuff: &highway.ExcitingBusiInfo{
					SenderUin:   uint64(c.Uin),
					ReceiverUin: uint64(groupUin),
					GroupCode:   uint64(groupUin),
				},
				FileEntry: &highway.ExcitingFileEntry{
					FileSize:  file.FileSize,
					Md5:       file.FileMd5,
					CheckKey:  uploadResp.Upload.CheckKey,
					Md5S2:     file.FileMd5,
					FileId:    uploadResp.Upload.FileId,
					UploadKey: uploadResp.Upload.FileKey,
				},
				ClientInfo: &highway.ExcitingClientInfo{
					ClientType:   3,
					AppId:        "100",
					TerminalType: 3,
					ClientVer:    "1.1.1",
					Unknown:      4,
				},
				FileNameInfo: &highway.ExcitingFileNameInfo{
					FileName: file.FileName,
				},
				Host: &highway.ExcitingHostConfig{
					Hosts: []*highway.ExcitingHostInfo{
						{
							Url: &highway.ExcitingUrlInfo{
								Unknown: 1,
								Host:    uploadResp.Upload.UploadIp,
							},
							Port: uploadResp.Upload.UploadPort,
						},
					},
				},
			},
			Unknown200: 0,
		}
		extStream, err := proto.Marshal(ext)
		if err != nil {
			return nil, err
		}
		if err = c.highwayUpload(71, file.FileStream, file.FileSize, file.FileMd5, extStream); err != nil {
			return nil, err
		}
	}
	req, err = oidb.BuildGroupSendFileReq(groupUin, uploadResp.Upload.FileId)
	if err != nil {
		return nil, err
	}
	resp, err = c.sendOidbPacketAndWait(req)
	if err != nil {
		return nil, err
	}
	if _, err = oidb.ParseGroupSendFileResp(resp); err != nil {
		return nil, err
	}
	return file, nil
}
