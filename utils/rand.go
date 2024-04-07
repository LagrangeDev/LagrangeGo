package utils

import (
	"math/rand/v2"
	"time"
)

// 这是 math/rand/v2 生成u32
// cpu: AMD Ryzen 7 5800H with Radeon Graphics
// BenchmarkRand-16        42223339                27.55 ns/op

// 这是使用原来的 math/rand
// cpu: AMD Ryzen 7 5800H with Radeon Graphics
// BenchmarkRand-16          154076              7780 ns/op

// RandU32 生成随机u32 用math/rand/v2，让你飞起来
func RandU32() uint32 {
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixMilli())))
	return r.Uint32()
}
