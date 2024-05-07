package ecdh

import (
	"crypto/ecdh"
	"crypto/rand"
)

type Exchanger interface {
	PublicKey() []byte
	SharedKey() []byte
	Exange(remote []byte) ([]byte, error)
}

type s192exchanger struct {
	provider    *provider
	publicKey   []byte
	shareKey    []byte
	compressKey bool
}

func (e *s192exchanger) PublicKey() []byte {
	return e.publicKey
}

func (e *s192exchanger) SharedKey() []byte {
	return e.shareKey
}

func (e *s192exchanger) Exange(newKey []byte) ([]byte, error) {
	return e.provider.keyExchange(newKey, e.compressKey)
}

func news192exchanger() (baseECDH *s192exchanger) {
	p, err := newProvider(newS192Curve())
	if err != nil {
		panic(err)
	}
	shk, err := p.keyExchange(ecdhS192PublicBytes, true)
	if err != nil {
		panic(err)
	}
	baseECDH = &s192exchanger{
		provider:    p,
		publicKey:   p.packPublic(true),
		shareKey:    shk,
		compressKey: true,
	}
	return
}

type p256exchanger struct {
	privateKey *ecdh.PrivateKey
	publicKey  *ecdh.PublicKey
	shareKey   []byte
}

func (e *p256exchanger) PublicKey() []byte {
	return e.publicKey.Bytes()
}

func (e *p256exchanger) SharedKey() []byte {
	return e.shareKey
}

func (e *p256exchanger) Exange(remote []byte) ([]byte, error) {
	rk, err := e.publicKey.Curve().NewPublicKey(remote)
	if err != nil {
		return nil, err
	}
	return e.privateKey.ECDH(rk)
}

func newp256exchanger() (baseECDH *p256exchanger) {
	p256 := ecdh.P256()
	privateKey, err := p256.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	pk, err := p256.NewPublicKey(ecdhP256PublicBytes)
	if err != nil {
		panic(err)
	}
	shk, err := privateKey.ECDH(pk)
	if err != nil {
		panic(err)
	}
	return &p256exchanger{
		privateKey: privateKey,
		publicKey:  privateKey.PublicKey(),
		shareKey:   shk,
	}
}
