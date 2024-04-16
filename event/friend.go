package event

import (
	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
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
)

func ParseFriendRequestNotice(msg *message.PushMsg, event *message.FriendRequest) *FriendRequest {
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
