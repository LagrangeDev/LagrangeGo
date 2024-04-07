package platform

import "runtime"

func GetSystem() string {
	return runtime.GOOS
}

func GetVersion() string {
	return runtime.GOARCH
}
