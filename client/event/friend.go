package event

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
)

type (
	FriendRequest struct {
		SourceUin uint32
		SourceUid string
		Msg       string
		Source    string
	}

	FriendRecall struct {
		FromUid  string
		Sequence uint64
		Time     uint32
		Random   uint32
	}

	Rename struct {
		SubType  uint32 // self 0 friend 1
		Uin      uint32
		Nickname string
	}
)

func ParseFriendRequestNotice(event *message.FriendRequest, msg *message.PushMsg) *FriendRequest {
	info := event.Info
	return &FriendRequest{
		SourceUin: msg.Message.ResponseHead.FromUin,
		SourceUid: info.SourceUid,
		Msg:       info.Message,
		Source:    info.Source,
	}
}

func ParseFriendRecallEvent(event *message.FriendRecall) *FriendRecall {
	info := event.Info
	return &FriendRecall{
		FromUid:  info.FromUid,
		Sequence: uint64(info.Sequence),
		Time:     info.Time,
		Random:   info.Random,
	}
}

func ParseFriendRenameEvent(event *message.FriendRenameMsg, uin uint32) *Rename {
	return &Rename{
		SubType:  1,
		Uin:      uin,
		Nickname: event.Body.Data.RenameData.NickName,
	}
}
