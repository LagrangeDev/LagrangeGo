package utils

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"
)

func ReadLine(s string) string {
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Print(s)
	rs, err := inputReader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(rs)
}

func Bool2Int(v bool) int {
	if v {
		return 1
	}
	return 0
}

func CloseIO(reader io.Reader) {
	if closer, ok := reader.(io.Closer); ok {
		_ = closer.Close()
	}
}

func NewUUID() string {
	u := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, u)
	if err != nil {
		return ""
	}

	// Set the version to 4 (randomly generated UUID)
	u[6] = (u[6] & 0x0f) | 0x40
	// Set the variant to RFC 4122
	u[8] = (u[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func Ternary[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

func Map[T any, U any](list []T, mapper func(T) U) []U {
	result := make([]U, len(list))
	for i, v := range list {
		result[i] = mapper(v)
	}
	return result
}
