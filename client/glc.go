package client

import (
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/packets/tlv"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

func (c *QQClient) QRCodeConfirmed() error {
	app := c.version()
	device := c.Device()
	response, err := c.sendUniPacketAndWait(
		"wtlogin.login",
		c.buildLoginPacket(c.Uin, "wtlogin.login", binary.NewBuilder(nil).
			WriteU16(0x09).
			WriteTLV(
				binary.NewBuilder(nil).WriteBytes(c.t106).Pack(0x106),
				tlv.T144(c.transport.Sig.Tgtgt, app, device),
				tlv.T116(app.SubSigmap),
				tlv.T142(app.PackageName, 0),
				tlv.T145(utils.MustParseHexStr(device.Guid)),
				tlv.T18(0, app.AppClientVersion, int(c.Uin), 0, 5, 0),
				tlv.T141([]byte("Unknown"), nil),
				tlv.T177(app.WTLoginSDK, 0),
				tlv.T191(0),
				tlv.T100(5, app.AppID, app.SubAppID, 8001, app.MainSigmap, 0),
				tlv.T107(1, 0x0d, 0, 1),
				tlv.T318(nil),
				binary.NewBuilder(nil).WriteBytes(c.t16a).Pack(0x16a),
				tlv.T166(5),
				tlv.T521(0x13, "basicim"),
			).ToBytes()))

	if err != nil {
		return err
	}

	return c.decodeLoginResponse(response, &c.transport.Sig)
}

func (c *QQClient) SessionLogin() error {
	// prefer session login
	c.infoln("Session found, try to login with session")
	c.Uin = c.transport.Sig.Uin
	if c.Online.Load() {
		return ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return err
	}
	err = c.Register()
	if err != nil {
		err = fmt.Errorf("failed to register session: %v", err)
		c.errorln(err)
		return err
	}
	return nil
}
