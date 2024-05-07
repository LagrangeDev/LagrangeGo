package crypto

import (
	_ "unsafe" // required by go:linkname
)

// randuint32 returns a lock free uint32 value.
//
//go:linkname randuint32 runtime.fastrand
func randuint32() uint32

func RandU32() uint32 {
	return randuint32()
}
