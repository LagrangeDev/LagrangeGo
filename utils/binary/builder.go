package binary

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"strconv"
	"unsafe"

	ftea "github.com/fumiama/gofastTEA"

	"github.com/LagrangeDev/LagrangeGo/utils"
)

type Builder struct {
	buffer *bytes.Buffer
	key    ftea.TEA
	usetea bool
	io.Writer
	io.ReaderFrom
}

func NewBuilder(key []byte) *Builder {
	return &Builder{
		buffer: bytes.NewBuffer(make([]byte, 0, 64)),
		key:    ftea.NewTeaCipher(key),
		usetea: len(key) == 16,
	}
}

func (b *Builder) Len() int {
	return b.buffer.Len()
}

func (b *Builder) append(v []byte) {
	b.buffer.Write(v)
}

// data 带 tea 加密, 不可用作提取 b.buffer
func (b *Builder) data() []byte {
	if b.usetea {
		return b.key.Encrypt(b.buffer.Bytes())
	}
	return b.buffer.Bytes()
}

func (b *Builder) pack(v any) error {
	return binary.Write(b.buffer, binary.BigEndian, v)
}

func (b *Builder) ToReader() io.Reader {
	return b.buffer
}

// ToBytes return data with tea encryption
func (b *Builder) ToBytes() []byte {
	return b.data()
}

// Pack TLV with tea encryption
func (b *Builder) Pack(typ uint16) []byte {
	buf := make([]byte, b.Len()+2+2+16)

	n := 0
	if b.usetea {
		n = b.key.EncryptTo(b.buffer.Bytes(), buf[2+2:])
	} else {
		n = copy(buf[2+2:], b.buffer.Bytes())
	}

	binary.BigEndian.PutUint16(buf[0:2], typ)         // type
	binary.BigEndian.PutUint16(buf[2:2+2], uint16(n)) // length

	return buf[:n+2+2]
}

func (b *Builder) WriteBool(v bool) *Builder {
	if v {
		b.buffer.WriteByte('1')
	} else {
		b.buffer.WriteByte('0')
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
	b.append(v)
	return b
}

func (b *Builder) WritePacketString(s, prefix string, withPrefix bool) *Builder {
	return b.WritePacketBytes(utils.S2B(s), prefix, withPrefix)
}

func (b *Builder) Write(p []byte) (n int, err error) {
	return b.buffer.Write(p)
}

func (b *Builder) ReadFrom(r io.Reader) (n int64, err error) {
	return io.Copy(b.buffer, r)
}

func (b *Builder) WriteBytes(v []byte, withLength bool) *Builder {
	if withLength {
		b.WriteU16(uint16(len(v)))
	}
	b.append(v)
	return b
}

func (b *Builder) WriteString(v string) *Builder {
	return b.WriteBytes(utils.S2B(v), true)
}

func (b *Builder) WriteStruct(data ...any) *Builder {
	for _, data := range data {
		err := b.pack(data)
		if err != nil {
			panic(err)
		}
	}
	return b
}

func (b *Builder) WriteU8(v uint8) *Builder {
	b.buffer.WriteByte(v)
	return b
}

func writeint[T ~uint16 | ~uint32 | ~uint64](b *Builder, v T) *Builder {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	b.buffer.Write(buf[8-unsafe.Sizeof(v):])
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

func (b *Builder) WriteTlv(tlvs ...[]byte) *Builder {
	b.WriteU16(uint16(len(tlvs)))
	for _, tlv := range tlvs {
		b.WriteBytes(tlv, false)
	}
	return b
}
