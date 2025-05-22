package binary

import (
	"encoding/binary"
	"io"
	"strconv"
	"unsafe"

	"github.com/LagrangeDev/LagrangeGo/utils"
)

type Reader struct {
	reader io.Reader
	buffer []byte
	pos    int
}

func ParseReader(reader io.Reader) *Reader {
	return &Reader{
		reader: reader,
	}
}

func NewReader(buffer []byte) *Reader {
	return &Reader{
		buffer: buffer,
	}
}

func (r *Reader) Len() int {
	if r.reader != nil {
		return -1
	}
	return len(r.buffer) - r.pos
}

func (r *Reader) ReadByte() (byte, error) {
	if r.pos >= len(r.buffer) {
		return 0, io.EOF
	}
	b := r.buffer[r.pos]
	r.pos++
	return b, nil
}

// String means read all available data and return them as a string
//
// if r.reader got error, it will returns as err.Error()
func (r *Reader) String() string {
	if r.reader != nil {
		data, err := io.ReadAll(r.reader)
		if err != nil {
			return err.Error()
		}
		return utils.B2S(data)
	}
	s := string(r.buffer[r.pos:])
	r.pos = 0
	r.buffer = r.buffer[:0]
	return s
}

// ReadAll means read all available data and return them
//
// if r.reader got error, it will return nil
func (r *Reader) ReadAll() []byte {
	if r.reader != nil {
		data, err := io.ReadAll(r.reader)
		if err != nil {
			return nil
		}
		return data
	}
	s := r.buffer[r.pos:]
	r.pos = 0
	r.buffer = r.buffer[:0]
	buf := make([]byte, len(s))
	copy(buf, s)
	return s
}

func (r *Reader) ReadU8() (v uint8) {
	if r.reader != nil {
		buf := make([]byte, 1)
		n, err := r.reader.Read(buf)
		if err != nil || n < 1 {
			// 读取失败或读取的数据不足，返回零值
			return 0
		}
		v = buf[0]
		return
	}
	// 确保缓冲区有足够的数据
	if r.pos >= len(r.buffer) {
		return 0
	}
	v = r.buffer[r.pos]
	r.pos++
	return
}

func readint[T ~uint16 | ~uint32 | ~uint64](r *Reader) (v T) {
	sz := unsafe.Sizeof(v)
	buf := make([]byte, 8)
	if r.reader != nil {
		n, err := r.reader.Read(buf[8-sz:])
		if err != nil || n < int(sz) {
			// 读取失败或读取的数据不足，返回零值
			return 0
		}
	} else {
		// 确保缓冲区有足够的数据
		if r.pos+int(sz) > len(r.buffer) {
			return 0
		}
		copy(buf[8-sz:], r.buffer[r.pos:r.pos+int(sz)])
		r.pos += int(sz)
	}
	v = (T)(binary.BigEndian.Uint64(buf))
	return
}

func (r *Reader) ReadU16() (v uint16) {
	return readint[uint16](r)
}

func (r *Reader) ReadU32() (v uint32) {
	return readint[uint32](r)
}

func (r *Reader) ReadU64() (v uint64) {
	return readint[uint64](r)
}

func (r *Reader) SkipBytes(length int) {
	if r.reader != nil {
		_, _ = r.reader.Read(make([]byte, length))
		return
	}
	r.pos += length
}

// ReadBytesNoCopy 不拷贝读取的数据, 用于读取后立即使用, 慎用
//
// 如需使用, 请确保 Reader 未被回收
func (r *Reader) ReadBytesNoCopy(length int) (v []byte) {
	if r.reader != nil {
		return r.ReadBytes(length)
	}
	// 确保缓冲区有足够的数据
	if r.pos+length > len(r.buffer) {
		return make([]byte, 0)
	}
	v = r.buffer[r.pos : r.pos+length]
	r.pos += length
	return
}

func (r *Reader) ReadBytes(length int) (v []byte) {
	// 返回一个全新的数组罢
	v = make([]byte, length)
	if r.reader != nil {
		n, err := io.ReadFull(r.reader, v)
		if err != nil || n < length {
			// 读取失败或读取的数据不足，返回空数组
			return make([]byte, 0)
		}
	} else {
		// 确保缓冲区有足够的数据
		if r.pos+length > len(r.buffer) {
			return make([]byte, 0)
		}
		copy(v, r.buffer[r.pos:r.pos+length])
		r.pos += length
	}
	return
}

func (r *Reader) ReadString(length int) string {
	return utils.B2S(r.ReadBytes(length))
}

func (r *Reader) SkipBytesWithLength(prefix string, withPerfix bool) {
	var length int
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
	if withPerfix {
		plus, err := strconv.Atoi(prefix[1:])
		if err != nil {
			panic(err)
		}
		length -= plus / 8
	}
	r.SkipBytes(length)
}

func (r *Reader) ReadBytesWithLength(prefix string, withPerfix bool) []byte {
	var length int
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
	if withPerfix {
		plus, err := strconv.Atoi(prefix[1:])
		if err != nil {
			panic(err)
		}
		length -= plus / 8
	}
	return r.ReadBytes(length)
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

func (r *Reader) ReadVarint() (int64, error) {
	return binary.ReadVarint(r)
}

func (r *Reader) ReadUvarint() (uint64, error) {
	return binary.ReadUvarint(r)
}
