package binary

// from https://github.com/Mrs4s/MiraiGo/blob/master/binary/pool.go

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"sync"
	"unsafe"

	"github.com/fumiama/orbyte"
	"github.com/fumiama/orbyte/pbuf"
)

var bufferPool = pbuf.NewBufferPool[teacfg]()

type Builder orbyte.Item[pbuf.UserBuffer[teacfg]]

func init() {
	x := *(**orbyte.Pool[pbuf.UserBuffer[teacfg]])(unsafe.Pointer(&bufferPool))
	x.SetManualDestroy(true)
	x.SetNoPutBack(true)
}

// NewBuilder 从池中取出一个 Builder
func NewBuilder(key ...byte) *Builder {
	b := bufferPool.NewBuffer(nil)
	b.P(func(ub *pbuf.UserBuffer[teacfg]) {
		ub.DAT.init(key)
	})
	return (*Builder)(b)
}

var gzipPool = sync.Pool{
	New: func() any {
		buf := new(bytes.Buffer)
		w := gzip.NewWriter(buf)
		return &GzipWriter{
			w:   w,
			buf: buf,
		}
	},
}

func acquireGzipWriter() *GzipWriter {
	ret := gzipPool.Get().(*GzipWriter)
	ret.buf.Reset()
	ret.w.Reset(ret.buf)
	return ret
}

func releaseGzipWriter(w *GzipWriter) {
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16
	if w.buf.Cap() < maxSize {
		w.buf.Reset()
		gzipPool.Put(w)
	}
}

type zlibWriter struct {
	w   *zlib.Writer
	buf *bytes.Buffer
}

var zlibPool = sync.Pool{
	New: func() any {
		buf := new(bytes.Buffer)
		w := zlib.NewWriter(buf)
		return &zlibWriter{
			w:   w,
			buf: buf,
		}
	},
}

func acquireZlibWriter() *zlibWriter {
	ret := zlibPool.Get().(*zlibWriter)
	ret.buf.Reset()
	ret.w.Reset(ret.buf)
	return ret
}

func releaseZlibWriter(w *zlibWriter) {
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16
	if w.buf.Cap() < maxSize {
		w.buf.Reset()
		zlibPool.Put(w)
	}
}
