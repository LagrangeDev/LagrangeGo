package client

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/wtlogin/loginState"

	"github.com/LagrangeDev/LagrangeGo/client/wtlogin/qrcodeState"

	binary2 "github.com/LagrangeDev/LagrangeGo/utils/binary"

	"github.com/LagrangeDev/LagrangeGo/client/wtlogin/tlv"
	"github.com/LagrangeDev/LagrangeGo/utils"

	"github.com/LagrangeDev/LagrangeGo/client/wtlogin"

	"github.com/LagrangeDev/LagrangeGo/info"
)

const Server = "msfwifi.3g.qq.com:8080"

var (
	resultStore   = NewResultStore()
	networkLogger = utils.GetLogger("network")
)

// NewQQclient 创建一个新的QQClient
func NewQQclient(uin uint32, signUrl string, appInfo *info.AppInfo, deviceInfo *info.DeviceInfo, sig *info.SigInfo) *QQClient {
	return &QQClient{
		uin:          uin,
		appInfo:      appInfo,
		deviceInfo:   deviceInfo,
		sig:          sig,
		signProvider: utils.SignProvider(signUrl),
		// 128应该够用了吧
		pushStore: make(chan *wtlogin.SSOPacket, 128),
		online:    false,
		tcp:       NewTCPClient(Server, 5),
	}
}

type QQClient struct {
	lock         sync.RWMutex
	uin          uint32
	appInfo      *info.AppInfo
	deviceInfo   *info.DeviceInfo
	sig          *info.SigInfo
	signProvider func(string, int, []byte) map[string]string

	pushStore chan *wtlogin.SSOPacket

	online bool

	t106 []byte
	t16a []byte

	tcp *TCPClient
}

func (c *QQClient) Login(password, qrcodePath string) (bool, error) {
	if len(c.sig.D2) != 0 { // prefer session login
		loginLogger.Infoln("Session found, try to login with session")
		return c.Register()
	}

	if len(c.sig.TempPwd) != 0 {
		c.KeyExchange()

		ret, err := c.TokenLogin(c.sig.TempPwd)
		if err != nil {
			return false, errors.New(fmt.Sprintf("EasyLogin fail %s", err))
		}
		if ret.Successful() {
			return c.Register()
		}
	}

	if password != "" {
		c.KeyExchange()

		for {
			ret, err := c.PasswordLogin(password)
			if err != nil {
				return false, err
			}
			if ret.Successful() {
				return c.Register()
			} else if ret == loginState.CaptchaVerify {
				loginLogger.Warningln("captcha verification required")
				c.sig.CaptchaInfo[0] = utils.ReadLine("ticket?->")
				c.sig.CaptchaInfo[1] = utils.ReadLine("rand_str?->")
			} else {
				loginLogger.Errorf("Unhandled exception raised: %s", ret.Name())
			}
		}
	} else {
		png, _, err := c.FecthQrcode()
		if err != nil {
			return false, err
		}
		_ = os.WriteFile(qrcodePath, png, 0666)
		loginLogger.Infof("qrcode saved to %s", qrcodePath)
		if c.QrcodeLogin(3) {
			return c.Register()
		}
	}
	return false, nil
}

func (c *QQClient) SendUniPacket(cmd string, buf []byte) error {
	seq := c.getSeq()
	var sign map[string]string
	if c.signProvider != nil {
		sign = c.signProvider(cmd, seq, buf)
	}
	packet := wtlogin.BuildUniPacket(int(c.uin), seq, cmd, sign, c.appInfo, c.deviceInfo, c.sig, buf)
	return c.Send(packet)
}

func (c *QQClient) SendUniPacketAndAwait(cmd string, buf []byte) (*wtlogin.SSOPacket, error) {
	seq := c.getSeq()
	var sign map[string]string
	if c.signProvider != nil {
		sign = c.signProvider(cmd, seq, buf)
	}
	packet := wtlogin.BuildUniPacket(int(c.uin), seq, cmd, sign, c.appInfo, c.deviceInfo, c.sig, buf)
	return c.SendAndWait(packet, seq, 5)
}

func (c *QQClient) FecthQrcode() ([]byte, string, error) {
	body := utils.NewPacketBuilder(nil).
		WriteU16(0).
		WriteU64(0).
		WriteU8(0).
		WriteTlv([][]byte{
			tlv.T16(c.appInfo.AppID, c.appInfo.SubAppID,
				utils.GetBytesFromHex(c.deviceInfo.Guid), c.appInfo.PTVersion, c.appInfo.PackageName),
			tlv.T1b(),
			tlv.T1d(c.appInfo.MiscBitmap),
			tlv.T33(utils.GetBytesFromHex(c.deviceInfo.Guid)),
			tlv.T35(c.appInfo.PTOSVersion),
			tlv.T66(c.appInfo.PTOSVersion),
			tlv.Td1(c.appInfo.OS, c.deviceInfo.DeviceName),
		}).WriteU8(3).Pack(-1)

	packet := wtlogin.BuildCode2dPacket(c.uin, 0x31, c.appInfo, body)
	response, err := c.SendUniPacketAndAwait("wtlogin.trans_emp", packet)
	if err != nil {
		return nil, "", err
	}

	decrypted := binary2.NewReader(response.Data)
	decrypted.ReadBytes(54)
	retCode := decrypted.ReadU8()
	qrsig := decrypted.ReadBytesWithLength("u16", false)
	tlvs := decrypted.ReadTlv()

	if retCode == 0 && tlvs[0x17] != nil {
		c.sig.Qrsig = qrsig
		urlreader := binary2.NewReader(tlvs[209])
		// 这样是不对的，调试后发现应该丢一个字节，然后读下一个字节才是数据的大小
		// string(binary2.NewReader(tlvs[209]).ReadBytesWithLength("u16", true))
		urlreader.ReadU8()
		return tlvs[0x17], string(urlreader.ReadBytesWithLength("u8", false)), nil
	}

	return nil, "", errors.New(fmt.Sprintf("err qr retcode %d", retCode))
}

func (c *QQClient) GetQrcodeResult() (qrcodeState.State, error) {
	loginLogger.Tracef("get qrcode result")
	if c.sig.Qrsig == nil {
		return -1, errors.New("no qrsig found, execute fetch_qrcode first")
	}

	body := utils.NewPacketBuilder(nil).
		WriteBytes(c.sig.Qrsig, "u16", false).
		WriteU64(0).
		WriteU32(0).
		WriteU8(0).
		WriteU8(0x83).Pack(-1)

	response, err := c.SendUniPacketAndAwait("wtlogin.trans_emp",
		wtlogin.BuildCode2dPacket(0, 0x12, c.appInfo, body))
	if err != nil {
		return -1, err
	}

	reader := binary2.NewReader(response.Data)
	//length := reader.ReadU32()
	reader.ReadBytes(8) // 4 + 4
	reader.ReadU16()    // cmd, 0x12
	reader.ReadBytes(40)
	_ = reader.ReadU32() // app id
	retCode := qrcodeState.State(reader.ReadU8())

	if retCode == 0 {
		reader.ReadBytes(4)
		c.uin = reader.ReadU32()
		reader.ReadBytes(4)
		t := reader.ReadTlv()
		c.t106 = t[0x18]
		c.t16a = t[0x19]
		c.sig.Tgtgt = t[0x1e]
	}

	return retCode, nil
}

func (c *QQClient) KeyExchange() {
	packet, err := c.SendUniPacketAndAwait(
		"trpc.login.ecdh.EcdhService.SsoKeyExchange",
		wtlogin.BuildKexExchangeRequest(c.uin, c.deviceInfo.Guid))
	if err != nil {
		networkLogger.Errorln(err)
		return
	}
	wtlogin.ParseKeyExchangeResponse(packet.Data, c.sig)
}

func (c *QQClient) PasswordLogin(password string) (loginState.State, error) {
	md5Password := utils.Md5Digest([]byte(password))

	cr := tlv.T106(
		c.appInfo.AppID,
		c.appInfo.AppClientVersion,
		int(c.uin),
		c.deviceInfo.Guid,
		md5Password,
		c.sig.Tgtgt,
		make([]byte, 4),
		true)[4:]

	packet, err := c.SendUniPacketAndAwait(
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLogin",
		buildNtloginRequest(c.uin, c.appInfo, c.deviceInfo, c.sig, cr))
	if err != nil {
		return -999, err
	}
	return ParseNtloginResponse(packet.Data, c.sig)
}

func (c *QQClient) TokenLogin(token []byte) (loginState.State, error) {
	packet, err := c.SendUniPacketAndAwait(
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLogin",
		buildNtloginRequest(c.uin, c.appInfo, c.deviceInfo, c.sig, token))
	if err != nil {
		return -999, err
	}
	return ParseNtloginResponse(packet.Data, c.sig)
}

func (c *QQClient) QrcodeLogin(refreshInterval int) bool {
	if c.sig.Qrsig == nil {
		loginLogger.Fatal("No QrSig found, fetch qrcode first")
	}

	for !c.tcp.IsClosed() {
		time.Sleep(time.Duration(refreshInterval) * time.Second)
		retCode, err := c.GetQrcodeResult()
		if err != nil {
			loginLogger.Error(err)
			return false
		}
		if !retCode.Waitable() {
			if !retCode.Success() {
				loginLogger.Fatal(retCode.Name())
			} else {
				break
			}
		}
	}

	app := c.appInfo
	device := c.deviceInfo
	body := utils.NewPacketBuilder(nil).
		WriteU16(0x09).
		WriteTlv([][]byte{
			utils.NewPacketBuilder(nil).WriteBytes(c.t106, "", true).Pack(0x106),
			tlv.T144(c.sig.Tgtgt, app, device),
			tlv.T116(app.SubSigmap),
			tlv.T142(app.PackageName, 0),
			tlv.T145(utils.GetBytesFromHex(device.Guid)),
			tlv.T18(0, app.AppClientVersion, int(c.uin), 0, 5, 0),
			tlv.T141([]byte("Unknown"), make([]byte, 0)),
			tlv.T177(app.WTLoginSDK, 0),
			tlv.T191(0),
			tlv.T100(5, app.AppID, app.SubAppID, 8001, app.MainSigmap, 0),
			tlv.T107(1, 0x0d, 0, 1),
			tlv.T318(make([]byte, 0)),
			utils.NewPacketBuilder(nil).WriteBytes(c.t16a, "", true).Pack(0x16a),
			tlv.T166(5),
			tlv.T521(0x13, "basicim"),
		}).Pack(-1)

	response, err := c.SendUniPacketAndAwait(
		"wtlogin.login",
		wtlogin.BuildLoginPacket(c.uin, "wtlogin.login", app, body))

	if err != nil {
		loginLogger.Fatal(err)
		return false
	}

	return wtlogin.DecodeLoginResponse(response.Data, c.sig)
}

func (c *QQClient) Register() (bool, error) {
	response, err := c.SendUniPacketAndAwait(
		"trpc.qq_new_tech.status_svc.StatusService.Register",
		wtlogin.BuildRegisterRequest(c.appInfo, c.deviceInfo))

	if err != nil {
		networkLogger.Error(err)
		return false, err
	}

	if wtlogin.ParseRegisterResponse(response.Data) {
		c.setOnline()
		networkLogger.Info("Register successful")
		return true, nil
	}
	networkLogger.Error("Register failure")
	return false, errors.New("register failure")
}

func (c *QQClient) SSOHeartbeat(calcLatency bool) int64 {
	startTime := time.Now().Unix()
	_, err := c.SendUniPacketAndAwait(
		"trpc.qq_new_tech.status_svc.StatusService.SsoHeartBeat",
		wtlogin.BuildSSOHeartbeatRequest())
	if err != nil {
		return 0
	}
	if calcLatency {
		return time.Now().Unix() - startTime
	}
	return 0
}

func (c *QQClient) Send(data []byte) error {
	return c.tcp.Write(data)
}

func (c *QQClient) SendAndWait(data []byte, seq, timeout int) (*wtlogin.SSOPacket, error) {
	resultStore.AddSeq(seq)
	err := c.tcp.Write(data)
	if err != nil {
		// 出错了要删掉
		resultStore.DeleteSeq(seq)
	}
	return resultStore.Fecth(seq, timeout)
}

func (c *QQClient) OnMessage(msgLen int) {
	raw, err := c.tcp.ReadBytes(msgLen)
	if err != nil {
		networkLogger.Errorf("read message error: %s", err)
		return
	}
	ssoHeader, err := wtlogin.ParseSSOHeader(raw, c.sig.D2Key)
	if err != nil {
		networkLogger.Errorf("ParseSSOHeader error %s", err)
		return
	}
	packet, err := wtlogin.ParseSSOFrame(ssoHeader.Dec, ssoHeader.Flag == 2)
	if err != nil {
		networkLogger.Errorf("ParseSSOFrame error %s", err)
		return
	}

	if packet.Seq > 0 { // uni rsp
		networkLogger.Debugf("%d(%d) -> %s, extra: %s", packet.Seq, packet.RetCode, packet.Cmd, packet.Extra)
		if packet.RetCode != 0 && resultStore.ContainSeq(packet.Seq) {
			networkLogger.Errorf("error ssopacket retcode: %d, extra: %s", packet.RetCode, packet.Extra)
			return
		} else if packet.RetCode != 0 {
			networkLogger.Errorf("Unexpected error on sso layer: %d: %s", packet.RetCode, packet.Extra)
			return
		}
		if !resultStore.ContainSeq(packet.Seq) {
			networkLogger.Warningf("Unknown packet: %s(%d), ignore", packet.Cmd, packet.Seq)
		} else {
			resultStore.AddResult(packet.Seq, packet)
		}
	} else { // server pushed
		networkLogger.Debugf("%d(%d) <- %s,extra %s", packet.Seq, packet.RetCode, packet.Cmd, packet.Extra)
		// 扔到chan里面算了。还是丢掉吧
		// c.pushStore <- packet
	}
}

func (c *QQClient) Loop() {
	err := c.Connect()
	if err != nil {

	}
	go c.ReadLoop()
}

func (c *QQClient) ReadLoop() {
	for !c.tcp.IsClosed() {
		lengthData, err := c.tcp.ReadBytes(4)
		if err != nil {
			networkLogger.Errorf("tcp read length error: %s", err)
			return
		}
		length := int(binary2.NewReader(lengthData).ReadU32() - 4)
		if length > 0 {
			c.OnMessage(length)
		} else {
			c.tcp.Close()
		}
	}
	c.OnDissconneted()
}

func (c *QQClient) Connect() error {
	err := c.tcp.Connect()
	if err != nil {
		return err
	}
	c.OnConnected()
	return nil
}

func (c *QQClient) DisConnect() {
	c.tcp.Close()
	c.OnDissconneted()
}

// setOnline 设置qq已经上线
func (c *QQClient) setOnline() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.online = true
}

func (c *QQClient) OnConnected() {

}

func (c *QQClient) OnDissconneted() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.online = false
}

func (c *QQClient) getSeq() int {
	defer func() {
		if c.sig.Sequence >= 0x8000 {
			c.sig.Sequence = 0
		}
		c.sig.Sequence++
	}()
	return c.sig.Sequence
}
