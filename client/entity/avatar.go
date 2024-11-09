package entity

import "fmt"

func FriendAvatar(uin uint32) string {
	return fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%v&s=640", uin)
}

func GroupAvatar(groupUin uint32) string {
	return fmt.Sprintf("https://p.qlogo.cn/gh/%v/%v/0/", groupUin, groupUin)
}
