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

func GetBytesFromHex(s string) []byte {
	result, _ := hex.DecodeString(s)
	return result
}

func ReadLine(s string) string {
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Print(s)
	rs, _ := inputReader.ReadString('\n')
	return strings.TrimSpace(rs)
}

func Bool2Int(v bool) int {
	if v {
		return 1
	}
	return 0
}
