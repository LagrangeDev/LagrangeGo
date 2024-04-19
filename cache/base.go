package cache

import (
	"sync"

	"github.com/LagrangeDev/LagrangeGo/entity"
)

type Cache struct {
	refreshLock sync.RWMutex
	// FriendCache 好友缓存 uin *entity.Friend
	FriendCache map[uint32]*entity.Friend
	// GroupInfoCache 群信息缓存 groupUin *entity.Group
	GroupInfoCache map[uint32]*entity.Group
	// GroupMemberCache 群内群员信息缓存 groupUin uin *entity.GroupMember
	GroupMemberCache map[uint32]map[uint32]*entity.GroupMember
}

func NewCache() *Cache {
	return &Cache{
		FriendCache:      make(map[uint32]*entity.Friend),
		GroupInfoCache:   make(map[uint32]*entity.Group),
		GroupMemberCache: make(map[uint32]map[uint32]*entity.GroupMember),
	}
}
