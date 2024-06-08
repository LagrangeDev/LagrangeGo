# 事件

> 使用`EventHandle[T].Subscribe(func(client *QQClient, event T))`来订阅并处理事件
> 
> 当指定的事件发生时，对应的回调函数将被执行

```go
qqclient.GroupMessageEvent.Subscribe(func(client *client.QQClient, event *message.GroupMessage) {
	// 你可以从event中获取事件的各个参数
    fmt.Println(event.ToString())
})
```

这段代码会将群聊收到的消息打印出来

:::warning 注意
bot自身发送的消息也将被作为事件处理，请开发者注意消息处理逻辑的编写，防止出现发送循环

例如，以下代码会导致无限循环发送消息

```go
qqclient.GroupMessageEvent.Subscribe(func(client *client.QQClient, event *message.GroupMessage) {
	if event.ToString() == "114514" {
        client.SendGroupMessage(event.GroupCode, []message.IMessageElement{NewText("114514")})
    }
})
```
:::

> `QQClient`目前支持的EventHandle

|                        EventHandle[T]                        |    描述    |
|:------------------------------------------------------------:|:--------:|
|             `EventHandle[*message.GroupMessage]`             |  群聊消息事件  |
|        `PrivateMessageEvent[*message.PrivateMessage]`        |  私聊消息事件  |
|           `TempMessageEvent[*message.TempMessage]`           | 临时会话消息事件 |
|           `GroupInvitedEvent[*event.GroupInvite]`            |  被邀请入群   |
| `GroupMemberJoinRequestEvent[*event.GroupMemberJoinRequest]` |   加群申请   |
|      `GroupMemberJoinEvent[*event.GroupMemberIncrease]`      |   成员入群   |
|     `GroupMemberLeaveEvent[*event.GroupMemberDecrease]`      |   成员退群   |
|              `GroupMuteEvent[*event.GroupMute]`              |   群聊禁言   |
|            `GroupRecallEvent[*event.GroupRecall]`            |  群聊撤回消息  |
|          `FriendRequestEvent[*event.FriendRequest]`          |   好友申请   |
|           `FriendRecallEvent[*event.FriendRecall]`           |  好友消息撤回  |
|                 `RenameEvent[*event.Rename]`                 |   昵称变动   |
