package info

type SigInfo struct {
	Uin         uint32
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
	Uid         string
	CaptchaInfo [3]string
}

func NewSigInfo(seq int) *SigInfo {
	return &SigInfo{
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
		Uid:         "",
		CaptchaInfo: [3]string{"", "", ""},
	}
}
