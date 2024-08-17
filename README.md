# LagrangeGo
ntqq 协议的golang实现 移植于 [Lagrange.Core](https://github.com/KonataDev/Lagrange.Core) /
[lagrange-python](https://github.com/LagrangeDev/lagrange-python) / [MiraiGo](https://github.com/Mrs4s/MiraiGo)

## 使用前声明
本项目为协议实现，不推荐直接使用。

## 使用方法

```bash
go get -u github.com/LagrangeDev/LagrangeGo
```

## 支持的功能

## 协议支持

<details>
  <summary>已完成功能/开发计划列表</summary>

**登录**
- [x] ~~账号密码登录~~
- [x] 二维码登录
- [ ] 验证码提交
- [ ] 设备锁验证
- [ ] 错误信息解析

**消息类型**
- [x] 文本
- [x] 图片
- [x] 语音
- [x] 表情
- [x] At
- [x] 回复
- [ ] 长消息(仅群聊/私聊)
- [ ] 链接分享
- [x] 小程序(暂只支持RAW)
- [ ] 短视频
- [x] 合并转发
- [x] 私聊文件&群文件(上传与接收信息)

**事件**
- [x] 好友消息
- [x] 群消息
- [ ] 临时会话消息
- [x] 登录号加群
- [x] 登录号退群(包含T出)
- [x] 新成员进群/退群
- [x] 群/好友消息撤回
- [x] 群禁言
- [ ] 群成员权限变更
- [x] 收到邀请进群通知
- [x] 收到其他用户进群请求
- [ ] 新好友
- [x] 新好友请求
- [ ] 客户端离线
- [ ] 群提示 (戳一戳/运气王等)

**主动操作**

_为防止滥用，不支持主动邀请新成员进群_

- [x] 发送群消息
- [x] 发送好友消息
- [ ] 发送临时会话消息
- [x] 获取/刷新群列表
- [x] 获取/刷新群成员列表
- [x] 获取/刷新好友列表
- [ ] 获取群荣誉 (龙王/群聊火焰等)
- [x] 处理加群请求
- [x] 处理被邀请加群请求
- [x] 处理好友请求
- [x] 撤回群消息
- [ ] 群公告设置
- [ ] 获取群文件下载链接
- [x] 群设置 (全体禁言/群名)
- [x] 修改群成员Card
- [x] 修改群成员头衔
- [ ] ~~群成员邀请~~
- [x] 群成员禁言/解除禁言
- [x] T出群成员
- [x] 戳一戳群友
- [ ] ~~获取陌生人信息~~

</details>

### 不支持的协议
**基于 [QQ钱包支付用户服务协议](https://www.tenpay.com/v2/html5/basic/public/agreement/protocol_mqq_pay.shtml) 不支持一切有关QQ钱包的协议**

>4.13 您不得利用本服务实施下列任一的行为：
>\
>     （9） **侵害QQ钱包支付服务系統；**

- [ ] ~~QQ钱包协议(收款/付款等)~~

### 贡献者

[![Contributors](https://contributors-img.web.app/image?repo=LagrangeDev/LagrangeGo)](https://github.com/LagrangeDev/LagrangeGo/graphs/contributors)

[MiraiGo](https://github.com/Mrs4s/MiraiGo)
[![Contributors](https://contributors-img.web.app/image?repo=Mrs4s/MiraiGo)](https://github.com/Mrs4s/MiraiGo/graphs/contributors)

