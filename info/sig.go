package info

type SigInfo struct {
	Uin         uint32
	Uid         string
	Sequence    int
	Tgtgt       []byte
	Tgt         []byte
	D2          []byte
	D2Key       []byte
	Qrsig       []byte
	ExchangeKey []byte
	KeySig      []byte
	Cookies     string
	UnusualSig  []byte
	TempPwd     []byte
	CaptchaInfo [3]string

	Nickname string
	Age      uint8
	Gender   uint8
}

func NewSigInfo(seq int) *SigInfo {
	return &SigInfo{
		Uid:         "",
		Sequence:    seq,
		Tgtgt:       make([]byte, 0),
		Tgt:         make([]byte, 0),
		D2:          make([]byte, 0),
		D2Key:       make([]byte, 16),
		Qrsig:       make([]byte, 0),
		ExchangeKey: make([]byte, 0),
		KeySig:      make([]byte, 0),
		Cookies:     "",
		UnusualSig:  make([]byte, 0),
		TempPwd:     make([]byte, 0),
		CaptchaInfo: [3]string{"", "", ""},
	}
}
