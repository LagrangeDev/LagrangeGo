package sign

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/http2"

	"github.com/LagrangeDev/LagrangeGo/utils"
)

var (
	signMap = map[string]struct{}{} // 只在启动时初始化, 无并发问题
)

func init() {
	signPkgList := []string{
		"trpc.o3.ecdh_access.EcdhAccess.SsoEstablishShareKey",
		"trpc.o3.ecdh_access.EcdhAccess.SsoSecureAccess",
		"trpc.o3.report.Report.SsoReport",
		"MessageSvc.PbSendMsg",
		// "wtlogin.trans_emp",
		"wtlogin.login",
		// "trpc.login.ecdh.EcdhService.SsoKeyExchange",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginNewDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLoginUnusualDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginUnusualDevice",
		"OidbSvcTrpcTcp.0x11ec_1",
		"OidbSvcTrpcTcp.0x758_1",
		"OidbSvcTrpcTcp.0x7c2_5",
		"OidbSvcTrpcTcp.0x10db_1",
		"OidbSvcTrpcTcp.0x8a1_7",
		"OidbSvcTrpcTcp.0x89a_0",
		"OidbSvcTrpcTcp.0x89a_15",
		"OidbSvcTrpcTcp.0x88d_0",
		"OidbSvcTrpcTcp.0x88d_14",
		"OidbSvcTrpcTcp.0x112a_1",
		"OidbSvcTrpcTcp.0x587_74",
		"OidbSvcTrpcTcp.0x1100_1",
		"OidbSvcTrpcTcp.0x1102_1",
		"OidbSvcTrpcTcp.0x1103_1",
		"OidbSvcTrpcTcp.0x1107_1",
		"OidbSvcTrpcTcp.0x1105_1",
		"OidbSvcTrpcTcp.0xf88_1",
		"OidbSvcTrpcTcp.0xf89_1",
		"OidbSvcTrpcTcp.0xf57_1",
		"OidbSvcTrpcTcp.0xf57_106",
		"OidbSvcTrpcTcp.0xf57_9",
		"OidbSvcTrpcTcp.0xf55_1",
		"OidbSvcTrpcTcp.0xf67_1",
		"OidbSvcTrpcTcp.0xf67_5",
	}

	for _, cmd := range signPkgList {
		signMap[cmd] = struct{}{}
	}
}

func containSignPKG(cmd string) bool {
	_, ok := signMap[cmd]
	return ok
}

func NewProviderURL(log func(msg string), rawUrls ...string) []Provider {
	if len(rawUrls) == 0 {
		return nil
	}
	providers := make([]Provider, len(rawUrls))
	for i, rawUrl := range rawUrls {
		localRawUrl := rawUrl
		providers[i] = func(cmd string, seq uint32, buf []byte, header map[string]string) (map[string]string, error) {
			if !containSignPKG(cmd) {
				return nil, nil
			}
			startTime := time.Now().UnixMilli()
			resp := signResponse{}
			sb := strings.Builder{}
			sb.WriteString(`{"cmd":"` + cmd + `",`)
			sb.WriteString(`"seq":` + strconv.Itoa(int(seq)) + `,`)
			sb.WriteString(`"src":"` + fmt.Sprintf("%x", buf) + `"}`)
			newHeaders := map[string]string{
				"Content-Type": "application/json",
			}
			for k, v := range header {
				newHeaders[k] = v
			}
			err := httpPost(localRawUrl, bytes.NewReader(utils.S2B(sb.String())), 8*time.Second, &resp, newHeaders)
			if err != nil || resp.Value.Sign == "" {
				err := httpGet(localRawUrl, map[string]string{
					"cmd": cmd,
					"seq": strconv.Itoa(int(seq)),
					"src": fmt.Sprintf("%x", buf),
				}, 8*time.Second, &resp, header)
				if err != nil {
					log(err.Error())
					return nil, err
				}
			}

			log(fmt.Sprintf("signed for [%s:%d](%dms)",
				cmd, seq, time.Now().UnixMilli()-startTime))

			return map[string]string{
				"sign":  resp.Value.Sign,
				"extra": resp.Value.Extra,
				"token": resp.Value.Token,
			}, nil
		}
	}
	return providers
}

func httpGet(rawUrl string, queryParams map[string]string, timeout time.Duration, target interface{}, header map[string]string) error {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create GET request: %w", err)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := http2.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("request timed out")
		}
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("request timed out")
			}
			return fmt.Errorf("failed to perform GET request: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return nil
}

func httpPost(rawUrl string, body io.Reader, timeout time.Duration, target interface{}, header map[string]string) error {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create POST request: %w", err)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := http2.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("request timed out")
		}
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("request timed out")
			}
			return fmt.Errorf("failed to perform POST request: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return nil
}

type signResponse struct {
	Value struct {
		Sign  string `json:"sign"`
		Extra string `json:"extra"`
		Token string `json:"token"`
	} `json:"value"`
}
