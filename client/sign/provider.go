package sign

import "github.com/LagrangeDev/LagrangeGo/client/auth"

type Response struct {
	Platform string `json:"platform"`
	Version  string `json:"version"`
	Value    struct {
		Sign  string `json:"sign"`
		Extra string `json:"extra"`
		Token string `json:"token"`
	} `json:"value"`
}

type Provider interface {
	Sign(cmd string, seq uint32, data []byte) (*Response, error)
	AddRequestHeader(header map[string]string)
	AddSignServer(signServers ...string)
	GetSignServer() []string
	SetAppInfo(app *auth.AppInfo)
}
