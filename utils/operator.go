package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func TimeStamp() int64 {
	return time.Now().Unix()
}

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
