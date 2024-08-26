package oidb

import "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"

func BuildFetchCookieReq(domains []string) (*OidbPacket, error) {
	body := oidb.OidbSvcTrpcTcp0X102A_0{Domain: domains}
	return BuildOidbPacket(0x102A, 0, &body, false, false)
}

func ParseFetchCookieResp(data []byte) ([]string, error) {
	resp, err := ParseTypedError[oidb.OidbSvcTrpcTcp0X102A_0Response](data)
	if err != nil {
		return nil, err
	}
	cookies := make([]string, len(resp.Urls))
	for i, urls := range resp.Urls {
		cookies[i] = string(urls.Value)
	}
	return cookies, nil
}
