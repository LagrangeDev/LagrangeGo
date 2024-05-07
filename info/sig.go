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

func NewSigInfo(seq int) SigInfo {
	return SigInfo{
		Sequence: seq,
		D2Key:    make([]byte, 16),
	}
}
