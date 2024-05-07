package binary

import (
	"encoding/binary"

	"github.com/LagrangeDev/LagrangeGo/utils"
)

type Reader struct {
	buffer []byte
	pos    int
}

func NewReader(buffer []byte) *Reader {
	return &Reader{
		buffer: buffer,
		pos:    0,
	}
}

func (r *Reader) GetRamin() int {
	return len(r.buffer) - r.pos
}

func (r *Reader) ReadU8() (v uint8) {
	v = r.buffer[r.pos]
	r.pos++
	return
}

func (r *Reader) ReadU16() (v uint16) {
	v = binary.BigEndian.Uint16(r.buffer[r.pos : r.pos+2])
	r.pos += 2
	return
}

func (r *Reader) ReadU32() (v uint32) {
	v = binary.BigEndian.Uint32(r.buffer[r.pos : r.pos+4])
	r.pos += 4
	return
}
func (r *Reader) ReadU64() (v uint64) {
	v = binary.BigEndian.Uint64(r.buffer[r.pos : r.pos+8])
	r.pos += 8
	return
}

func (r *Reader) SkipBytes(length int) {
	r.pos += length
}

// ReadBytesNoCopy 不拷贝读取的数据, 用于读取后立即使用, 慎用
//
// 如需使用, 请确保 Reader 未被回收
func (r *Reader) ReadBytesNoCopy(length int) (v []byte) {
	v = r.buffer[r.pos : r.pos+length]
	r.pos += length
	return
}

func (r *Reader) ReadBytes(length int) (v []byte) {
	// 返回一个全新的数组罢
	v = make([]byte, length)
	copy(v, r.buffer[r.pos:r.pos+length])
	r.pos += length
	return
}

func (r *Reader) ReadString(length int) string {
	return utils.B2S(r.ReadBytes(length))
}

func (r *Reader) SkipBytesWithLength(prefix string, withPerfix bool) {
	var length int
	if withPerfix {
		switch prefix {
		case "u8":
			length = int(r.ReadU8() - 1)
		case "u16":
			length = int(r.ReadU16() - 2)
		case "u32":
			length = int(r.ReadU32() - 4)
		case "u64":
			length = int(r.ReadU64() - 8)
		default:
			panic("invaild prefix")
		}
	} else {
		switch prefix {
		case "u8":
			length = int(r.ReadU8())
		case "u16":
			length = int(r.ReadU16())
		case "u32":
			length = int(r.ReadU32())
		case "u64":
			length = int(r.ReadU64())
		default:
			panic("invaild prefix")
		}
	}
	r.pos += length
}

func (r *Reader) ReadBytesWithLength(prefix string, withPerfix bool) (v []byte) {
	var length int
	if withPerfix {
		switch prefix {
		case "u8":
			length = int(r.ReadU8() - 1)
		case "u16":
			length = int(r.ReadU16() - 2)
		case "u32":
			length = int(r.ReadU32() - 4)
		case "u64":
			length = int(r.ReadU64() - 8)
		default:
			panic("invaild prefix")
		}
	} else {
		switch prefix {
		case "u8":
			length = int(r.ReadU8())
		case "u16":
			length = int(r.ReadU16())
		case "u32":
			length = int(r.ReadU32())
		case "u64":
			length = int(r.ReadU64())
		default:
			panic("invaild prefix")
		}
	}
	v = make([]byte, length)
	copy(v, r.buffer[r.pos:r.pos+length])
	r.pos += length
	return
}

func (r *Reader) ReadStringWithLength(prefix string, withPerfix bool) string {
	return utils.B2S(r.ReadBytesWithLength(prefix, withPerfix))
}

func (r *Reader) ReadTlv() (result map[uint16][]byte) {
	result = map[uint16][]byte{}
	count := r.ReadU16()

	for i := 0; i < int(count); i++ {
		tag := r.ReadU16()
		result[tag] = r.ReadBytes(int(r.ReadU16()))
	}
	return
}

// go的残废泛型
//func (r *Reader) ReadStruct(){
//
//}

func (r *Reader) ReadI8() (v int8) {
	return int8(r.ReadU8())
}

func (r *Reader) ReadI16() (v int16) {
	return int16(r.ReadU16())
}

func (r *Reader) ReadI32() (v int32) {
	return int32(r.ReadU32())
}

func (r *Reader) ReadI64() (v int64) {
	return int64(r.ReadU64())
}
