package entity

type (
	RKeyType uint32
	RKeyMap  map[RKeyType]*RKeyInfo
)

const (
	FriendRKey RKeyType = 10
	GroupRKey  RKeyType = 20
)

type RKeyInfo struct {
	RKeyType   RKeyType
	RKey       string
	CreateTime uint64
	ExpireTime uint64
}
