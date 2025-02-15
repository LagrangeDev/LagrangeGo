package sign

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

var (
	signMap    = map[string]struct{}{} // 只在启动时初始化, 无并发问题
	httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}
)

func init() {
	signPkgList := []string{
		"trpc.o3.ecdh_access.EcdhAccess.SsoEstablishShareKey",
		"trpc.o3.ecdh_access.EcdhAccess.SsoSecureAccess",
		"trpc.o3.report.Report.SsoReport",
		"MessageSvc.PbSendMsg",
		//"wtlogin.trans_emp",
		"wtlogin.login",
		//"trpc.login.ecdh.EcdhService.SsoKeyExchange",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLogin",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginNewDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLoginUnusualDevice",
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginUnusualDevice",
		"OidbSvcTrpcTcp.0x11ec_1",
		"OidbSvcTrpcTcp.0x758_1", // create group
		"OidbSvcTrpcTcp.0x7c1_1",
		"OidbSvcTrpcTcp.0x7c2_5", // request friend
		"OidbSvcTrpcTcp.0x10db_1",
		"OidbSvcTrpcTcp.0x8a1_7", // request group
		"OidbSvcTrpcTcp.0x89a_0",
		"OidbSvcTrpcTcp.0x89a_15",
		"OidbSvcTrpcTcp.0x88d_0", // fetch group detail
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
		"OidbSvcTrpcTcp.0x6d9_4",
	}

	for _, cmd := range signPkgList {
		signMap[cmd] = struct{}{}
	}
}

func ContainSignPKG(cmd string) bool {
	_, ok := signMap[cmd]
	return ok
}

func httpGet[T any](rawURL string, queryParams map[string]string, timeout time.Duration, header http.Header) (target T, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
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
		return
	}
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	return doHTTP[T](ctx, req)
}

func httpPost[T any](rawURL string, body io.Reader, timeout time.Duration, header http.Header) (target T, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		err = errors.Wrap(err, "create POST")
		return
	}
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	req.Header.Add("Content-Type", "application/json")
	return doHTTP[T](ctx, req)
}

//nolint:bodyclose
func doHTTP[T any](ctx context.Context, req *http.Request) (target T, err error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			err = ctx.Err()
			return
		}
		resp, err = httpClient.Do(req)
		if err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				err = ctx.Err()
				return
			}
			err = errors.Wrap(err, "perform POST")
			return
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	if err = json.NewDecoder(resp.Body).Decode(&target); err != nil {
		err = errors.Wrap(err, "unmarshal response")
		return
	}

	return
}
