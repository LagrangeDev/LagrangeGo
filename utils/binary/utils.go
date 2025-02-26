package binary

// from https://github.com/Mrs4s/MiraiGo/blob/master/binary/utils.go

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"net"
)

type GzipWriter struct {
	w   *gzip.Writer
	buf *bytes.Buffer
}

func (w *GzipWriter) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w *GzipWriter) Close() error {
	return w.w.Close()
}

func (w *GzipWriter) Bytes() []byte {
	return w.buf.Bytes()
}

func ZlibUncompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, _ := zlib.NewReader(b)
	defer r.Close()
	_, _ = out.ReadFrom(r)
	return out.Bytes()
}

func ZlibCompress(data []byte) []byte {
	zw := acquireZlibWriter()
	_, _ = zw.w.Write(data)
	_ = zw.w.Close()
	ret := make([]byte, len(zw.buf.Bytes()))
	copy(ret, zw.buf.Bytes())
	releaseZlibWriter(zw)
	return ret
}

func GZipCompress(data []byte) []byte {
	gw := acquireGzipWriter()
	_, _ = gw.Write(data)
	_ = gw.Close()
	ret := make([]byte, len(gw.buf.Bytes()))
	copy(ret, gw.buf.Bytes())
	releaseGzipWriter(gw)
	return ret
}

func GZipUncompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, _ := gzip.NewReader(b)
	defer r.Close()
	_, _ = out.ReadFrom(r)
	return out.Bytes()
}

func UInt32ToIPV4Address(i uint32) string {
	ip := net.IP{0, 0, 0, 0}
	binary.LittleEndian.PutUint32(ip, i)
	return ip.String()
}
