//go:build !go1.22

package crypto

import (
	_ "unsafe" // required by go:linkname
)

// randuint32 returns a lock free uint32 value.
//
// Too much legacy code has go:linkname references
// to runtime.fastrand and friends, so keep these around for now.
// Code should migrate to math/rand/v2.Uint64,
// which is just as fast, but that's only available in Go 1.22+.
// It would be reasonable to remove these in Go 1.24.
// Do not call these from package runtime.
//
//go:linkname randuint32 runtime.fastrand
func randuint32() uint32

func RandU32() uint32 {
	return randuint32()
}
