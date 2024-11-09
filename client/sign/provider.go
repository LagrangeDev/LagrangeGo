package sign

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
}
