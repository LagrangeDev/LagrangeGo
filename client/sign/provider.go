package sign

import (
	"encoding/hex"
	"encoding/json"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
)

type (
	HexData []byte

	Request struct {
		Command string  `json:"command"`
		Seq     int     `json:"seq"`
		Body    HexData `json:"body"`
		Uin     uint32  `json:"uin"`
		GUID    string  `json:"guid"`
		Qua     string  `json:"qua"`
	}

	Response struct {
		Code    uint32 `json:"code"`
		Message string `json:"message"`
		Value   struct {
			SecSign  HexData `json:"sec_sign"`
			SecToken HexData `json:"sec_token"`
			SecExtra HexData `json:"sec_extra"`
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
	Sign(cmd string, seq uint32, data []byte, uin uint32, guid, qua string) (*Response, error)
	AddRequestHeader(header map[string]string)
	SetAppInfo(app *auth.AppInfo)
	Release()
}
