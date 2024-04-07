package utils

import (
	"fmt"
	"testing"
)

func TestSign(t *testing.T) {

	sp := SignProvider("https://sign.libfekit.so/api/sign")
	data := sp("wtlogin.login", 7, []byte{1, 2, 3, 8, 9, 6, 3})
	fmt.Println(data)
}

func TestTimeStamp(t *testing.T) {
	fmt.Println(TimeStamp())
}
