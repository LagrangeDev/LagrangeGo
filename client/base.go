package client

// 部分借鉴 https://github.com/Mrs4s/MiraiGo/blob/master/client/client.go

import (
	"crypto/md5"
	"net/http"
	"net/http/cookiejar"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"

	"golang.org/x/net/publicsuffix"

	"github.com/LagrangeDev/LagrangeGo/utils/log"

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
func NewClient(uin uint32, password string) *QQClient {
	return NewClientMD5(uin, md5.Sum([]byte(password)))
}

func NewClientEmpty() *QQClient {
	return NewClientMD5(0, [16]byte{})
}

func NewClientMD5(uin uint32, passwordMD5 [16]byte) *QQClient {
	cookieContainer, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := &QQClient{
		Uin:         uin,
		oicq:        oicq.NewCodec(int64(uin)),
		PasswordMD5: passwordMD5,
		ticket: &TicketService{
			client: &http.Client{Jar: cookieContainer},
			sKey:   &keyInfo{},
		},
		alive: true,
	}
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

	t106        []byte
	t16a        []byte
	PasswordMD5 [16]byte

	UA string

	TCP            network.TCPClient // todo: combine other protocol state into one struct
	ConnectTime    time.Time
	transport      network.Transport
	oicq           *oicq.Codec
	logger         log.Logger
	highwaySession highway.Session
	ticket         *TicketService

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

	GroupJoinEvent  EventHandle[*event.GroupMemberIncrease] // bot进群
	GroupLeaveEvent EventHandle[*event.GroupMemberDecrease] // bot 退群

	GroupInvitedEvent                 EventHandle[*event.GroupInvite]            // 被邀请入群
	GroupMemberJoinRequestEvent       EventHandle[*event.GroupMemberJoinRequest] // 加群申请
	GroupMemberJoinEvent              EventHandle[*event.GroupMemberIncrease]    // 成员入群
	GroupMemberLeaveEvent             EventHandle[*event.GroupMemberDecrease]    // 成员退群
	GroupMuteEvent                    EventHandle[*event.GroupMute]
	GroupDigestEvent                  EventHandle[*event.GroupDigestEvent] // 精华消息
	GroupRecallEvent                  EventHandle[*event.GroupRecall]
	GroupMemberPermissionChangedEvent EventHandle[*event.GroupMemberPermissionChanged]
	GroupNameUpdatedEvent             EventHandle[*event.GroupNameUpdated]
	GroupReactionEvent                EventHandle[*event.GroupReactionEvent]
	MemberSpecialTitleUpdatedEvent    EventHandle[*event.MemberSpecialTitleUpdated]
	NewFriendEvent                    EventHandle[*event.NewFriend]
	NewFriendRequestEvent             EventHandle[*event.NewFriendRequest] // 好友申请
	FriendRecallEvent                 EventHandle[*event.FriendRecall]
	RenameEvent                       EventHandle[*event.Rename]
	FriendNotifyEvent                 EventHandle[event.INotifyEvent]
	GroupNotifyEvent                  EventHandle[event.INotifyEvent]

	// client event handles
	eventHandlers     eventHandlers
	DisconnectedEvent EventHandle[*DisconnectedEvent]
}

// AddSignServer 设置签名服务器url
func (c *QQClient) AddSignServer(signServers ...string) {
	if c.signProvider == nil {
		c.UseSignProvider(sign.NewSigner(c.debug))
	}
	c.signProvider.AddSignServer(signServers...)
}

// AddSignHeader 设置签名服务器签名时的额外http header
func (c *QQClient) AddSignHeader(header map[string]string) {
	if c.signProvider == nil {
		c.UseSignProvider(sign.NewSigner(c.debug))
	}
	c.signProvider.AddRequestHeader(header)
}

func (c *QQClient) UseVersion(app *auth.AppInfo) {
	c.transport.Version = app
	c.highwaySession.AppID = uint32(app.AppID)
	c.highwaySession.SubAppID = uint32(app.SubAppID)
	c.UA = "LagrangeGo qq/" + app.PackageSign
	if c.signProvider == nil {
		c.signProvider = sign.NewSigner(c.debug)
	}
	c.signProvider.SetAppInfo(app)
}

func (c *QQClient) UseSignProvider(p sign.Provider) {
	c.signProvider = p
	if c.version() != nil {
		c.signProvider.SetAppInfo(c.version())
	}
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

func (c *QQClient) SetCustomServer(servers []netip.AddrPort) {
	c.servers = append(servers, c.servers...)
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

func (c *QQClient) sendOidbPacketAndWait(pkt *oidb.Packet) ([]byte, error) {
	return c.sendUniPacketAndWait(pkt.Cmd, pkt.Data)
}

func (c *QQClient) sendUniPacketAndWait(cmd string, buf []byte) ([]byte, error) {
	seq, packet, err := c.uniPacket(cmd, buf)
	if err != nil {
		return nil, err
	}
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

func (c *QQClient) webSsoRequest(host, webCmd, data string) (string, error) {
	s := strings.Split(host, `.`)
	sub := ""
	for i := len(s) - 1; i >= 0; i-- {
		sub += s[i]
		if i != 0 {
			sub += "_"
		}
	}
	cmd := "MQUpdateSvc_" + sub + ".web." + webCmd
	req, _ := proto.Marshal(&service.WebSsoRequestBody{
		Type: proto.Uint32(0),
		Data: proto.Some(data),
	})
	rspData, err := c.sendUniPacketAndWait(cmd, req)
	if err != nil {
		return "", errors.Wrap(err, "send web sso request error")
	}
	rsp := &service.WebSsoResponseBody{}
	if err = proto.Unmarshal(rspData, rsp); err != nil {
		return "", errors.Wrap(err, "unmarshal response error")
	}
	return rsp.Data.Unwrap(), nil
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
