package sign

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LagrangeDev/LagrangeGo/utils"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
)

type (
	Status uint32

	Client struct {
		lock         sync.RWMutex
		signCount    atomic.Uint32
		instances    []*Instance
		app          *auth.AppInfo
		httpClient   *http.Client
		extraHeaders http.Header
		log          func(string)
		lastTestTime time.Time
	}

	Instance struct {
		server  string
		latency atomic.Uint32
		status  atomic.Uint32
	}
)

const (
	OK Status = iota
	Down
)

var VersionMismatchError = errors.New("sign version mismatch")

func NewSignClient(appinfo *auth.AppInfo, log func(string), signServers ...string) *Client {
	client := &Client{
		instances: utils.Map[string, *Instance](signServers, func(s string) *Instance {
			return &Instance{server: s}
		}),
		app:        appinfo,
		httpClient: &http.Client{},
		extraHeaders: http.Header{
			"User-Agent": []string{"qq/" + appinfo.CurrentVersion},
		},
		log: log,
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
	c.instances = append(c.instances, utils.Map[string, *Instance](signServers, func(s string) *Instance {
		return &Instance{server: s}
	})...)
}

func (c *Client) GetSignServer() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return utils.Map[*Instance, string](c.instances, func(sign *Instance) string {
		return sign.server
	})
}

func (c *Client) getAvailableSign() *Instance {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, i := range c.instances {
		if Status(i.status.Load()) == OK {
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
	if !containSignPKG(cmd) {
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
				sign.status.Store(uint32(Down))
				continue
			} else if resp.Version != c.app.CurrentVersion && resp.Value.Extra != c.app.SignExtraHexLower && resp.Value.Extra != c.app.SignExtraHexUpper {
				return nil, VersionMismatchError
			}
			c.log(fmt.Sprintf("signed for [%s:%d](%dms)",
				cmd, seq, time.Now().UnixMilli()-startTime))
			c.signCount.Add(1)
			return resp, nil
		} else {
			break
		}
	}
	// 全寄了，重新再测下
	go c.test()
	return nil, errors.New("all sign service down")
}

func (c *Client) test() {
	c.lock.Lock()
	if time.Now().Before(c.lastTestTime.Add(10 * time.Minute)) {
		return
	}
	c.lastTestTime = time.Now()
	c.lock.Unlock()
	for _, i := range c.instances {
		i.test()
	}
	c.sortByLatency()
}

func (i *Instance) sign(cmd string, seq uint32, buf []byte, header http.Header) (*Response, error) {
	if !containSignPKG(cmd) {
		return nil, nil
	}
	resp := Response{}
	sb := strings.Builder{}
	sb.WriteString(`{"cmd":"` + cmd + `",`)
	sb.WriteString(`"seq":` + strconv.Itoa(int(seq)) + `,`)
	sb.WriteString(`"src":"` + fmt.Sprintf("%x", buf) + `"}`)
	err := httpPost(i.server, bytes.NewReader(utils.S2B(sb.String())), 8*time.Second, &resp, header)
	if err != nil || resp.Value.Sign == "" {
		err := httpGet(i.server, map[string]string{
			"cmd": cmd,
			"seq": strconv.Itoa(int(seq)),
			"src": fmt.Sprintf("%x", buf),
		}, 8*time.Second, &resp, header)
		if err != nil {
			return nil, err
		}
	}
	return &resp, nil
}

func (i *Instance) test() {
	startTime := time.Now().UnixMilli()
	resp, err := i.sign("wtlogin.login", 1, []byte{11, 45, 14}, nil)
	if err != nil || resp.Value.Sign == "" {
		i.status.Store(uint32(Down))
		i.latency.Store(99999)
		return
	}
	// 有长连接的情况，取两次平均值
	resp, err = i.sign("wtlogin.login", 1, []byte{11, 45, 14}, nil)
	if err != nil || resp.Value.Sign == "" {
		i.status.Store(uint32(Down))
		i.latency.Store(99999)
		return
	}
	// 粗略计算，应该足够了
	i.latency.Store(uint32(time.Now().UnixMilli()-startTime) / 2)
}
