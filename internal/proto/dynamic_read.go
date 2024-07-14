package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

const (
	WireVarint     = 0
	WireFixed64    = 1
	WireBytes      = 2
	WireStartGroup = 3
	WireEndGroup   = 4
	WireFixed32    = 5
	WireSFixed32   = 5 | (1 << 3)
	WireSFixed64   = 1 | (1 << 3)
)

type TypeValue struct {
	Type  reflect.Type
	Vaule any
}

func readVarint(r io.Reader) (int64, error) {
	var result int64
	var shift uint
	for {
		var b byte
		err := binary.Read(r, binary.LittleEndian, &b)
		if err != nil {
			return 0, err
		}
		result |= int64(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return result, nil
}

func readFixed64(r io.Reader) (uint64, error) {
	var result uint64
	err := binary.Read(r, binary.LittleEndian, &result)
	return result, err
}

// 从字节数组读取一个定长32位整数
func readFixed32(r io.Reader) (uint32, error) {
	var result uint32
	err := binary.Read(r, binary.LittleEndian, &result)
	return result, err
}

func readSFixed32(r io.Reader) (int32, error) {
	var result int32
	err := binary.Read(r, binary.LittleEndian, &result)
	return result, err
}

// 从字节数组读取有符号定长64位整数
func readSFixed64(r io.Reader) (int64, error) {
	var result int64
	err := binary.Read(r, binary.LittleEndian, &result)
	return result, err
}

// 从字节数组读取字节切片
func readBytes(r io.Reader) ([]byte, error) {
	length, err := readVarint(r)
	if err != nil {
		return nil, err
	}
	result := make([]byte, length)
	_, err = io.ReadFull(r, result)
	return result, err
}

func ReadField(targetField int64, data []byte) ([]*TypeValue, error) {
	r := bytes.NewReader(data)
	result := make([]*TypeValue, 0, 2)
	for r.Len() > 0 {
		key, err := readVarint(r)
		if err != nil {
			return nil, err
		}
		field := key >> 3
		wireType := key & 0x7
		var value any
		switch wireType {
		case WireVarint:
			value, err = readVarint(r)
		case WireFixed64:
			value, err = readFixed64(r)
		case WireBytes:
			value, err = readBytes(r)
		case WireFixed32:
			value, err = readFixed32(r)
		default:
			return nil, fmt.Errorf("unsupported wire type: %d", wireType)
		}
		if field == targetField {
			result = append(result, &TypeValue{
				Type:  reflect.TypeOf(value),
				Vaule: value,
			})
		}
	}
	return result, nil
}
