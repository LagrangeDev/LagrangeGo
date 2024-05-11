package binary

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func TestBuilder(t *testing.T) {
	r := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	data := NewBuilder(nil).WriteLenBytes(r).ToBytes()
	exp, err := hex.DecodeString("0009010203040506070809")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, exp) {
		t.Fatal("expected", hex.EncodeToString(exp), "but got", hex.EncodeToString(data))
	}
}

// from https://github.com/Mrs4s/MiraiGo/blob/master/binary/writer_test.go

func BenchmarkNewBuilder128(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewBuilder(nil).WriteBytes(test).ToBytes()
		}
	})
}

func BenchmarkNewBuilder128_3(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewBuilder(nil).
				WriteBytes(test).
				WriteBytes(test).
				WriteBytes(test).
				ToBytes()
		}
	})
}

func BenchmarkNewBuilder128_5(b *testing.B) {
	test := make([]byte, 128)
	rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewBuilder(nil).
				WriteBytes(test).
				WriteBytes(test).
				WriteBytes(test).
				WriteBytes(test).
				WriteBytes(test).
				ToBytes()
		}
	})
}
