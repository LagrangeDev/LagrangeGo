package sign

import (
	"encoding/hex"
	"encoding/json"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
)

type (
	HexData []byte

	Request struct {
		Cmd string  `json:"cmd"`
		Seq int     `json:"seq"`
		Src HexData `json:"src"`
	}

	Response struct {
		Platform string `json:"platform"`
		Version  string `json:"version"`
		Value    struct {
			Sign  HexData `json:"sign"`
			Extra HexData `json:"extra"`
			Token HexData `json:"token"`
		} `json:"value"`
	}
)

func (h HexData) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(h))
}

func (h *HexData) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return err
	}
	*h = decoded
	return nil
}

type Provider interface {
	Sign(cmd string, seq uint32, data []byte) (*Response, error)
	AddRequestHeader(header map[string]string)
	AddSignServer(signServers ...string)
	GetSignServer() []string
	SetAppInfo(app *auth.AppInfo)
}
