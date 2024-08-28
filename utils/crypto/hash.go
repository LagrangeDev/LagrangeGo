package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"io"
	"reflect"
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

func ComputeMd5AndLengthWithLimit(r io.ReadSeeker, limit int64) ([]byte, uint32) {
	_, _ = r.Seek(0, io.SeekStart)
	defer func(r io.ReadSeeker, offset int64, whence int) {
		_, _ = r.Seek(offset, whence)
	}(r, 0, io.SeekStart)
	h := md5.New()
	lr := io.LimitedReader{R: r, N: limit}
	length, _ := io.Copy(h, &lr)
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
	_, _ = r.Seek(0, io.SeekStart)
	md5r := md5.New()
	sha1r := sha1.New()
	mr := io.MultiWriter(md5r, sha1r)
	length, _ := io.Copy(mr, r)
	return md5r.Sum(nil), sha1r.Sum(nil), uint64(length)
}

func ComputeBlockSha1(r io.ReadSeeker, blockSize int) [][]byte {
	_, _ = r.Seek(0, io.SeekStart)
	defer func(r io.ReadSeeker, offset int64, whence int) {
		_, _ = r.Seek(offset, whence)
	}(r, 0, io.SeekStart)
	result := make([][]byte, 0, 5)
	sha1r := sha1.New()
	buffer := make([]byte, blockSize)
	for {
		n, _ := io.ReadFull(r, buffer)
		if n != blockSize {
			sha1r.Write(buffer[:n])
			break
		}
		sha1r.Write(buffer)
		result = append(result, GetSha1Status(sha1r))
	}
	result = append(result, sha1r.Sum(nil))
	return result
}

func GetSha1Status(h hash.Hash) []byte {
	if reflect.TypeOf(h).String() != "*sha1.digest" {
		return nil
	}
	result := make([]byte, 20)
	// 反射大法好
	s := reflect.ValueOf(h).Elem().Field(0)
	for i := 0; i < 5; i++ {
		binary.LittleEndian.PutUint32(result[i*4:], uint32(s.Index(i).Uint()))
	}
	return result
}
