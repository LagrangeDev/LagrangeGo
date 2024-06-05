# 创建一个bot实例

## QQClient

位于`github.com/LagrangeDev/LagrangeGo/client`

> 创建一个QQClient，参数分别是qq号，sign地址，appinfo

```go
qqclient := client.NewClient(0, "https://sign.lagrangecore.org/api/sign", appInfo)
```

> 使用指定的sig
```go
qqclient.UseSig(sig)
```

## appInfo

位于`github.com/LagrangeDev/LagrangeGo/client/auth`

> 使用内置的appInfo

```go
appInfo := auth.AppList["linux"]
```

## DeviceInfo

位于`github.com/LagrangeDev/LagrangeGo/client/auth`

> 创建一个新的DeviceInfo，使用随机数字作为参数
```go
deviceInfo := NewDeviceInfo(114514)
```

> 
```go
// 加载DeviceInfo，如果指定的路径不存在，则返回一个新的info并保存
deviceInfo := auth.LoadOrSaveDevice(path)

// 保存DeviceInfo
deviceInfo.Save(path)
```

## SigInfo

位于`github.com/LagrangeDev/LagrangeGo/client/auth`

> sig的序列化与反序列化
```go
// 序列化，得到的data可自行存储
data, err := sig.Marshal()

// 反序列化
sig, err := UnmarshalSigInfo(data, true)
```
