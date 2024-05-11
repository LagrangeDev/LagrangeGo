package event

import (
	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
)

func ParseSelfRenameEvent(event *message.SelfRenameMsg, sig *auth.SigInfo) *Rename {
	sig.Nickname = event.Body.RenameData.NickName
	return &Rename{
		Uin:      event.Body.Uin,
		Nickname: event.Body.RenameData.NickName,
	}
}
