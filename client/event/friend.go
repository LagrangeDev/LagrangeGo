package event

import (
	"fmt"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
)

type (
	NewFriendRequest struct {
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

	// FriendPokeEvent 好友戳一戳事件 from miraigo
	FriendPokeEvent struct {
		Sender   uint32
		Receiver uint32
		Suffix   string
		Action   string
	}
)

func ParseFriendRequestNotice(event *message.FriendRequest, msg *message.PushMsg) *NewFriendRequest {
	info := event.Info
	return &NewFriendRequest{
		SourceUin: msg.Message.ResponseHead.FromUin,
		SourceUid: info.SourceUid,
		Msg:       info.Message,
		Source:    info.Source,
	}
}

func (fe *FriendRecall) ResolveUin(f func(uid string) uint32) {
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

func (fe *Rename) ResolveUin(f func(uid string) uint32) {
	fe.Uin = f(fe.Uid)
}

func ParseFriendRenameEvent(event *message.FriendRenameMsg) *Rename {
	return &Rename{
		SubType:  1,
		Uid:      event.Body.Data.Uid,
		Nickname: event.Body.Data.RenameData.NickName,
	}
}

func ParsePokeEvent(event *message.GeneralGrayTipInfo) *FriendPokeEvent {
	e := FriendPokeEvent{}
	e.Action = "戳了戳"
	for _, data := range event.MsgTemplParam {
		switch data.Key {
		case "uin_str1":
			sender, _ := strconv.Atoi(data.Value)
			e.Sender = uint32(sender)
		case "uin_str2":
			receiver, _ := strconv.Atoi(data.Value)
			e.Receiver = uint32(receiver)
		case "suffix_str":
			e.Suffix = data.Value
		case "alt_str1":
			e.Action = data.Value
		}
	}
	return &e
}

func (g *FriendPokeEvent) From() uint32 {
	return g.Sender
}

func (g *FriendPokeEvent) Content() string {
	if g.Suffix != "" {
		return fmt.Sprintf("%d%s%d的%s", g.Sender, g.Action, g.Receiver, g.Suffix)
	} else {
		return fmt.Sprintf("%d%s%d", g.Sender, g.Action, g.Receiver)
	}
}
