package oidb

import (
	"encoding/hex"
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/message"
)

func BuildAiCharacterListService(groupUin uint32, chatType entity.ChatType) (*Packet, error) {
	return BuildOidbPacket(0x929D, 0, &oidb.OidbSvcTrpcTcp0X929D_0_Req{GroupUin: groupUin, ChatType: uint32(chatType)}, false, false)
}

func ParseAiCharacterListService(data []byte) (*entity.AiCharacterList, error) {
	var rsp oidb.OidbSvcTrpcTcp0X929D_0_Rsp
	_, err := ParseOidbPacket(data, &rsp)
	if err != nil {
		return nil, err
	}
	var ret entity.AiCharacterList
	for _, property := range rsp.Property {
		info := entity.AiCharacterInfo{Type: property.Type}
		for _, v := range property.Value {
			info.Characters = append(info.Characters, entity.AiCharacter{
				Name:     v.CharacterName,
				VoiceID:  v.CharacterId,
				VoiceURL: v.CharacterVoiceUrl,
			})
		}
		ret.List = append(ret.List, info)
	}
	if len(ret.List) == 0 {
		return nil, errors.New("ai character list nil")
	}
	return &ret, nil
}

func BuildGroupAiRecordService(groupUin uint32, voiceID, text string, chatType entity.ChatType, chatID uint32) (*Packet, error) {
	return BuildOidbPacket(0x929B, 0, &oidb.OidbSvcTrpcTcp0X929B_0_Req{
		GroupUin:      groupUin,
		VoiceId:       voiceID,
		Text:          text,
		ChatType:      uint32(chatType),
		ClientMsgInfo: &oidb.OidbSvcTrpcTcp0X929B_0_Req_ClientMsgInfo{MsgRandom: chatID},
	}, false, false)
}

func ParseGroupAiRecordService(data []byte) (*message.VoiceElement, error) {
	var rsp oidb.OidbSvcTrpcTcp0X929B_0_Rsp
	_, err := ParseOidbPacket(data, &rsp)
	if err != nil {
		return nil, err
	}
	if rsp.MsgInfo == nil || len(rsp.MsgInfo.MsgInfoBody) == 0 {
		return nil, errors.New("rsp msg info nil")
	}
	index := rsp.MsgInfo.MsgInfoBody[0].Index
	elem := &message.VoiceElement{
		UUID:     index.FileUuid,
		Name:     index.Info.FileName,
		Size:     index.Info.FileSize,
		Duration: index.Info.Time,
		MsgInfo:  rsp.MsgInfo,
		// Compat ??
	}
	elem.Md5, _ = hex.DecodeString(index.Info.FileHash)
	elem.Sha1, _ = hex.DecodeString(index.Info.FileSha1)
	return elem, nil
}
