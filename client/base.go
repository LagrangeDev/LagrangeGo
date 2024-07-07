package client

// 部分借鉴 https://github.com/Mrs4s/MiraiGo/blob/master/client/client.go

import (
	"errors"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/RomiChan/syncx"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/client/event"
	"github.com/LagrangeDev/LagrangeGo/client/internal/cache"
	"github.com/LagrangeDev/LagrangeGo/client/internal/highway"
	"github.com/LagrangeDev/LagrangeGo/client/internal/network"
	"github.com/LagrangeDev/LagrangeGo/client/internal/oicq"
	"github.com/LagrangeDev/LagrangeGo/client/packets/oidb"
	"github.com/LagrangeDev/LagrangeGo/client/packets/wtlogin"
	"github.com/LagrangeDev/LagrangeGo/client/sign"
	"github.com/LagrangeDev/LagrangeGo/message"
)

// NewClient 创建一个新的 QQ Client
func NewClient(uin uint32, signUrl string, appInfo *auth.AppInfo) *QQClient {
	client := &QQClient{
		Uin:  uin,
		oicq: oicq.NewCodec(int64(uin)),
		highwaySession: highway.Session{
			AppID:    uint32(appInfo.AppID),
			SubAppID: uint32(appInfo.SubAppID),
		},
		alive: true,
	}
	client.signProvider = sign.NewProviderURL(signUrl, func(msg string) {
		client.debugln(msg)
	})
	client.transport.Version = appInfo
	client.transport.Sig.D2Key = make([]byte, 0, 16)
	client.highwaySession.Uin = &client.transport.Sig.Uin
	client.Online.Store(false)
	client.TCP.PlannedDisconnect(client.plannedDisconnect)
	client.TCP.UnexpectedDisconnect(client.unexpectedDisconnect)
	return client
}

type QQClient struct {
	Uin          uint32
	signProvider sign.Provider

	stat Statistics
	once sync.Once

	Online atomic.Bool

	t106 []byte
	t16a []byte

	TCP            network.TCPClient // todo: combine other protocol state into one struct
	ConnectTime    time.Time
	transport      network.Transport
	oicq           *oicq.Codec
	logger         Logger
	highwaySession highway.Session

	// internal state
	handlers        syncx.Map[uint32, *handlerInfo]
	waiters         syncx.Map[string, func(any, error)]
	initServerOnce  sync.Once
	servers         []netip.AddrPort
	currServerIndex int
	retryTimes      int
	alive           bool

	cache cache.Cache

	// event handles
	GroupMessageEvent   EventHandle[*message.GroupMessage]
	PrivateMessageEvent EventHandle[*message.PrivateMessage]
	TempMessageEvent    EventHandle[*message.TempMessage]

	SelfGroupMessageEvent   EventHandle[*message.GroupMessage]
	SelfPrivateMessageEvent EventHandle[*message.PrivateMessage]
	SelfTempMessageEvent    EventHandle[*message.TempMessage]

	GroupInvitedEvent           EventHandle[*event.GroupInvite]            // 被邀请入群
	GroupMemberJoinRequestEvent EventHandle[*event.GroupMemberJoinRequest] // 加群申请
	GroupMemberJoinEvent        EventHandle[*event.GroupMemberIncrease]    // 成员入群
	GroupMemberLeaveEvent       EventHandle[*event.GroupMemberDecrease]    // 成员退群
	GroupMuteEvent              EventHandle[*event.GroupMute]
	GroupDigestEvent            EventHandle[*event.GroupDigestEvent] // 精华消息
	GroupRecallEvent            EventHandle[*event.GroupRecall]
	NewFriendRequestEvent       EventHandle[*event.NewFriendRequest] // 好友申请
	FriendRecallEvent           EventHandle[*event.FriendRecall]
	RenameEvent                 EventHandle[*event.Rename]

	// client event handles
	eventHandlers     eventHandlers
	DisconnectedEvent EventHandle[*ClientDisconnectedEvent]
}

func (c *QQClient) version() *auth.AppInfo {
	return c.transport.Version
}

func (c *QQClient) Device() *auth.DeviceInfo {
	return c.transport.Device
}

func (c *QQClient) UseDevice(d *auth.DeviceInfo) {
	c.transport.Device = d
}

func (c *QQClient) UseSig(s auth.SigInfo) {
	c.transport.Sig = s
}

func (c *QQClient) Sig() *auth.SigInfo {
	return &c.transport.Sig
}

func (c *QQClient) Release() {
	if c.Online.Load() {
		c.Disconnect()
	}
	c.alive = false
}

func (c *QQClient) NickName() string {
	return c.transport.Sig.Nickname
}

func (c *QQClient) sendOidbPacketAndWait(pkt *oidb.OidbPacket) ([]byte, error) {
	return c.sendUniPacketAndWait(pkt.Cmd, pkt.Data)
}

func (c *QQClient) sendUniPacketAndWait(cmd string, buf []byte) ([]byte, error) {
	seq, packet := c.uniPacket(cmd, buf)
	pkt, err := c.sendAndWait(seq, packet)
	if err != nil {
		return nil, err
	}
	rsp, ok := pkt.([]byte)
	if !ok {
		return nil, errors.New("cannot parse response to bytes")
	}
	return rsp, nil
}

func (c *QQClient) doHeartbeat() {
	for c.Online.Load() {
		time.Sleep(270 * time.Second)
		if !c.Online.Load() {
			continue
		}
		startTime := time.Now().UnixMilli()
		_, err := c.sendUniPacketAndWait(
			"trpc.qq_new_tech.status_svc.StatusService.SsoHeartBeat",
			wtlogin.BuildSSOHeartbeatRequest())
		if errors.Is(err, network.ErrConnectionClosed) {
			continue
		}
		if err != nil {
			c.error("heartbeat err %s", err)
		} else {
			c.debug("heartbeat %dms to server", time.Now().UnixMilli()-startTime)
			//TODO: times
		}
	}
	c.debugln("heartbeat task stoped")
}

// setOnline 设置qq已经上线
func (c *QQClient) setOnline() {
	c.Online.Store(true)
}

func (c *QQClient) getAndIncreaseSequence() uint32 {
	return atomic.AddUint32(&c.transport.Sig.Sequence, 1) % 0x8000
}

func (c *QQClient) getSequence() uint32 {
	return atomic.LoadUint32(&c.transport.Sig.Sequence) % 0x8000
}
