package event

import "github.com/LagrangeDev/LagrangeGo/packets/pb/message"

func ParseSelfRenameEvent(event *message.SelfRenameMsg) *Rename {
	return &Rename{
		Uin:      event.Body.Uin,
		Nickname: event.Body.RenameData.NickName,
	}
}
