package info

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
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
		Guid:          fmt.Sprintf("%X", crypto.MD5Digest(utils.S2B(strconv.Itoa(uin)))),
		DeviceName:    fmt.Sprintf("Lagrange-%X", crypto.MD5Digest(utils.S2B(strconv.Itoa(uin)))[0:4]),
		SystemKernel:  fmt.Sprintf("%s %s", platform.System(), platform.Version()),
		KernelVersion: platform.Version(),
	}
}

func LoadOrSaveDevice(path string) (*DeviceInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		deviceinfo := NewDeviceInfo(int(crypto.RandU32()))
		return deviceinfo, deviceinfo.Save(path)
	}
	var dinfo DeviceInfo
	err = json.Unmarshal(data, &dinfo)
	if err != nil {
		deviceinfo := NewDeviceInfo(int(crypto.RandU32()))
		return deviceinfo, deviceinfo.Save(path)
	}
	return &dinfo, nil
}

func (deviceInfo *DeviceInfo) Save(path string) error {
	data, err := json.Marshal(deviceInfo)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, data, 0666)
	if err != nil {
		return err
	}
	return nil
}
