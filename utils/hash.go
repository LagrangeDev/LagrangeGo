package utils

import (
	"crypto/md5"
	"crypto/sha1"
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

func Sha1Digest(v []byte) []byte {
	sha1er := sha1.New()
	sha1er.Write(v)
	return sha1er.Sum(nil)
}
