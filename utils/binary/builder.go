package binary

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

type PackType int

const PackTypeNone PackType = -1

type Builder struct {
	buffer []byte
	key    crypto.TEA
	usetea bool
}

func NewBuilder(key []byte) *Builder {
	return &Builder{
		buffer: make([]byte, 0, 64),
		key:    crypto.NewTeaCipher(key),
		usetea: len(key) == 16,
	}
}

func (b *Builder) Len() int {
	return len(b.buffer)
}

func (b *Builder) append(v []byte) {
	b.buffer = append(b.buffer, v...)
}

func (b *Builder) Buffer() []byte {
	return b.buffer
}

func (b *Builder) Data() []byte {
	if b.usetea {
		return b.key.Encrypt(b.buffer)
	}
	return b.buffer
}

func (b *Builder) pack(v any) *Builder {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, v)
	b.append(buf.Bytes())
	return b
}

func (b *Builder) Pack(typ PackType) []byte {
	if typ != PackTypeNone {
		// 或许这里是tlv
		buf := make([]byte, b.Len()+4)
		binary.BigEndian.PutUint16(buf[0:2], uint16(typ))           // type
		binary.BigEndian.PutUint16(buf[2:4], uint16(len(b.Data()))) // length
		copy(buf[4:], b.Data())                                     // type + length + value
		return buf
	}
	return b.Data()
}

func (b *Builder) WriteBool(v bool) *Builder {
	if v {
		b.buffer = append(b.buffer, 1)
	} else {
		b.buffer = append(b.buffer, 0)
	}

	return b
}

// WritePacketBytes default prefix = "", withPrefix = true
func (b *Builder) WritePacketBytes(v []byte, prefix string, withPrefix bool) *Builder {
	if withPrefix {
		switch prefix {
		case "":
		case "u8":
			b.WriteU8(uint8(len(v) + 1))
		case "u16":
			b.WriteU16(uint16(len(v) + 2))
		case "u32":
			b.WriteU32(uint32(len(v) + 4))
		case "u64":
			b.WriteU64(uint64(len(v) + 8))
		default:
			panic("Invaild prefix")
		}
	} else {
		switch prefix {
		case "":
		case "u8":
			b.WriteU8(uint8(len(v)))
		case "u16":
			b.WriteU16(uint16(len(v)))
		case "u32":
			b.WriteU32(uint32(len(v)))
		case "u64":
			b.WriteU64(uint64(len(v)))
		default:
			panic("Invaild prefix")
		}
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

/*
func (b *Builder) WriteTlv(tlvs ...[]byte) *Builder {
	b.WriteU16(uint16(len(tlvs)))
	for _, tlv := range tlvs {
		b.WriteBytes(tlv, false)
	}
	return b
}
*/

func (b *Builder) WritePacketTlv(tlvs ...[]byte) *Builder {
	b.WriteU16(uint16(len(tlvs)))
	for _, tlv := range tlvs {
		b.WritePacketBytes(tlv, "", true)
	}
	return b
}
