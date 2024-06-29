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
		FromUin  uint32
		FromUid  string
		Sequence uint64
		Time     uint32
		Random   uint32
	}

	Rename struct {
		SubType  uint32 // self 0 friend 1
		Uin      uint32
		Uid      string
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

func (fe *FriendRecall) Preprocess(f func(uid string) uint32) {
	fe.FromUin = f(fe.FromUid)
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

func (fe *Rename) Preprocess(f func(uid string) uint32) {
	fe.Uin = f(fe.Uid)
}

func ParseFriendRenameEvent(event *message.FriendRenameMsg) *Rename {
	return &Rename{
		SubType:  1,
		Uid:      event.Body.Data.Uid,
		Nickname: event.Body.Data.RenameData.NickName,
	}
}
