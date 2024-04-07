package info

import (
	"fmt"
	"testing"
)

func TestS(t *testing.T) {
	sig := SigInfo{
		Sequence:    256,
		Tgtgt:       []byte{2, 3},
		Tgt:         []byte{8, 5},
		D2:          nil,
		D2Key:       nil,
		Qrsig:       nil,
		ExchangeKey: nil,
		KeySig:      nil,
		Cookies:     "",
		UnusualSig:  nil,
		TempPwd:     nil,
		Uid:         "",
		CaptchaInfo: [3]string{"123", "456", "789"},
	}
	encoded := Encode(&sig)
	fmt.Println(len(encoded))
	nsig := Decode(encoded, true)
	fmt.Println(nsig.Sequence, nsig.Tgtgt, nsig.CaptchaInfo)
}
