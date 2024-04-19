package event

import (
	"github.com/LagrangeDev/LagrangeGo/info"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
)

func ParseSelfRenameEvent(event *message.SelfRenameMsg, sig *info.SigInfo) *Rename {
	sig.Nickname = event.Body.RenameData.NickName
	return &Rename{
		Uin:      event.Body.Uin,
		Nickname: event.Body.RenameData.NickName,
	}
}
