package main

import (
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/info"
)

func main() {
	appInfo := info.AppList["linux"]
	deviceInfo := &info.DeviceInfo{
		Guid:          "cfcd208495d565ef66e7dff9f98764da",
		DeviceName:    "Lagrange-DCFCD07E",
		SystemKernel:  "Windows 10.0.22631",
		KernelVersion: "10.0.22631",
	}
	sig := info.NewSigInfo(8848)

	qqclient := client.NewQQclient(0, "", appInfo, deviceInfo, sig)
	qqclient.Loop()

	_, err := qqclient.Login("", "./qrcode.png")
	if err != nil {
		fmt.Println(err)
	}

	select {}
}
