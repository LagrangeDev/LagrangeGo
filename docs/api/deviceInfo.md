# DeviceInfo

位于`github.com/LagrangeDev/LagrangeGo/client/auth`

> 创建一个新的DeviceInfo，使用随机数字作为参数
```go
deviceInfo := NewDeviceInfo(114514)
```

> 其他使用方法
```go
// 加载DeviceInfo，如果指定的路径不存在，则返回一个新的info并保存
deviceInfo := auth.LoadOrSaveDevice(path)

// 保存DeviceInfo
deviceInfo.Save(path)
```
