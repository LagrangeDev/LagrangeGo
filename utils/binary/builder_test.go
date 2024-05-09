package binary

import (
	"fmt"
	"testing"
)

func TestBuilder(t *testing.T) {
	r := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	data := NewBuilder(nil).WriteLenBytes(r).ToBytes()
	fmt.Printf("%x\n", data)
}

func build() []byte {
	r := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	return NewBuilder(nil).WriteLenBytes(r).WriteLenBytes(r).ToBytes()
}

func TestReader(t *testing.T) {
	data := build()
	reader := NewReader(data)
	b := reader.ReadBytesWithLength("u16", false)
	b2 := reader.ReadBytesWithLength("u16", false)
	fmt.Printf("%x\n%x\n", b, b2)
}
