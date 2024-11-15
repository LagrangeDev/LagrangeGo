//go:build go1.22

package crypto

import "math/rand/v2"

func RandU32() uint32 {
	return rand.Uint32()
}
