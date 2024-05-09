package entity

import "fmt"

type Friend struct {
	Uin          uint32
	Uid          string
	Nickname     string
	Remarks      string
	PersonalSign string
	Avatar       string
}

func NewFriend(uin uint32, uid, nickname, remarks, personalSign string) *Friend {
	return &Friend{
		Uin:          uin,
		Uid:          uid,
		Nickname:     nickname,
		Remarks:      remarks,
		PersonalSign: personalSign,
		Avatar:       fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%v&s=640", uin),
	}
}
