package entity

type User struct {
	Uin          uint32
	UID          string
	Nickname     string
	Remarks      string
	PersonalSign string
	Avatar       string
	Age          uint32
	Sex          uint32 // 1男 2女 255不可见
	Level        uint32
	Source       string // 好友来源

	QID string

	Country string
	City    string
	School  string

	VipLevel uint32
}
