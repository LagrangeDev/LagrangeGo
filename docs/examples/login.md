# 登录

## 自动选择登录方式（建议使用此方式）

```go
err = qqclient.Login("password", "path/of/qrcode.png")
```

:::tip 提示
当sig内有登录信息时，会优先进行快速登录

密码为空则是扫码登录
:::

## 二维码登录

> 首先获取登录二维码

```go
qrcode, url, err = qqclient.FecthQRCode()
```

返回的元组包括
| 字段 | 类型 | 备注 |
|:------:|:------:|:-------:|
| `qrcode` | []byte | 二维码图片数据 |
| `url`    | string | 二维码内容链接 |
| `err`    | error | 错误信息 |

:::tip 提示
二维码内容链接需要被转换为二维码图片后通过手机 App 扫码登录, 不要直接访问链接
:::

## 密码登录

> 不保证可用性

```go
err = qqclient.PasswordLogin("password")
```
