package client

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/packet.go

import "github.com/LagrangeDev/LagrangeGo/client/internal/network"

func (c *QQClient) uniPacket(command string, body []byte) (uint32, []byte, error) {
	seq := c.getAndIncreaseSequence()
	sign, err := c.signProvider.Sign(command, seq, body)
	if err != nil {
		return 0, nil, err
	}
	req := network.Request{
		SequenceID:  seq,
		Uin:         int64(c.Uin),
		Sign:        sign,
		CommandName: command,
		Body:        body,
	}
	return seq, c.transport.PackPacket(&req), nil
}
