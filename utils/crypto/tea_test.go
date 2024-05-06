package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestTeaOldNew(t *testing.T) {
	s := "hello tea"
	key := []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
	tea := NewTeaCipher(key)
	encryp := oldencrypt([]byte(s), key)
	raw := string(olddecrypt(encryp, key))
	if s == raw {
		t.Log("pass")
	} else {
		t.Error("fatal")
	}
	encnew := tea.Encrypt([]byte(s))
	raw = string(olddecrypt(encnew, key))
	if s == raw {
		t.Log("pass")
	} else {
		t.Error("fatal")
	}
	encnew = oldencrypt([]byte(s), key)
	raw = string(tea.Decrypt(encnew))
	if s == raw {
		t.Log("pass")
	} else {
		t.Error("fatal")
	}
}

// below from https://github.com/Mrs4s/MiraiGo/blob/master/binary/tea_test.go

var testTEA = NewTeaCipher([]byte("0123456789ABCDEF"))

const (
	KEY = iota
	DAT
	ENC
)

var sampleData = func() [][3]string {
	out := [][3]string{
		{"0123456789ABCDEF", "MiraiGO Here", "b7b2e52af7f5b1fbf37fc3d5546ac7569aecd01bbacf09bf"},
		{"0123456789ABCDEF", "LXY Testing~", "9d0ab85aa14f5434ee83cd2a6b28bf306263cdf88e01264c"},

		{"0123456789ABCDEF", "s", "528e8b5c48300b548e94262736ebb8b7"},
		{"0123456789ABCDEF", "long long long long long long long", "95715fab6efbd0fd4b76dbc80bd633ebe805849dbc242053b06557f87e748effd9f613f782749fb9fdfa3f45c0c26161"},

		{"LXY1226    Mrs4s", "LXY Testing~", "ab20caa63f3a6503a84f3cb28f9e26b6c18c051e995d1721"},
	}
	for i := range out {
		c, _ := hex.DecodeString(out[i][ENC])
		out[i][ENC] = string(c)
	}
	return out
}()

func TestTEA(t *testing.T) {
	// Self Testing
	for _, sample := range sampleData {
		tea := NewTeaCipher([]byte(sample[KEY]))
		dat := string(tea.Decrypt([]byte(sample[ENC])))
		if dat != sample[DAT] {
			t.Fatalf("error decrypt %v %x", sample, dat)
		}
		enc := string(tea.Encrypt([]byte(sample[DAT])))
		dat = string(tea.Decrypt([]byte(enc)))
		if dat != sample[DAT] {
			t.Fatal("error self test", sample)
		}
	}

	key := make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	// Random data testing
	for i := 1; i < 0xFF; i++ {
		_, err := rand.Read(key)
		if err != nil {
			panic(err)
		}
		tea := NewTeaCipher(key)

		dat := make([]byte, i)
		_, err = rand.Read(dat)
		if err != nil {
			panic(err)
		}
		enc := tea.Encrypt(dat)
		dec := tea.Decrypt(enc)
		if !bytes.Equal(dat, dec) {
			t.Fatalf("error in %d, %x %x %x", i, key, dat, enc)
		}
	}
}

func benchEncrypt(b *testing.B, data []byte) {
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testTEA.Encrypt(data)
	}
}

func benchDecrypt(b *testing.B, data []byte) {
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	data = testTEA.Encrypt(data)
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testTEA.Decrypt(data)
	}
}

func BenchmarkTEAen(b *testing.B) {
	b.Run("16", func(b *testing.B) {
		data := make([]byte, 16)
		benchEncrypt(b, data)
	})
	b.Run("256", func(b *testing.B) {
		data := make([]byte, 256)
		benchEncrypt(b, data)
	})
	b.Run("4K", func(b *testing.B) {
		data := make([]byte, 1024*4)
		benchEncrypt(b, data)
	})
	b.Run("32K", func(b *testing.B) {
		data := make([]byte, 1024*32)
		benchEncrypt(b, data)
	})
}

func BenchmarkTEAde(b *testing.B) {
	b.Run("16", func(b *testing.B) {
		data := make([]byte, 16)
		benchDecrypt(b, data)
	})
	b.Run("256", func(b *testing.B) {
		data := make([]byte, 256)
		benchDecrypt(b, data)
	})
	b.Run("4K", func(b *testing.B) {
		data := make([]byte, 4096)
		benchDecrypt(b, data)
	})
	b.Run("32K", func(b *testing.B) {
		data := make([]byte, 1024*32)
		benchDecrypt(b, data)
	})
}

func BenchmarkTEA_encode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testTEA.encode(114514)
	}
}

func BenchmarkTEA_decode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testTEA.decode(114514)
	}
}

// above from https://github.com/Mrs4s/MiraiGo/blob/master/binary/tea_test.go

func benchOldEncrypt(b *testing.B, data []byte) {
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		olddecrypt(data, []byte("0123456789ABCDEF"))
	}
}

func benchOldDecrypt(b *testing.B, data []byte) {
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	data = testTEA.Encrypt(data)
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		olddecrypt(data, []byte("0123456789ABCDEF"))
	}
}

func BenchmarkOldTEAen(b *testing.B) {
	b.Run("16", func(b *testing.B) {
		data := make([]byte, 16)
		benchOldEncrypt(b, data)
	})
	b.Run("256", func(b *testing.B) {
		data := make([]byte, 256)
		benchOldEncrypt(b, data)
	})
	b.Run("4K", func(b *testing.B) {
		data := make([]byte, 1024*4)
		benchOldEncrypt(b, data)
	})
	b.Run("32K", func(b *testing.B) {
		data := make([]byte, 1024*32)
		benchOldEncrypt(b, data)
	})
}

func BenchmarkOldTEAde(b *testing.B) {
	b.Run("16", func(b *testing.B) {
		data := make([]byte, 16)
		benchOldDecrypt(b, data)
	})
	b.Run("256", func(b *testing.B) {
		data := make([]byte, 256)
		benchOldDecrypt(b, data)
	})
	b.Run("4K", func(b *testing.B) {
		data := make([]byte, 4096)
		benchOldDecrypt(b, data)
	})
	b.Run("32K", func(b *testing.B) {
		data := make([]byte, 1024*32)
		benchOldDecrypt(b, data)
	})
}

// 使用utils里面的reader会循环引用

type testreader struct {
	buffer []byte
	pos    int
}

func newtestreader(buffer []byte) *testreader {
	return &testreader{
		buffer: buffer,
		pos:    0,
	}
}

func (r *testreader) ReadU32() (v uint32) {
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
	readera := newtestreader(a)
	readerb := newtestreader(b)

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

	reader1 := newtestreader(k)
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

	reader2 := newtestreader(v)
	val0 = reader2.ReadU32()
	val1 = reader2.ReadU32()

	v0 = uint64(val0)
	v1 = uint64(val1)

	sum := uint64(0)
	for i := 0; i < ROUND; i++ {
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

	reader1 := newtestreader(k)
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

	reader2 := newtestreader(v)
	val0 = reader2.ReadU32()
	val1 = reader2.ReadU32()

	v0 = uint64(val0)
	v1 = uint64(val1)

	sum := DELTA << 4 & OP
	for i := 0; i < ROUND; i++ {
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
	for i := 0; i < filln; i++ {
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

func oldencrypt(data, key []byte) []byte {
	return encrypt(data, key)
}

func olddecrypt(data, key []byte) []byte {
	return decrypt(data, key)
}
