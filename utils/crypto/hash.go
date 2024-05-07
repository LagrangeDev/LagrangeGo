package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
)

func MD5Digest(v []byte) []byte {
	h := md5.New()
	h.Write(v)
	return h.Sum(make([]byte, 0, md5.Size))
}

func SHA256Digest(v []byte) []byte {
	h := sha256.New()
	h.Write(v)
	return h.Sum(make([]byte, 0, sha256.Size))
}

func SHA1Digest(v []byte) []byte {
	h := sha1.New()
	h.Write(v)
	return h.Sum(make([]byte, 0, sha1.Size))
}
