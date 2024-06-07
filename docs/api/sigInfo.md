# SigInfo

位于`github.com/LagrangeDev/LagrangeGo/client/auth`

> sig的序列化与反序列化
```go
// 序列化，得到的data可自行存储
data, err := sig.Marshal()

// 反序列化
sig, err := UnmarshalSigInfo(data, true)
```

> 存储与加载sig示例
```go
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
```
