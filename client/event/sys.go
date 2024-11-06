package event

import "github.com/LagrangeDev/LagrangeGo/client/packets/pb/system"

type Kicked struct {
	Tips   string
	Tittle string
}

func ParseKickedEvent(e *system.ServiceKickNTResponse) *Kicked {
	return &Kicked{
		Tips:   e.Tips,
		Tittle: e.Title,
	}
}
