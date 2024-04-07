package ecdh

import (
	"fmt"
	"testing"
)

func TestKeyExchange(t *testing.T) {
	p256 := ECDH["prime256v1"]
	s192 := ECDH["secp192k1"]

	fmt.Printf("%x\n", p256.GetShareKey())
	fmt.Printf("%x\n", s192.GetShareKey())

	//alice := &BaseECDH{
	//	provider:    newECDHProvider(CURVE["prime256v1"]),
	//	publicKey:   nil,
	//	shareKey:    nil,
	//	compressKey: false,
	//}
	//alice.publicKey = alice.provider.packPublic(false)
	//bob := &BaseECDH{
	//	provider:    newECDHProvider(CURVE["prime256v1"]),
	//	publicKey:   nil,
	//	shareKey:    nil,
	//	compressKey: false,
	//}
	//bob.publicKey = bob.provider.packPublic(false)
	//alice.shareKey = alice.provider.keyExchange(bob.publicKey, false)
	//bob.shareKey = bob.provider.keyExchange(alice.publicKey, false)
	//
	//fmt.Printf("%d\n", alice.provider.secret)
	//fmt.Printf("%d\n", bob.provider.secret)
	//
	//if string(alice.shareKey) != string(bob.shareKey) {
	//	fmt.Printf("alice.shareKey: %x\n", alice.shareKey)
	//	fmt.Printf("bob.shareKey: %x\n", bob.shareKey)
	//	t.Errorf("alice.shareKey != bob.shareKey")
	//} else {
	//	fmt.Printf("alice.shareKey: %x\n", alice.shareKey)
	//	fmt.Printf("bob.shareKey: %x\n", bob.shareKey)
	//	fmt.Println("alice.shareKey == bob.shareKey")
	//}
}
