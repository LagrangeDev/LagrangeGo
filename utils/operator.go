package utils

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

func TimeStamp() int64 {
	return time.Now().Unix()
}

func MustParseHexStr(s string) []byte {
	result, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return result
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
