# 创建一个bot实例

## QQClient

位于`github.com/LagrangeDev/LagrangeGo/client`

> 创建一个QQClient，参数分别是qq号，[sign地址](/guide/sign)，[appinfo](/api/appInfo)

```go
qqclient := client.NewClient(0, "https://sign.lagrangecore.org/api/sign", appInfo)
```

> 使用指定的[DeviceInfo](/api/deviceInfo)
```go
qqclient.UseDevice(deviceInfo)
```

> 使用指定的[sig](/api/sigInfo)
```go
qqclient.UseSig(sig)
```
