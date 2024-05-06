package utils

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

type PacketBuilder struct {
	buffer     []byte
	encryptKey []byte
}

func NewPacketBuilder(encryptKey []byte) *PacketBuilder {
	return &PacketBuilder{
		buffer:     make([]byte, 0, 64),
		encryptKey: encryptKey,
	}
}

// WriteByte就是i8

// WriteBytes 重写 默认参数 prefix = "", withPerfix = true
func (b *PacketBuilder) WriteBytes(v []byte, prefix string, withPerfix bool) *PacketBuilder {
	if withPerfix {
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

func (b *PacketBuilder) WriteString(s, prefix string, withPerfix bool) *PacketBuilder {
	return b.WriteBytes([]byte(s), prefix, withPerfix)

}

func (b *PacketBuilder) Len() int {
	return len(b.buffer)
}

func (b *PacketBuilder) append(v []byte) {
	b.buffer = append(b.buffer, v...)
}

func (b *PacketBuilder) Buffer() []byte {
	return b.buffer
}

func (b *PacketBuilder) Data() []byte {
	if b.encryptKey != nil {
		return crypto.QQTeaEncrypt(b.buffer, b.encryptKey)
	}
	return b.buffer
}

func (b *PacketBuilder) pack(v any) *PacketBuilder {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, v)
	b.append(buf.Bytes())
	return b
}

func (b *PacketBuilder) Pack(typ int) []byte {
	if typ != -1 {
		// 或许这里是tlv
		buf := make([]byte, 0)
		buf = binary.BigEndian.AppendUint16(buf, uint16(typ))           // type
		buf = binary.BigEndian.AppendUint16(buf, uint16(len(b.Data()))) // length
		return append(buf, b.Data()...)                                 // type + length + value
	}
	return b.Data()
}

func (b *PacketBuilder) WriteBool(v bool) *PacketBuilder {
	if v {
		b.buffer = append(b.buffer, 1)
	} else {
		b.buffer = append(b.buffer, 0)
	}

	return b
}

func (b *PacketBuilder) WriteStruct(datas ...any) *PacketBuilder {
	for _, data := range datas {
		b.pack(data)
	}
	return b
}

func (b *PacketBuilder) WriteU8(v uint8) *PacketBuilder {
	b.buffer = append(b.buffer, v)
	return b
}

func (b *PacketBuilder) WriteU16(v uint16) *PacketBuilder {
	b.buffer = binary.BigEndian.AppendUint16(b.buffer, v)
	return b
}

func (b *PacketBuilder) WriteU32(v uint32) *PacketBuilder {
	b.buffer = binary.BigEndian.AppendUint32(b.buffer, v)
	return b
}

func (b *PacketBuilder) WriteU64(v uint64) *PacketBuilder {
	b.buffer = binary.BigEndian.AppendUint64(b.buffer, v)
	return b
}

func (b *PacketBuilder) WriteI8(v int8) *PacketBuilder {
	return b.WriteU8(byte(v))
}

func (b *PacketBuilder) WriteI16(v int16) *PacketBuilder {
	return b.WriteU16(uint16(v))
}

func (b *PacketBuilder) WriteI32(v int32) *PacketBuilder {
	return b.WriteU32(uint32(v))
}

func (b *PacketBuilder) WriteI64(v int64) *PacketBuilder {
	return b.WriteU64(uint64(v))
}

func (b *PacketBuilder) WriteFloat(v float32) *PacketBuilder {
	return b.WriteU32(math.Float32bits(v))
}
func (b *PacketBuilder) WriteDouble(v float64) *PacketBuilder {
	return b.WriteU64(math.Float64bits(v))
}
func (b *PacketBuilder) WriteTlv(tlvs [][]byte) *PacketBuilder {
	b.WriteU16(uint16(len(tlvs)))
	for _, tlv := range tlvs {
		b.WriteBytes(tlv, "", true)
	}
	return b
}
