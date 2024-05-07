package binary

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"

	ftea "github.com/fumiama/gofastTEA"

	"github.com/LagrangeDev/LagrangeGo/utils"
)

type Builder struct {
	buffer []byte
	key    ftea.TEA
	usetea bool
}

func NewBuilder(key []byte) *Builder {
	return &Builder{
		buffer: make([]byte, 0, 64),
		key:    ftea.NewTeaCipher(key),
		usetea: len(key) == 16,
	}
}

func (b *Builder) Len() int {
	return len(b.buffer)
}

func (b *Builder) append(v []byte) {
	b.buffer = append(b.buffer, v...)
}

// data 带 tea 加密, 不可用作提取 b.buffer
func (b *Builder) data() []byte {
	if b.usetea {
		return b.key.Encrypt(b.buffer)
	}
	return b.buffer
}

func (b *Builder) pack(v any) {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, v)
	b.append(buf.Bytes())
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
		n = b.key.EncryptTo(b.buffer, buf[2+2:])
	} else {
		n = copy(buf[2+2:], b.buffer)
	}

	binary.BigEndian.PutUint16(buf[0:2], typ)         // type
	binary.BigEndian.PutUint16(buf[2:2+2], uint16(n)) // length

	return buf[:n+2+2]
}

func (b *Builder) WriteBool(v bool) *Builder {
	if v {
		b.buffer = append(b.buffer, 1)
	} else {
		b.buffer = append(b.buffer, 0)
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
		panic("Invaild prefix")
	}
	b.append(v)
	return b
}

func (b *Builder) WritePacketString(s, prefix string, withPrefix bool) *Builder {
	return b.WritePacketBytes(utils.S2B(s), prefix, withPrefix)
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
		b.pack(data)
	}
	return b
}

func (b *Builder) WriteU8(v uint8) *Builder {
	b.buffer = append(b.buffer, v)
	return b
}

func (b *Builder) WriteU16(v uint16) *Builder {
	b.buffer = binary.BigEndian.AppendUint16(b.buffer, v)
	return b
}

func (b *Builder) WriteU32(v uint32) *Builder {
	b.buffer = binary.BigEndian.AppendUint32(b.buffer, v)
	return b
}

func (b *Builder) WriteU64(v uint64) *Builder {
	b.buffer = binary.BigEndian.AppendUint64(b.buffer, v)
	return b
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
