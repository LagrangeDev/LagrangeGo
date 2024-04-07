package binary

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/Redmomn/LagrangeGo/utils/crypto"
)

type Builder struct {
	buffer     []byte
	encryptKey []byte
}

func NewBuilder(encryptKey []byte) *Builder {
	return &Builder{
		buffer:     make([]byte, 0),
		encryptKey: encryptKey,
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
	if b.encryptKey != nil {
		return crypto.QQTeaEncrypt(b.buffer, b.encryptKey)
	}
	return b.buffer
}

func (b *Builder) pack(v any) *Builder {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, v)
	b.append(buf.Bytes())
	return b
}

func (b *Builder) Pack(typ int) []byte {
	if typ != -1 {
		// 或许这里是tlv
		buf := new(bytes.Buffer)
		_ = binary.Write(buf, binary.BigEndian, typ)           // type
		_ = binary.Write(buf, binary.BigEndian, len(b.Data())) // length
		return append(buf.Bytes(), b.Data()...)                // type + length + value
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

func (b *Builder) WriteByte(v byte) *Builder {
	b.buffer = append(b.buffer, v)
	return b.pack(v)
}

func (b *Builder) WriteBytes(v []byte, withLength bool) *Builder {
	if withLength {
		b.WriteU16(uint16(len(v)))
	}
	b.append(v)
	return b
}

func (b *Builder) WriteString(v string) *Builder {
	return b.WriteBytes([]byte(v), true)
}

func (b *Builder) WriteStruct(datas ...any) *Builder {
	for _, data := range datas {
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
func (b *Builder) WriteTlv(tlvs [][]byte) *Builder {
	b.WriteU16(uint16(len(tlvs)))
	for _, tlv := range tlvs {
		b.WriteBytes(tlv, false)
	}
	return b
}
