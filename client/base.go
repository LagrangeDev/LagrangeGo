package client

// 部分借鉴 https://github.com/Mrs4s/MiraiGo/blob/master/client/client.go

import (
	"sync/atomic"
	"time"

	"github.com/LagrangeDev/LagrangeGo/cache"

	"github.com/LagrangeDev/LagrangeGo/event"

	"github.com/LagrangeDev/LagrangeGo/info"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/packets/oidb"
	"github.com/LagrangeDev/LagrangeGo/packets/wtlogin"
	"github.com/LagrangeDev/LagrangeGo/utils"
	binary2 "github.com/LagrangeDev/LagrangeGo/utils/binary"
)

const msfwifiServer = "msfwifi.3g.qq.com:8080"

// NewQQclient 创建一个新的QQClient
func NewQQclient(uin uint32, signUrl string, appInfo *info.AppInfo, deviceInfo *info.DeviceInfo, sig *info.SigInfo) *QQClient {
	client := &QQClient{
		Uin:          uin,
		appInfo:      appInfo,
		deviceInfo:   deviceInfo,
		sig:          sig,
		signProvider: utils.SignProvider(signUrl),
		// 128应该够用了吧
		pushStore: make(chan *wtlogin.SSOPacket, 128),
		stopChan:  make(chan struct{}),
		tcp:       &TCPClient{},
		cache:     &cache.Cache{},
	}
	client.Online.Store(false)
	return client
}

type QQClient struct {
	Uin          uint32
	appInfo      *info.AppInfo
	deviceInfo   *info.DeviceInfo
	sig          *info.SigInfo
	signProvider func(string, uint32, []byte) map[string]string

	pushStore chan *wtlogin.SSOPacket

	Online   atomic.Bool
	stopChan chan struct{}

	t106 []byte
	t16a []byte

	tcp *TCPClient

	cache *cache.Cache

	highwayUri      map[uint32][]string
	highwaySequence atomic.Uint32
	sigSession      []byte

	GroupMessageEvent           EventHandle[*message.GroupMessage]
	PrivateMessageEvent         EventHandle[*message.PrivateMessage]
	TempMessageEvent            EventHandle[*message.TempMessage]
	GroupInvitedEvent           EventHandle[*event.GroupInvite]            // 邀请入群
	GroupMemberJoinRequestEvent EventHandle[*event.GroupMemberJoinRequest] // 加群申请
	GroupMemberJoinEvent        EventHandle[*event.GroupMemberIncrease]    // 成员入群
	GroupMemberLeaveEvent       EventHandle[*event.GroupMemberDecrease]    // 成员退群
	GroupMuteEvent              EventHandle[*event.GroupMute]
	GroupRecallEvent            EventHandle[*event.GroupRecall]
	FriendRequestEvent          EventHandle[*event.FriendRequest] // 好友申请
	FriendRecallEvent           EventHandle[*event.FriendRecall]
	RenameEvent                 EventHandle[*event.Rename]
}

func (c *QQClient) NickName() string {
	return c.sig.Nickname
}

func (c *QQClient) SendOidbPacket(pkt *oidb.OidbPacket) error {
	return c.SendUniPacket(pkt.Cmd, pkt.Data)
}

func (c *QQClient) SendOidbPacketAndWait(pkt *oidb.OidbPacket) (*wtlogin.SSOPacket, error) {
	return c.SendUniPacketAndAwait(pkt.Cmd, pkt.Data)
}

func (c *QQClient) SendUniPacket(cmd string, buf []byte) error {
	seq := c.getAndIncreaseSequence()
	var sign map[string]string
	if c.signProvider != nil {
		sign = c.signProvider(cmd, seq, buf)
	}
	packet := wtlogin.BuildUniPacket(int(c.Uin), seq, cmd, sign, c.appInfo, c.deviceInfo, c.sig, buf)
	return c.Send(packet)
}

func (c *QQClient) SendUniPacketAndAwait(cmd string, buf []byte) (*wtlogin.SSOPacket, error) {
	seq := c.getAndIncreaseSequence()
	var sign map[string]string
	if c.signProvider != nil {
		sign = c.signProvider(cmd, seq, buf)
	}
	packet := wtlogin.BuildUniPacket(int(c.Uin), seq, cmd, sign, c.appInfo, c.deviceInfo, c.sig, buf)
	return c.SendAndWait(packet, int(seq), 5)
}

func (c *QQClient) Send(data []byte) error {
	return c.tcp.Write(data)
}

func (c *QQClient) SendAndWait(data []byte, seq int, timeout int) (*wtlogin.SSOPacket, error) {
	fetcher.AddSeq(seq)
	err := c.tcp.Write(data)
	if err != nil {
		// 出错了要删掉
		fetcher.DeleteSeq(seq)
	}
	return fetcher.Fecth(seq, timeout)
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

func (c *QQClient) ssoHeartBeatLoop() {
heartBeatLoop:
	for {
		select {
		case <-c.stopChan:
			break heartBeatLoop
		case <-time.After(270 * time.Second):
			if !c.Online.Load() {
				continue heartBeatLoop
			}
			startTime := time.Now().UnixMilli()
			_, err := c.SendUniPacketAndAwait(
				"trpc.qq_new_tech.status_svc.StatusService.SsoHeartBeat",
				wtlogin.BuildSSOHeartbeatRequest())
			if err != nil {
				networkLogger.Errorf("heartbeat err %s", err)
			}
			networkLogger.Debugf("heartbeat %dms to server", time.Now().UnixMilli()-startTime)
		}
	}
	networkLogger.Debug("heartbeat task stoped")
}

func (c *QQClient) OnMessage(msgLen int) {
	raw, err := c.tcp.ReadBytes(msgLen)
	if err != nil {
		networkLogger.Errorf("read message error: %s", err)
		return
	}
	go func(c *QQClient, raw []byte) {
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
			if packet.RetCode != 0 && fetcher.ContainSeq(packet.Seq) {
				networkLogger.Errorf("error ssopacket retcode: %d, extra: %s", packet.RetCode, packet.Extra)
				return
			} else if packet.RetCode != 0 {
				networkLogger.Errorf("Unexpected error on sso layer: %d: %s", packet.RetCode, packet.Extra)
				return
			}
			if !fetcher.ContainSeq(packet.Seq) {
				networkLogger.Warningf("Unknown packet: %s(%d), ignore", packet.Cmd, packet.Seq)
			} else {
				fetcher.AddResult(packet.Seq, packet)
			}
		} else { // server pushed
			if fn, ok := listeners[packet.Cmd]; ok {
				networkLogger.Debugf("Server Push(%d) <- %s, extra: %s", packet.RetCode, packet.Cmd, packet.Extra)
				msg, err := fn(c, packet)
				if err != nil {
					return
				}
				go OnEvent(c, msg)
			} else {
				networkLogger.Warningf("unsupported command: %s", packet.Cmd)
			}
		}
	}(c, raw)
}

func (c *QQClient) ReadLoop() {
	for !c.tcp.IsClosed() {
		lengthData, err := c.tcp.ReadBytes(4)
		if err != nil {
			networkLogger.Errorf("tcp read length error: %s", err)
			break
		}
		length := int(binary2.NewReader(lengthData).ReadU32() - 4)
		if length > 0 {
			c.OnMessage(length)
		} else {
			c.tcp.Close()
		}
	}
	c.OnDisconnected()
}

func (c *QQClient) Loop() error {
	err := c.Connect()
	if err != nil {
		return err
	}
	go c.ReadLoop()
	go c.ssoHeartBeatLoop()
	return nil
}

func (c *QQClient) Connect() error {
	err := c.tcp.Connect(msfwifiServer, 5*time.Second)
	if err != nil {
		return err
	}
	c.OnConnected()
	return nil
}

func (c *QQClient) DisConnect() {
	c.tcp.Close()
	c.OnDisconnected()
}

// Stop 停止整个client，一旦停止不能重新连接
func (c *QQClient) Stop() {
	c.DisConnect()
	close(c.stopChan)
}

// setOnline 设置qq已经上线
func (c *QQClient) setOnline() {
	c.Online.Store(true)
}

func (c *QQClient) OnConnected() {

}

func (c *QQClient) OnDisconnected() {
	c.Online.Store(false)
}

func (c *QQClient) getAndIncreaseSequence() uint32 {
	return atomic.AddUint32(&c.sig.Sequence, 1) % 0x8000
}

func (c *QQClient) getSequence() uint32 {
	return atomic.LoadUint32(&c.sig.Sequence) % 0x8000
}
