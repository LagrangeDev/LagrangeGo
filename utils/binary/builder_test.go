package binary

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mrand "math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestBuilder(t *testing.T) {
	r := make([]byte, 4096)
	_, err := rand.Read(r)
	if err != nil {
		t.Fatal(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(4096)
	for i := 0; i < 4096; i++ {
		go func(exp []byte) {
			defer wg.Done()
			defer runtime.GC()
			data := NewBuilder().WriteBytes(exp).ToBytes()
			if !bytes.Equal(data, exp) {
				panic(fmt.Sprint("expected ", hex.EncodeToString(exp),
					" but got ", hex.EncodeToString(data)))
			}
		}(r[:i])
	}
	wg.Wait()
	wg.Add(256)
	for i := 0; i < 256; i++ {
		go func(exp []byte) {
			defer wg.Done()
			defer runtime.GC()
			time.Sleep(time.Millisecond * time.Duration(mrand.Intn(10)+1))
			data := NewBuilder().WriteBytes(exp).Pack(0x2333)
			for i := 0; i < 4096; i++ {
				newdata := NewBuilder().WriteBytes(exp).Pack(0x2333)
				if !bytes.Equal(data, newdata) {
					panic("unexpected")
				}
				runtime.GC()
			}

		}(r[:i])
	}
	wg.Wait()
}

// from https://github.com/Mrs4s/MiraiGo/blob/master/binary/writer_test.go

func BenchmarkNewBuilder128(b *testing.B) {
	test := make([]byte, 128)
	_, _ = rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewBuilder().WriteBytes(test).ToBytes()
		}
	})
}

func BenchmarkNewBuilder128_3(b *testing.B) {
	test := make([]byte, 128)
	_, _ = rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewBuilder().
				WriteBytes(test).
				WriteBytes(test).
				WriteBytes(test).
				ToBytes()
		}
	})
}

func BenchmarkNewBuilder128_5(b *testing.B) {
	test := make([]byte, 128)
	_, _ = rand.Read(test)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = NewBuilder().
				WriteBytes(test).
				WriteBytes(test).
				WriteBytes(test).
				WriteBytes(test).
				WriteBytes(test).
				ToBytes()
		}
	})
}
