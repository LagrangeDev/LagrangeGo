# SigInfo

位于`github.com/LagrangeDev/LagrangeGo/client/auth`

> sig的序列化与反序列化
```go
// 序列化，得到的data可自行存储
data, err := sig.Marshal()

// 反序列化
sig, err := UnmarshalSigInfo(data, true)
```
