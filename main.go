// nolint
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

var mainLogger = utils.GetLogger("main")

func main() {
	appInfo := auth.AppList["linux"]
	deviceInfo := &auth.DeviceInfo{
		Guid:          "cfcd208495d565ef66e7dff9f98764da",
		DeviceName:    "Lagrange-DCFCD07E",
		SystemKernel:  "Windows 10.0.22631",
		KernelVersion: "10.0.22631",
	}

	qqclient := client.NewClient(0, "https://sign.lagrangecore.org/api/sign", appInfo)
	qqclient.UseDevice(deviceInfo)
	data, err := os.ReadFile("sig.bin")
	if err != nil {
		mainLogger.Warnln("read sig error:", err)
	} else {
		sig, err := auth.UnmarshalSigInfo(data, true)
		if err != nil {
			mainLogger.Warnln("load sig error:", err)
		} else {
			qqclient.UseSig(sig)
		}
	}

	qqclient.GroupMessageEvent.Subscribe(func(client *client.QQClient, event *message.GroupMessage) {
		if event.ToString() == "114514" {
			img, _ := os.ReadFile("testgroup.png")
			_, err := client.SendGroupMessage(event.GroupCode, []message.IMessageElement{&message.GroupImageElement{Stream: img}})
			if err != nil {
				return
			}
		}
	})

	qqclient.PrivateMessageEvent.Subscribe(func(client *client.QQClient, event *message.PrivateMessage) {
		img, _ := os.ReadFile("testprivate.png")
		_, err := client.SendPrivateMessage(event.Sender.Uin, []message.IMessageElement{&message.FriendImageElement{Stream: img}})
		if err != nil {
			return
		}
	})

	err = qqclient.Login("", "qrcode.png")
	if err != nil {
		mainLogger.Errorln("login err:", err)
		return
	}

	defer qqclient.Release()

	defer func() {
		data, err = qqclient.Sig().Marshal()
		if err != nil {
			mainLogger.Errorln("marshal sig.bin err:", err)
			return
		}
		err = os.WriteFile("sig.bin", data, 0644)
		if err != nil {
			mainLogger.Errorln("write sig.bin err:", err)
			return
		}
		mainLogger.Infoln("sig saved into sig.bin")
	}()

	// setup the main stop channel
	mc := make(chan os.Signal, 2)
	signal.Notify(mc, os.Interrupt, syscall.SIGTERM)
	for {
		switch <-mc {
		case os.Interrupt, syscall.SIGTERM:
			return
		}
	}
}
