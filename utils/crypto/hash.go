package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"io"
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

// form MiraiGo

func ComputeMd5AndLength(r io.ReadSeeker) ([]byte, uint32) {
	_, _ = r.Seek(0, io.SeekStart)
	defer func(r io.ReadSeeker, offset int64, whence int) {
		_, _ = r.Seek(offset, whence)
	}(r, 0, io.SeekStart)
	h := md5.New()
	length, _ := io.Copy(h, r)
	return h.Sum(nil), uint32(length)
}

func ComputeSha1AndLength(r io.ReadSeeker) ([]byte, uint32) {
	_, _ = r.Seek(0, io.SeekStart)
	defer func(r io.ReadSeeker, offset int64, whence int) {
		_, _ = r.Seek(offset, whence)
	}(r, 0, io.SeekStart)
	h := sha1.New()
	length, _ := io.Copy(h, r)
	return h.Sum(nil), uint32(length)
}

func ComputeMd5AndSha1AndLength(r io.ReadSeeker) ([]byte, []byte, uint64) {
	defer func(r io.ReadSeeker, offset int64, whence int) {
		_, _ = r.Seek(offset, whence)
	}(r, 0, io.SeekStart)
	md5r := md5.New()
	sha1r := sha1.New()
	_, _ = r.Seek(0, io.SeekStart)
	length, _ := io.Copy(md5r, r)
	_, _ = r.Seek(0, io.SeekStart)
	_, _ = io.Copy(sha1r, r)
	return md5r.Sum(nil), sha1r.Sum(nil), uint64(length)
}
