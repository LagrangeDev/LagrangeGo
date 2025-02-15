package sign

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

type (
	Client struct {
		lock         sync.RWMutex
		signCount    atomic.Uint32
		instances    []*remote
		app          *auth.AppInfo
		httpClient   *http.Client
		extraHeaders http.Header
		log          func(string, ...any)
		lastTestTime time.Time
	}

	remote struct {
		server  string
		latency atomic.Uint32
	}
)

const (
	serverLatencyDown = math.MaxUint32
)

var ErrVersionMismatch = errors.New("sign version mismatch")

func NewSigner(log func(string, ...any), signServers ...string) *Client {
	client := &Client{
		instances: utils.Map(signServers, func(s string) *remote {
			return &remote{server: s}
		}),
		httpClient:   &http.Client{},
		extraHeaders: http.Header{},
		log:          log,
	}
	go client.test()
	return client
}

func (c *Client) AddRequestHeader(header map[string]string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for k, v := range header {
		c.extraHeaders.Add(k, v)
	}
}

func (c *Client) AddSignServer(signServers ...string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.instances = append(c.instances, utils.Map[string, *remote](signServers, func(s string) *remote {
		return &remote{server: s}
	})...)
}

func (c *Client) GetSignServer() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return utils.Map(c.instances, func(sign *remote) string {
		return sign.server
	})
}

func (c *Client) SetAppInfo(app *auth.AppInfo) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.app = app
	c.extraHeaders.Set("User-Agent", "qq/"+app.CurrentVersion)
}

func (c *Client) getAvailableSign() *remote {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, i := range c.instances {
		if i.latency.Load() < serverLatencyDown {
			return i
		}
	}
	return nil
}

func (c *Client) sortByLatency() {
	c.lock.Lock()
	defer c.lock.Unlock()
	sort.SliceStable(c.instances, func(i, j int) bool {
		return c.instances[i].latency.Load() < c.instances[j].latency.Load()
	})
}

func (c *Client) Sign(cmd string, seq uint32, data []byte) (*Response, error) {
	if !ContainSignPKG(cmd) {
		return nil, nil
	}
	if time.Now().After(c.lastTestTime.Add(30 * time.Minute)) {
		go c.test()
	}
	startTime := time.Now().UnixMilli()
	for {
		if sign := c.getAvailableSign(); sign != nil {
			resp, err := sign.sign(cmd, seq, data, c.extraHeaders)
			if err != nil {
				sign.latency.Store(serverLatencyDown)
				continue
			} else if resp.Version != c.app.CurrentVersion && resp.Value.Extra != c.app.SignExtraHexLower && resp.Value.Extra != c.app.SignExtraHexUpper {
				return nil, ErrVersionMismatch
			}
			c.log(fmt.Sprintf("signed for [%s:%d](%dms)",
				cmd, seq, time.Now().UnixMilli()-startTime))
			c.signCount.Add(1)
			return resp, nil
		}
		break
	}
	// 全寄了，重新再测下
	go c.test()
	return nil, errors.New("all sign service down")
}

func (c *Client) test() {
	c.lock.Lock()
	if time.Now().Before(c.lastTestTime.Add(10 * time.Minute)) {
		c.lock.Unlock()
		return
	}
	c.lastTestTime = time.Now()
	c.lock.Unlock()
	for _, i := range c.instances {
		i.test()
	}
	c.sortByLatency()
}

func (i *remote) sign(cmd string, seq uint32, buf []byte, header http.Header) (*Response, error) {
	if !ContainSignPKG(cmd) {
		return nil, nil
	}
	sb := strings.Builder{}
	sb.WriteString(`{"cmd":"` + cmd + `",`)
	sb.WriteString(`"seq":` + strconv.Itoa(int(seq)) + `,`)
	sb.WriteString(`"src":"` + fmt.Sprintf("%x", buf) + `"}`)
	resp, err := httpPost[Response](i.server, bytes.NewReader(utils.S2B(sb.String())), 8*time.Second, header)
	if err != nil || resp.Value.Sign == "" {
		resp, err = httpGet[Response](i.server, map[string]string{
			"cmd": cmd,
			"seq": strconv.Itoa(int(seq)),
			"src": fmt.Sprintf("%x", buf),
		}, 8*time.Second, header)
		if err != nil {
			return nil, err
		}
	}
	return &resp, nil
}

func (i *remote) test() {
	startTime := time.Now().UnixMilli()
	resp, err := i.sign("wtlogin.login", 1, []byte{11, 45, 14}, nil)
	if err != nil || resp.Value.Sign == "" {
		i.latency.Store(serverLatencyDown)
		return
	}
	// 有长连接的情况，取两次平均值
	resp, err = i.sign("wtlogin.login", 1, []byte{11, 45, 14}, nil)
	if err != nil || resp.Value.Sign == "" {
		i.latency.Store(serverLatencyDown)
		return
	}
	// 粗略计算，应该足够了
	i.latency.Store(uint32(time.Now().UnixMilli()-startTime) / 2)
}
