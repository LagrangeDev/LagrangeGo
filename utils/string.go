package utils

// from https://github.com/Mrs4s/MiraiGo/blob/master/utils/string.go

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"unsafe"
)

// B2S converts byte slice to a string without memory allocation.
func B2S(b []byte) string {
	size := len(b)
	if size == 0 {
		return ""
	}
	return unsafe.String(&b[0], size)
}

// S2B converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func S2B(s string) (b []byte) {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func MustParseHexStr(s string) []byte {
	result, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return result
}

func NewTrace() string {
	randomBytes := make([]byte, 16+8)

	if _, err := rand.Read(randomBytes); err != nil {
		return ""
	}

	trace := fmt.Sprintf("00-%x-%x-01", randomBytes[:16], randomBytes[16:])
	return trace
}
