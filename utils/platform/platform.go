package platform

import "runtime"

func System() string {
	return runtime.GOOS
}

func Version() string {
	return runtime.GOARCH
}
