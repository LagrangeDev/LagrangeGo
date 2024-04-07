package binary

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

// WriteU8(8).WriteU16(15).WriteU64(88).WriteI32(-59)
var buffer, _ = hex.DecodeString("08000f0000000000000058ffffffc5")

func Read(n int) {
	for _ = range n {
		buf := make([]byte, len(buffer))
		copy(buf, buffer)
		reader := bytes.NewReader(buf)
		var u8 uint8
		var u16 uint16
		var u64 uint64
		var i32 int32

		_ = binary.Read(reader, binary.BigEndian, &u8)
		_ = binary.Read(reader, binary.BigEndian, &u16)
		_ = binary.Read(reader, binary.BigEndian, &u64)
		_ = binary.Read(reader, binary.BigEndian, &i32)
	}
}

func TestName(t *testing.T) {
	NoRead(1)
}

func NoRead(n int) {
	for _ = range n {
		buf := make([]byte, len(buffer))
		copy(buf, buffer)
		_ = buf[0]
		_ = binary.BigEndian.Uint16(buf[1:3])
		_ = binary.BigEndian.Uint64(buf[3:11])
		_ = int32(binary.BigEndian.Uint32(buf[11:]))
	}
}

func BenchmarkReader(b *testing.B) {
	b.Run("read 1t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Read(1)
		}
	})
	b.Run("read 8t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Read(8)
		}
	})
	b.Run("read 16t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Read(16)
		}
	})
	b.Run("read 32t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Read(32)
		}
	})
	b.Run("read 64t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Read(64)
		}
	})
	b.Run("read 128t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Read(128)
		}
	})
	// -------------------------------------------
	b.Run("no read 1t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NoRead(1)
		}
	})
	b.Run("no read 8t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NoRead(8)
		}
	})
	b.Run("no read 16t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NoRead(16)
		}
	})
	b.Run("no read 32t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NoRead(32)
		}
	})
	b.Run("no read 64t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NoRead(64)
		}
	})
	b.Run("no read 128t", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NoRead(128)
		}
	})
}
