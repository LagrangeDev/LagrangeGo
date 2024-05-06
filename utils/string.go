package utils

// from https://github.com/Mrs4s/MiraiGo/blob/master/utils/string.go

import "unsafe"

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
