package sign

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
)

type (
	Client struct {
		lock         sync.RWMutex
		server       string
		httpClient   *http.Client
		extraHeaders http.Header
		log          func(string, ...any)
	}
)

var ErrVersionMismatch = errors.New("sign version mismatch")

func NewSigner(log func(string, ...any), token, signServer string) *Client {
	if log == nil {
		log = func(string, ...any) {}
	}
	client := &Client{
		server:     signServer,
		httpClient: &http.Client{},
		extraHeaders: http.Header{
			"Authorization": []string{"Bearer " + token},
		},
		log: log,
	}
	return client
}

func (c *Client) Release() {}

func (c *Client) AddRequestHeader(header map[string]string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for k, v := range header {
		c.extraHeaders.Add(k, v)
	}
}

func (c *Client) SetAppInfo(app *auth.AppInfo) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.extraHeaders.Set("User-Agent", "qq/"+app.CurrentVersion)
}

func (c *Client) signServer() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.server
}

func (c *Client) Sign(cmd string, seq uint32, data []byte, uin uint32, guid, qua string) (*Response, error) {
	if !ContainSignPKG(cmd) {
		return nil, nil
	}
	server := c.signServer()
	if server == "" {
		return nil, errors.New("sign server not configured")
	}
	startTime := time.Now().UnixMilli()
	resp, err := sign(server, cmd, seq, data, uin, guid, qua, c.extraHeaders)
	if err != nil {
		return nil, err
	}
	c.log(fmt.Sprintf("signed for [%s:%d](%dms)",
		cmd, seq, time.Now().UnixMilli()-startTime))
	return resp, nil
}

func sign(server, cmd string, seq uint32, buf []byte, uin uint32, guid, qua string, header http.Header) (*Response, error) {
	if !ContainSignPKG(cmd) {
		return nil, nil
	}
	req := Request{
		Command: cmd,
		Seq:     int(seq),
		Body:    buf,
		Uin:     uin,
		GUID:    guid,
		Qua:     qua,
	}
	data, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	resp, err := httpPost[Response](server, bytes.NewReader(data), 8*time.Second, header)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
