package crypto

import (
	"bytes"
	"encoding/binary"
)

// 使用utils里面的reader会循环引用

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

func (r *Reader) ReadU32() (v uint32) {
	v = binary.BigEndian.Uint32(r.buffer[r.pos : r.pos+4])
	r.pos += 4
	return
}

const (
	DELTA = uint64(0x9E3779B9)
	ROUND = 16
	OP    = uint64(0xFFFFFFFF)
)

func xor(a, b []byte) []byte {
	var a1, a2, b1, b2 uint32
	readera := NewReader(a)
	readerb := NewReader(b)

	a1 = readera.ReadU32()
	a2 = readera.ReadU32()
	b1 = readerb.ReadU32()
	b2 = readerb.ReadU32()
	//_ = binary.Read(readera, binary.BigEndian, &a1)
	//_ = binary.Read(readera, binary.BigEndian, &a2)
	//_ = binary.Read(readerb, binary.BigEndian, &b1)
	//_ = binary.Read(readerb, binary.BigEndian, &b2)

	buf := make([]byte, 0)
	buf = binary.BigEndian.AppendUint32(buf, (a1^b1)&uint32(OP))
	buf = binary.BigEndian.AppendUint32(buf, (a2^b2)&uint32(OP))

	return buf
}

func teaCode(v, k []byte) []byte {
	var key0, key1, key2, key3, val0, val1 uint32
	var k0, k1, k2, k3, v0, v1 uint64

	//reader1 := bytes.NewReader(k)
	//_ = binary.Read(reader1, binary.BigEndian, &key0)
	//_ = binary.Read(reader1, binary.BigEndian, &key1)
	//_ = binary.Read(reader1, binary.BigEndian, &key2)
	//_ = binary.Read(reader1, binary.BigEndian, &key3)

	reader1 := NewReader(k)
	key0 = reader1.ReadU32()
	key1 = reader1.ReadU32()
	key2 = reader1.ReadU32()
	key3 = reader1.ReadU32()

	k0 = uint64(key0)
	k1 = uint64(key1)
	k2 = uint64(key2)
	k3 = uint64(key3)

	//reader2 := bytes.NewReader(v)
	//_ = binary.Read(reader2, binary.BigEndian, &val0)
	//_ = binary.Read(reader2, binary.BigEndian, &val1)

	reader2 := NewReader(v)
	val0 = reader2.ReadU32()
	val1 = reader2.ReadU32()

	v0 = uint64(val0)
	v1 = uint64(val1)

	sum := uint64(0)
	for _ = range ROUND {
		sum += DELTA
		v0 += ((OP & (v1 << 4)) + k0) ^ (v1 + sum) ^ ((OP & (v1 >> 5)) + k1)
		v0 &= OP
		v1 += ((OP & (v0 << 4)) + k2) ^ (v0 + sum) ^ ((OP & (v0 >> 5)) + k3)
		v1 &= OP
	}

	buf := make([]byte, 0)
	buf = binary.BigEndian.AppendUint32(buf, uint32(v0))
	buf = binary.BigEndian.AppendUint32(buf, uint32(v1))

	return buf
}

func teaDecipher(v, k []byte) []byte {
	var key0, key1, key2, key3, val0, val1 uint32
	var k0, k1, k2, k3, v0, v1 uint64

	//reader1 := bytes.NewReader(k)
	//_ = binary.Read(reader1, binary.BigEndian, &key0)
	//_ = binary.Read(reader1, binary.BigEndian, &key1)
	//_ = binary.Read(reader1, binary.BigEndian, &key2)
	//_ = binary.Read(reader1, binary.BigEndian, &key3)

	reader1 := NewReader(k)
	key0 = reader1.ReadU32()
	key1 = reader1.ReadU32()
	key2 = reader1.ReadU32()
	key3 = reader1.ReadU32()

	k0 = uint64(key0)
	k1 = uint64(key1)
	k2 = uint64(key2)
	k3 = uint64(key3)

	//reader2 := bytes.NewReader(v)
	//_ = binary.Read(reader2, binary.BigEndian, &val0)
	//_ = binary.Read(reader2, binary.BigEndian, &val1)

	reader2 := NewReader(v)
	val0 = reader2.ReadU32()
	val1 = reader2.ReadU32()

	v0 = uint64(val0)
	v1 = uint64(val1)

	sum := DELTA << 4 & OP
	for _ = range ROUND {
		v1 -= ((v0 << 4) + k2) ^ (v0 + sum) ^ ((v0 >> 5) + k3)
		v1 &= OP
		v0 -= ((v1 << 4) + k0) ^ (v1 + sum) ^ ((v1 >> 5) + k1)
		v0 &= OP
		sum -= DELTA
		sum &= OP
	}

	buf := make([]byte, 0)
	buf = binary.BigEndian.AppendUint32(buf, uint32(v0))
	buf = binary.BigEndian.AppendUint32(buf, uint32(v1))

	return buf
}

func preprocess(data []byte) []byte {
	dataLen := len(data)
	// 保证取余出来是正数
	filln := ((8-(dataLen+2))%8+8)%8 + 2
	fills := make([]byte, 0)
	for _ = range filln {
		fills = append(fills, 220)
	}
	return append(append(append([]byte{byte((filln - 2) | 0xf8)}, fills...), data...), make([]byte, 7)...)
}

func encrypt(data, key []byte) []byte {
	if len(key) < 16 {
		panic("key length must be 16")
	}

	data = preprocess(data)
	tr := make([]byte, 8)
	to := make([]byte, 8)
	result := make([]byte, 0)
	for i := 0; i < len(data); i += 8 {
		o := xor(data[i:i+8], tr)
		tr = xor(teaCode(o, key), to)
		to = o
		result = append(result, tr...)
	}
	return result
}

func decrypt(text, key []byte) []byte {
	if len(key) < 16 {
		panic("key length must be 16")
	}

	dataLen := len(text)
	plain := teaDecipher(text, key)
	pos := (plain[0] & 0x07) + 2
	ret := plain
	precrypt := text[0:8]
	for i := 8; i < dataLen; i += 8 {
		x := xor(teaDecipher(xor(text[i:i+8], plain), key), precrypt)
		plain = xor(x, precrypt)
		precrypt = text[i : i+8]
		ret = append(ret, x...)
	}
	if !bytes.Equal(ret[len(ret)-7:], make([]byte, 7)) {
		return nil
	}

	return ret[pos+1 : len(ret)-7]
}

func QQTeaEncrypt(data, key []byte) []byte {
	return encrypt(data, key)
}

func QQTeaDecrypt(data, key []byte) []byte {
	return decrypt(data, key)
}
