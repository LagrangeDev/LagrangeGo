package client

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/packet.go

import (
	"github.com/LagrangeDev/LagrangeGo/client/internal/network"
)

func (c *QQClient) uniPacket(command string, body []byte) (uint32, []byte) {
	seq := c.getAndIncreaseSequence()
	var sign map[string]string
	if c.signProvider != nil {
		sign = c.signProvider(command, seq, body)
	}
	req := network.Request{
		SequenceID:  seq,
		Uin:         int64(c.Uin),
		Sign:        sign,
		CommandName: command,
		Body:        body,
	}
	return seq, c.transport.PackPacket(&req)
}
