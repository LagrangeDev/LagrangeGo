package binary

import (
	"io"
	"testing"
)

// TestReaderEmptyBuffer 测试空缓冲区的情况
func TestReaderEmptyBuffer(t *testing.T) {
	// 测试空缓冲区
	r := NewReader([]byte{})

	// 测试ReadU8
	if v := r.ReadU8(); v != 0 {
		t.Errorf("ReadU8 with empty buffer should return 0, got %d", v)
	}

	// 测试ReadU16
	if v := r.ReadU16(); v != 0 {
		t.Errorf("ReadU16 with empty buffer should return 0, got %d", v)
	}

	// 测试ReadU32
	if v := r.ReadU32(); v != 0 {
		t.Errorf("ReadU32 with empty buffer should return 0, got %d", v)
	}

	// 测试ReadU64
	if v := r.ReadU64(); v != 0 {
		t.Errorf("ReadU64 with empty buffer should return 0, got %d", v)
	}

	// 测试ReadBytes
	if bytes := r.ReadBytes(10); len(bytes) != 0 {
		t.Errorf("ReadBytes with empty buffer should return empty slice, got %v", bytes)
	}

	// 测试ReadBytesNoCopy
	if bytes := r.ReadBytesNoCopy(10); len(bytes) != 0 {
		t.Errorf("ReadBytesNoCopy with empty buffer should return empty slice, got %v", bytes)
	}
}

// TestReaderIncompleteData 测试不完整数据的情况
func TestReaderIncompleteData(t *testing.T) {
	// 测试不完整数据 - 只有1个字节
	r := NewReader([]byte{0x01})

	// 测试ReadU16 (需要2字节)
	if v := r.ReadU16(); v != 0 {
		t.Errorf("ReadU16 with incomplete data should return 0, got %d", v)
	}

	// 重置Reader
	r = NewReader([]byte{0x01, 0x02})

	// 测试ReadU32 (需要4字节)
	if v := r.ReadU32(); v != 0 {
		t.Errorf("ReadU32 with incomplete data should return 0, got %d", v)
	}

	// 测试ReadBytes超出可用长度
	r = NewReader([]byte{0x01, 0x02, 0x03})
	if bytes := r.ReadBytes(10); len(bytes) != 0 {
		t.Errorf("ReadBytes with insufficient data should return empty slice, got %v", bytes)
	}
}

// TestReaderWithIOReader 测试使用io.Reader的情况
func TestReaderWithIOReader(t *testing.T) {
	// 创建一个会返回错误的Reader
	errReader := &errorReader{}
	r := ParseReader(errReader)

	// 测试ReadU8
	if v := r.ReadU8(); v != 0 {
		t.Errorf("ReadU8 with error reader should return 0, got %d", v)
	}

	// 测试ReadU16
	if v := r.ReadU16(); v != 0 {
		t.Errorf("ReadU16 with error reader should return 0, got %d", v)
	}

	// 测试ReadU32
	if v := r.ReadU32(); v != 0 {
		t.Errorf("ReadU32 with error reader should return 0, got %d", v)
	}

	// 测试ReadBytes
	if bytes := r.ReadBytes(10); len(bytes) != 0 {
		t.Errorf("ReadBytes with error reader should return empty slice, got %v", bytes)
	}

	// 测试ReadAll
	if data := r.ReadAll(); data != nil {
		t.Errorf("ReadAll with error reader should return nil, got %v", data)
	}
}

// TestReaderNormalData 测试正常数据的情况
func TestReaderNormalData(t *testing.T) {
	// 准备测试数据
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	r := NewReader(data)

	// 测试ReadU8
	if v := r.ReadU8(); v != 0x01 {
		t.Errorf("ReadU8 should return 0x01, got 0x%02x", v)
	}

	// 测试ReadU16
	if v := r.ReadU16(); v != 0x0203 {
		t.Errorf("ReadU16 should return 0x0203, got 0x%04x", v)
	}

	// 测试ReadU32
	if v := r.ReadU32(); v != 0x04050607 {
		t.Errorf("ReadU32 should return 0x04050607, got 0x%08x", v)
	}

	// 测试ReadByte
	if b, err := r.ReadByte(); err != nil || b != 0x08 {
		t.Errorf("ReadByte should return 0x08, got 0x%02x, err: %v", b, err)
	}

	// 测试读取完所有数据后的ReadByte
	if _, err := r.ReadByte(); err != io.EOF {
		t.Errorf("ReadByte after end should return EOF, got %v", err)
	}
}

// TestReaderShortRead 测试短读的情况
func TestReaderShortRead(t *testing.T) {
	// 创建一个会返回短读的Reader
	shortReader := &shortReader{data: []byte{0x01, 0x02, 0x03, 0x04}}
	r := ParseReader(shortReader)

	// 测试ReadBytes
	if bytes := r.ReadBytes(10); len(bytes) != 0 {
		t.Errorf("ReadBytes with short reader should return empty slice, got %v", bytes)
	}
}

// 辅助测试结构

// errorReader 总是返回错误的Reader
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

// shortReader 总是返回短读的Reader
type shortReader struct {
	data []byte
	pos  int
}

func (r *shortReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	// 只读取一个字节，模拟短读
	p[0] = r.data[r.pos]
	r.pos++
	return 1, nil
}
