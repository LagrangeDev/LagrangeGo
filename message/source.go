package message

// from https://github.com/Mrs4s/MiraiGo/blob/master/message/source.go

type SourceType byte

// MessageSourceType 常量
const (
	SourcePrivate SourceType = 1 << iota
	SourceGroup   SourceType = 1 << iota
)

func (t SourceType) String() string {
	switch t {
	case SourcePrivate:
		return "私聊"
	case SourceGroup:
		return "群聊"
	default:
		return "unknown"
	}
}

// Source 消息来源
type Source struct {
	SourceType SourceType
	PrimaryID  int64 // 群号/QQ号/guild_id
}
