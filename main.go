// nolint
package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/packets/wtlogin/loginstate"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

var (
	dumpspath = "dump"
)

func main() {
	appInfo := auth.AppList["linux"]["3.2.10-25765"]
	deviceInfo := &auth.DeviceInfo{
		GUID:          "cfcd208495d565ef66e7dff9f98764da",
		DeviceName:    "Lagrange-DCFCD07E",
		SystemKernel:  "Windows 10.0.22631",
		KernelVersion: "10.0.22631",
	}

	qqclient := client.NewClient(0, "", appInfo, "https://sign.lagrangecore.org/api/sign/25765")
	qqclient.SetLogger(protocolLogger{})
	qqclient.UseDevice(deviceInfo)
	data, err := os.ReadFile("sig.bin")
	if err != nil {
		logrus.Warnln("read sig error:", err)
	} else {
		sig, err := auth.UnmarshalSigInfo(data, true)
		if err != nil {
			logrus.Warnln("load sig error:", err)
		} else {
			qqclient.UseSig(sig)
		}
	}

	qqclient.GroupMessageEvent.Subscribe(func(client *client.QQClient, event *message.GroupMessage) {
		if event.ToString() == "114514" {
			img, _ := message.NewFileImage("testgroup.png")
			_, err := client.SendGroupMessage(event.GroupUin, []message.IMessageElement{img})
			if err != nil {
				return
			}
		}
	})

	qqclient.PrivateMessageEvent.Subscribe(func(client *client.QQClient, event *message.PrivateMessage) {
		img, _ := message.NewFileImage("testprivate.png")
		_, err := client.SendPrivateMessage(event.Sender.Uin, []message.IMessageElement{img})
		if err != nil {
			return
		}
	})

	err = func(c *client.QQClient) error {
		err := c.FastLogin()
		if err == nil {
			return nil
		}

		ret, err := c.PasswordLogin()
		for {
			if err != nil {
				logger.Errorf("密码登录失败: %s", err)
				break
			}
			if ret.Successful() {
				return nil
			}
			switch ret {
			case loginstate.CaptchaVerify:
				logger.Warnln("captcha verification required")
				logger.Warnln(c.Sig().CaptchaURL)
				aid := strings.Split(strings.Split(c.Sig().CaptchaURL, "&sid=")[1], "&")[0]
				logger.Warnln("ticket?->")
				ticket := utils.ReadLine()
				logger.Warnln("rand_str?->")
				randStr := utils.ReadLine()
				ret, err = c.CommitCaptcha(ticket, randStr, aid)
				continue
			case loginstate.NewDeviceVerify:
				vf, err := c.GetNewDeviceVerifyURL()
				if err != nil {
					return err
				}
				logger.Infoln(vf)
				err = c.NewDeviceVerify(vf)
				if err != nil {
					return err
				}
			default:
				logger.Errorf("Unhandled exception raised: %s", ret.Name())
			}
		}
		logger.Infoln("login with qrcode")
		png, _, err := c.FetchQRCodeDefault()
		if err != nil {
			return err
		}
		qrcodePath := "qrcode.png"
		err = os.WriteFile(qrcodePath, png, 0666)
		if err != nil {
			return err
		}
		logger.Infof("qrcode saved to %s", qrcodePath)
		for {
			retCode, err := c.GetQRCodeResult()
			if err != nil {
				logger.Errorln(err)
				return err
			}
			if retCode.Waitable() {
				time.Sleep(3 * time.Second)
				continue
			}
			if !retCode.Success() {
				return errors.New(retCode.Name())
			}
			break
		}
		return c.QRCodeLogin()
	}(qqclient)

	if err != nil {
		logger.Errorln("login err:", err)
		return
	}
	logger.Infoln("login successed")

	defer qqclient.Release()

	defer func() {
		data, err = qqclient.Sig().Marshal()
		if err != nil {
			logger.Errorln("marshal sig.bin err:", err)
			return
		}
		err = os.WriteFile("sig.bin", data, 0644)
		if err != nil {
			logger.Errorln("write sig.bin err:", err)
			return
		}
		logger.Infoln("sig saved into sig.bin")
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

// protocolLogger from https://github.com/Mrs4s/go-cqhttp/blob/a5923f179b360331786a6509eb33481e775a7bd1/cmd/gocq/main.go#L501
type protocolLogger struct{}

const fromProtocol = "Lgr -> "

func (p protocolLogger) Info(format string, arg ...any) {
	logger.Infof(fromProtocol+format, arg...)
}

func (p protocolLogger) Warning(format string, arg ...any) {
	logger.Warnf(fromProtocol+format, arg...)
}

func (p protocolLogger) Debug(format string, arg ...any) {
	logger.Debugf(fromProtocol+format, arg...)
}

func (p protocolLogger) Error(format string, arg ...any) {
	logger.Errorf(fromProtocol+format, arg...)
}

func (p protocolLogger) Dump(data []byte, format string, arg ...any) {
	message := fmt.Sprintf(format, arg...)
	if _, err := os.Stat(dumpspath); err != nil {
		err = os.MkdirAll(dumpspath, 0o755)
		if err != nil {
			logger.Errorf("出现错误 %v. 详细信息转储失败", message)
			return
		}
	}
	dumpFile := path.Join(dumpspath, fmt.Sprintf("%v.dump", time.Now().Unix()))
	logger.Errorf("出现错误 %v. 详细信息已转储至文件 %v 请连同日志提交给开发者处理", message, dumpFile)
	_ = os.WriteFile(dumpFile, data, 0o644)
}

const (
	// 定义颜色代码
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorYellow = "\x1b[33m"
	colorGreen  = "\x1b[32m"
	colorBlue   = "\x1b[34m"
	colorWhite  = "\x1b[37m"
)

var logger = logrus.New()

func init() {
	logger.SetLevel(logrus.TraceLevel)
	logger.SetFormatter(&ColoredFormatter{})
	logger.SetOutput(colorable.NewColorableStdout())
}

type ColoredFormatter struct{}

func (f *ColoredFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 获取当前时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 根据日志级别设置相应的颜色
	var levelColor string
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = colorBlue
	case logrus.InfoLevel:
		levelColor = colorGreen
	case logrus.WarnLevel:
		levelColor = colorYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = colorRed
	default:
		levelColor = colorWhite
	}

	return utils.S2B(fmt.Sprintf("[%s] [%s%s%s]: %s\n",
		timestamp, levelColor, strings.ToUpper(entry.Level.String()), colorReset, entry.Message)), nil
}
