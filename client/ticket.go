package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/RomiChan/syncx"
)

type (
	keyInfo struct {
		key        string
		expireTime time.Time
	}

	TicketService struct {
		client *http.Client
		psKeys syncx.Map[string, *keyInfo]
		sKey   *keyInfo
	}

	Cookies struct {
		uin   uint32
		SKey  string
		PsKey string
	}
)

func (c *QQClient) SendRequestWithCookie(request *http.Request, domain string) (*http.Response, error) {
	cookies, err := c.GetCookies(domain)
	if err != nil {
		return nil, err
	}
	request.AddCookie(&http.Cookie{Name: "skey", Value: cookies.SKey})
	request.AddCookie(&http.Cookie{Name: "p_uin", Value: strconv.Itoa(int(cookies.uin))})
	request.AddCookie(&http.Cookie{Name: "p_skey", Value: cookies.PsKey})
	return c.ticket.client.Do(request)
}

func (c *QQClient) GetSkey() (string, error) {
	if time.Now().Before(c.ticket.sKey.expireTime) {
		return c.ticket.sKey.key, nil
	}
	clientKey, err := c.FetchClientKey()
	if err != nil {
		return "", err
	}
	jump := "https%3A%2F%2Fh5.qzone.qq.com%2Fqqnt%2Fqzoneinpcqq%2Ffriend%3Frefresh%3D0%26clientuin%3D0%26darkMode%3D0&keyindex=19&random=2599"
	u, _ := url.Parse(fmt.Sprintf("https://ssl.ptlogin2.qq.com/jump?ptlang=1033&clientuin=%d&clientkey=%s&u1=%s",
		c.Uin, clientKey, jump))
	_, err = c.ticket.client.Get(u.String())
	if err != nil {
		return "", err
	}
	for _, cookie := range c.ticket.client.Jar.Cookies(u) {
		if cookie.Name == "skey" {
			c.ticket.sKey.key = cookie.Value
			c.ticket.sKey.expireTime = time.Now().Add(24 * time.Hour)
			break
		}
	}
	return c.ticket.sKey.key, nil
}

func (c *QQClient) GetCsrfToken() (int, error) {
	skey, err := c.GetSkey()
	if err != nil {
		return -1, err
	}

	hash := 5381
	for _, ch := range skey {
		hash += (hash << 5) + int(ch)
	}
	return hash & 2147483647, nil
}

func (c *QQClient) GetCookies(domain string) (*Cookies, error) {
	skey, err := c.GetSkey()
	if err != nil {
		return nil, err
	}
	var token string
	if tokenTime, ok := c.ticket.psKeys.Load(domain); ok {
		if time.Now().Before(tokenTime.expireTime) {
			cookies, err := c.FetchCookies([]string{domain})
			if err != nil {
				return nil, err
			}
			token = cookies[0]
			c.ticket.psKeys.Store(domain, &keyInfo{
				key:        token,
				expireTime: time.Now().Add(24 * time.Hour),
			})
		} else {
			token = tokenTime.key
		}
	} else {
		cookies, err := c.FetchCookies([]string{domain})
		if err != nil {
			return nil, err
		}
		token = cookies[0]
		c.ticket.psKeys.Store(domain, &keyInfo{
			key:        token,
			expireTime: time.Now().Add(24 * time.Hour),
		})
	}
	return &Cookies{
		uin:   c.Uin,
		SKey:  skey,
		PsKey: token,
	}, nil
}

func GTK(s string) int {
	hash := 5381
	for _, ch := range s {
		hash += (hash<<5)&2147483647 + int(ch)&2147483647
		hash &= 2147483647
	}
	return hash
}
