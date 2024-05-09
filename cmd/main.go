// nolint
package main

import (
	info2 "github.com/LagrangeDev/LagrangeGo/internal/info"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

var mainLogger = utils.GetLogger("main")

func main() {
	// 一点优化
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())
	appInfo := info2.AppList["linux"]
	deviceInfo := &info2.DeviceInfo{
		Guid:          "cfcd208495d565ef66e7dff9f98764da",
		DeviceName:    "Lagrange-DCFCD07E",
		SystemKernel:  "Windows 10.0.22631",
		KernelVersion: "10.0.22631",
	}
	sig, err := info2.LoadSig("./sig.bin")
	if err != nil {
		mainLogger.Errorln("load sig error:", err)
		return
	}
	qqclient := client.NewQQclient(0, "https://sign.lagrangecore.org/api/sign", appInfo, deviceInfo, &sig)

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

	err = qqclient.Loop()
	if err != nil {
		mainLogger.Errorln("quit client loop:", err)
		return
	}

	err = qqclient.Login("", "./qrcode.png")
	if err != nil {
		mainLogger.Errorln("login err:", err)
		return
	}

	err = info2.SaveSig(&sig, "./sig.bin")
	if err != nil {
		mainLogger.Errorln("save sig.bin err:", err)
		return
	}

	select {}
}
