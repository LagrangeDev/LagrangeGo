package utils

import (
	"math/rand"
	"time"
)

func RandU32() uint32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Uint32()
}
