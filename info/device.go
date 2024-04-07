package info

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/utils/platform"

	"github.com/LagrangeDev/LagrangeGo/utils"
)

type DeviceInfo struct {
	Guid          string `json:"guid"`
	DeviceName    string `json:"device_name"`
	SystemKernel  string `json:"system_kernel"`
	KernelVersion string `json:"kernel_version"`
}

func NewDeviceInfo(uin int) *DeviceInfo {
	return &DeviceInfo{
		Guid:       strconv.Itoa(uin),
		DeviceName: fmt.Sprintf("Lagrange-%X", utils.Md5Digest([]byte(strconv.Itoa(uin)))[0:4]),
		// TODO 这里好像不能获取os的KernelVersion,使用go arch罢
		SystemKernel:  fmt.Sprintf("%s %s", runtime.GOOS, platform.GetVersion()),
		KernelVersion: runtime.GOARCH,
	}
}
