package binary

import (
	"encoding/binary"
	"io"
	"math"
	"net"
	"strconv"
	"unsafe"

	ftea "github.com/fumiama/gofastTEA"
	"github.com/fumiama/orbyte"
	"github.com/fumiama/orbyte/pbuf"

	"github.com/LagrangeDev/LagrangeGo/utils"
)

type teacfg struct {
	key    ftea.TEA
	usetea bool
}

func (tc *teacfg) init(key []byte) {
	tc.key = ftea.NewTeaCipher(key)
	tc.usetea = len(key) == 16
}

func (b *Builder) p(f func(*pbuf.UserBuffer[teacfg])) {
	(*orbyte.Item[pbuf.UserBuffer[teacfg]])(b).P(f)
}

// ToBytes return data with tea encryption if key is set
//
// GC 安全, 返回的数据在 Builder 被销毁之后仍能被正确读取,
// 但是只能调用一次, 调用后 Builder 即失效
func (b *Builder) ToBytes() (data []byte) {
	b.p(func(ub *pbuf.UserBuffer[teacfg]) {
		if ub.DAT.usetea {
			data = ub.DAT.key.Encrypt(ub.Bytes())
			return
		}
		data = make([]byte, ub.Len())
		copy(data, ub.Bytes())
	})
	return
}

// Pack TLV with tea encryption if key is set
//
// GC 安全, 返回的数据在 Builder 被销毁之后仍能被正确读取,
// 调用后 Builder 仍有效
func (b *Builder) Pack(typ uint16) (data []byte) {
	b.p(func(buf *pbuf.UserBuffer[teacfg]) {
		data = make([]byte, buf.Len()+2+2+16)

		n := 0
		if buf.DAT.usetea {
			n = buf.DAT.key.EncryptTo(buf.Bytes(), data[2+2:])
		} else {
			n = copy(data[2+2:], buf.Bytes())
		}

		binary.BigEndian.PutUint16(data[0:2], typ)         // type
		binary.BigEndian.PutUint16(data[2:2+2], uint16(n)) // length

		data = data[:n+2+2]
	})
	return
}

func (b *Builder) WriteBool(v bool) *Builder {
	if v {
		b.WriteU8('1')
	} else {
		b.WriteU8('0')
	}
	return b
}

// WritePacketBytes prefix must not be empty
func (b *Builder) WritePacketBytes(v []byte, prefix string, withPrefix bool) *Builder {
	n := len(v)
	if withPrefix {
		plus, err := strconv.Atoi(prefix[1:])
		if err != nil {
			panic(err)
		}
		n += plus / 8
	}
	switch prefix {
	case "u8":
		b.WriteU8(uint8(n))
	case "u16":
		b.WriteU16(uint16(n))
	case "u32":
		b.WriteU32(uint32(n))
	case "u64":
		b.WriteU64(uint64(n))
	default:
		panic("invaild prefix")
	}
	b.WriteBytes(v)
	return b
}

func (b *Builder) WritePacketString(s, prefix string, withPrefix bool) *Builder {
	return b.WritePacketBytes(utils.S2B(s), prefix, withPrefix)
}

// Write for impl. io.Writer
func (b *Builder) Write(p []byte) (n int, err error) {
	b.p(func(ub *pbuf.UserBuffer[teacfg]) {
		n, err = ub.Write(p)
	})
	return
}

func (b *Builder) EncryptAndWrite(key []byte, data []byte) *Builder {
	_, _ = b.Write(ftea.NewTeaCipher(key).Encrypt(data))
	return b
}

// ReadFrom for impl. io.ReaderFrom
func (b *Builder) ReadFrom(r io.Reader) (n int64, err error) {
	b.p(func(ub *pbuf.UserBuffer[teacfg]) {
		n, err = io.Copy(ub, r)
	})
	return
}

func (b *Builder) WriteLenBytes(v []byte) *Builder {
	b.WriteU16(uint16(len(v)))
	b.WriteBytes(v)
	return b
}

func (b *Builder) WriteBytes(v []byte) *Builder {
	_, _ = b.Write(v)
	return b
}

func (b *Builder) WriteLenString(v string) *Builder {
	return b.WriteLenBytes(utils.S2B(v))
}

func (b *Builder) WriteStruct(data ...any) *Builder {
	b.p(func(ub *pbuf.UserBuffer[teacfg]) {
		for _, data := range data {
			_ = binary.Write(ub, binary.BigEndian, data)
		}
	})
	return b
}

func (b *Builder) WriteU8(v uint8) *Builder {
	b.p(func(ub *pbuf.UserBuffer[teacfg]) {
		ub.WriteByte(v)
	})
	return b
}

func writeint[T ~uint16 | ~uint32 | ~uint64](b *Builder, v T) *Builder {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	_, _ = b.Write(buf[8-unsafe.Sizeof(v):])
	return b
}

func (b *Builder) WriteU16(v uint16) *Builder {
	return writeint(b, v)
}

func (b *Builder) WriteU32(v uint32) *Builder {
	return writeint(b, v)
}

func (b *Builder) WriteU64(v uint64) *Builder {
	return writeint(b, v)
}

func (b *Builder) WriteI8(v int8) *Builder {
	return b.WriteU8(byte(v))
}

func (b *Builder) WriteI16(v int16) *Builder {
	return b.WriteU16(uint16(v))
}

func (b *Builder) WriteI32(v int32) *Builder {
	return b.WriteU32(uint32(v))
}

func (b *Builder) WriteI64(v int64) *Builder {
	return b.WriteU64(uint64(v))
}

func (b *Builder) WriteFloat(v float32) *Builder {
	return b.WriteU32(math.Float32bits(v))
}

func (b *Builder) WriteDouble(v float64) *Builder {
	return b.WriteU64(math.Float64bits(v))
}

func (b *Builder) WriteTLV(tlvs ...[]byte) *Builder {
	b.WriteU16(uint16(len(tlvs)))
	_, _ = io.Copy(b, (*net.Buffers)(&tlvs))
	return b
}
