package ecdh

import (
	"encoding/hex"
)

var (
	ECDH               map[string]*BaseECDH
	EcdhPrimePublic, _ = hex.DecodeString("049D1423332735980EDABE7E9EA451B3395B6F35250DB8FC56F25889F628CBAE3E8E73077914071EEEBC108F4E0170057792BB17AA303AF652313D17C1AC815E79")
	EcdhSecpPublic, _  = hex.DecodeString("04928D8850673088B343264E0C6BACB8496D697799F37211DEB25BB73906CB089FEA9639B4E0260498B51A992D50813DA8")
)

type BaseECDH struct {
	provider    *ECDHProvider
	publicKey   []byte
	shareKey    []byte
	compressKey bool
}

func (e *BaseECDH) GetPublicKey() []byte {
	return e.publicKey
}

func (e *BaseECDH) GetShareKey() []byte {
	return e.shareKey
}

func (e *BaseECDH) Exange(newKey []byte) []byte {
	return e.provider.keyExchange(newKey, e.compressKey)
}

// newECDHPrime exchange key
func newECDHPrime() (baseECDH *BaseECDH) {
	baseECDH = &BaseECDH{
		provider:    newECDHProvider(CURVE["prime256v1"]),
		publicKey:   nil,
		shareKey:    nil,
		compressKey: false,
	}
	baseECDH.publicKey = baseECDH.provider.packPublic(false)
	baseECDH.shareKey = baseECDH.provider.keyExchange(EcdhPrimePublic, false)
	return
}

// newECDHSecp login and others
func newECDHSecp() (baseECDH *BaseECDH) {
	baseECDH = &BaseECDH{
		provider:    newECDHProvider(CURVE["secp192k1"]),
		publicKey:   nil,
		shareKey:    nil,
		compressKey: true,
	}
	baseECDH.publicKey = baseECDH.provider.packPublic(true)
	baseECDH.shareKey = baseECDH.provider.keyExchange(EcdhSecpPublic, true)
	return
}

func initImpl() {
	ECDH = map[string]*BaseECDH{
		"secp192k1":  newECDHSecp(),
		"prime256v1": newECDHPrime(),
	}
}
