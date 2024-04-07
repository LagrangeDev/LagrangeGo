package binary

import (
	"fmt"
	"testing"
)

func TestBuilder(t *testing.T) {
	r := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	data := NewBuilder(nil).WriteBytes(r, true).Pack(-1)
	fmt.Printf("%x\n", data)
}

func build() []byte {
	r := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	return NewBuilder(nil).WriteBytes(r, true).WriteBytes(r, true).Pack(-1)
}

func TestReader(t *testing.T) {
	data := build()
	reader := NewReader(data)
	b := reader.ReadBytesWithLength("u16", false)
	b2 := reader.ReadBytesWithLength("u16", false)
	fmt.Printf("%x\n%x\n", b, b2)
}
