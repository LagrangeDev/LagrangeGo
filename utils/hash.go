package utils

import (
	"crypto/md5"
	"crypto/sha256"
)

func Md5Digest(v []byte) []byte {
	hasher := md5.New()
	hasher.Write(v)
	return hasher.Sum(nil)
}

func Sha256Digest(v []byte) []byte {
	sha256er := sha256.New()
	sha256er.Write(v)
	return sha256er.Sum(nil)
}
