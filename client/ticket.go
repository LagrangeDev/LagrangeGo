package client

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/publicsuffix"

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
	}
)

var (
	sKey               = keyInfo{}
	cookieContainer, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	ticketService      = &TicketService{
		client: &http.Client{
			Jar: cookieContainer,
		},
	}
)

func SendRequest(request *http.Request) (*http.Response, error) {
	// 应该不需要考虑cookie的问题
	return ticketService.client.Do(request)
}

func (c *QQClient) GetSkey() (string, error) {
	if time.Now().Before(sKey.expireTime) {
		return sKey.key, nil
	}
	clientKey, err := c.FetchClientKey()
	if err != nil {
		return "", err
	}
	jump := "https%3A%2F%2Fh5.qzone.qq.com%2Fqqnt%2Fqzoneinpcqq%2Ffriend%3Frefresh%3D0%26clientuin%3D0%26darkMode%3D0&keyindex=19&random=2599"
	u, _ := url.Parse(fmt.Sprintf("https://ssl.ptlogin2.qq.com/jump?ptlang=1033&clientuin=%d&clientkey=%s&u1=%s",
		c.Uin, clientKey, jump))
	_, err = ticketService.client.Get(u.String())
	if err != nil {
		return "", err
	}
	for _, cookie := range cookieContainer.Cookies(u) {
		if cookie.Name == "skey" {
			sKey.key = cookie.Value
			sKey.expireTime = time.Now().Add(24 * time.Hour)
			break
		}
	}
	return sKey.key, nil
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

func (c *QQClient) GetCookies(domain string) (string, error) {
	skey, err := c.GetSkey()
	if err != nil {
		return "", err
	}
	var token string
	if tokenTime, ok := ticketService.psKeys.Load(domain); ok {
		if time.Now().Before(tokenTime.expireTime) {
			cookies, err := c.FetchCookies([]string{domain})
			if err != nil {
				return "", err
			}
			token = cookies[0]
			ticketService.psKeys.Store(domain, &keyInfo{
				key:        token,
				expireTime: time.Now().Add(24 * time.Hour),
			})
		} else {
			token = tokenTime.key
		}
	} else {
		cookies, err := c.FetchCookies([]string{domain})
		if err != nil {
			return "", err
		}
		token = cookies[0]
		ticketService.psKeys.Store(domain, &keyInfo{
			key:        token,
			expireTime: time.Now().Add(24 * time.Hour),
		})
	}
	return fmt.Sprintf("p_uin=o%d; p_skey=%s; skey=%s; uin=o%d", c.Uin, token, skey, c.Uin), nil
}
