package cache

import "github.com/LagrangeDev/LagrangeGo/client/entity"

// GetUid 根据uin获取uid
func (c *Cache) GetUid(uin uint32, groupUin ...uint32) string {
	if len(groupUin) == 0 {
		if friend, ok := getCacheOf[entity.Friend](c, uin); ok {
			return friend.Uid
		}
		return ""
	}
	if group, ok := getCacheOf[Cache](c, groupUin[0]); ok {
		if member, ok := getCacheOf[entity.GroupMember](group, uin); ok {
			return member.Uid
		}
	}
	return ""
}

// GetUin 根据uid获取uin
func (c *Cache) GetUin(uid string, groupUin ...uint32) (uin uint32) {
	if len(groupUin) == 0 {
		rangeCacheOf[entity.Friend](c, func(k uint32, friend *entity.Friend) bool {
			if friend.Uid == uid {
				uin = k
				return false
			}
			return true
		})
		return uin
	}
	if group, ok := getCacheOf[Cache](c, groupUin[0]); ok {
		rangeCacheOf[entity.GroupMember](group, func(k uint32, member *entity.GroupMember) bool {
			if member.Uid == uid {
				uin = k
				return false
			}
			return true
		})
		return uin
	}
	return 0
}

// GetFriend 获取好友信息
func (c *Cache) GetFriend(uin uint32) *entity.Friend {
	v, _ := getCacheOf[entity.Friend](c, uin)
	return v
}

// GetAllFriends 获取所有好友信息
func (c *Cache) GetAllFriends() map[uint32]*entity.Friend {
	friends := make(map[uint32]*entity.Friend, 64)
	rangeCacheOf[entity.Friend](c, func(k uint32, friend *entity.Friend) bool {
		friends[k] = friend
		return true
	})
	return friends
}

// GetGroupInfo 获取群信息
func (c *Cache) GetGroupInfo(groupUin uint32) *entity.Group {
	v, _ := getCacheOf[entity.Group](c, groupUin)
	return v
}

// GetAllGroupsInfo 获取所有群信息
func (c *Cache) GetAllGroupsInfo() map[uint32]*entity.Group {
	groups := make(map[uint32]*entity.Group, 64)
	rangeCacheOf[entity.Group](c, func(k uint32, v *entity.Group) bool {
		groups[k] = v
		return true
	})
	return groups
}

// GetGroupMember 获取群成员信息
func (c *Cache) GetGroupMember(uin, groupUin uint32) *entity.GroupMember {
	if group, ok := getCacheOf[Cache](c, groupUin); ok {
		v, _ := getCacheOf[entity.GroupMember](group, uin)
		return v
	}
	return nil
}

// GetGroupMembers 获取指定群所有群成员信息
func (c *Cache) GetGroupMembers(groupUin uint32) map[uint32]*entity.GroupMember {
	members := make(map[uint32]*entity.GroupMember, 64)
	if group, ok := getCacheOf[Cache](c, groupUin); ok {
		rangeCacheOf[entity.GroupMember](group, func(k uint32, member *entity.GroupMember) bool {
			members[member.Uin] = member
			return true
		})
	}
	return members
}

// GetRKeyInfo 获取指定类型的RKey信息
func (c *Cache) GetRKeyInfo(rkeyType entity.RKeyType) *entity.RKeyInfo {
	v, _ := getCacheOf[entity.RKeyInfo](c, rkeyType)
	return v
}

// GetAllRkeyInfo 获取所有RKey信息
func (c *Cache) GetAllRkeyInfo() entity.RKeyMap {
	infos := make(map[entity.RKeyType]*entity.RKeyInfo, 2)
	rangeCacheOf[entity.RKeyInfo](c, func(k entity.RKeyType, v *entity.RKeyInfo) bool {
		infos[k] = v
		return true
	})
	return infos
}

// FriendCacheIsEmpty 好友信息缓存是否为空
func (c *Cache) FriendCacheIsEmpty() bool {
	return !hasRefreshed[entity.Friend](c)
}

// GroupMembersCacheIsEmpty 群成员缓存是否为空
func (c *Cache) GroupMembersCacheIsEmpty() bool {
	return !hasRefreshed[Cache](c)
}

// GroupMemberCacheIsEmpty 指定群的群成员缓存是否为空
func (c *Cache) GroupMemberCacheIsEmpty(groupUin uint32) bool {
	if group, ok := getCacheOf[Cache](c, groupUin); ok {
		return !hasRefreshed[entity.GroupMember](group)
	}
	return true
}

// GroupInfoCacheIsEmpty 群信息缓存是否为空
func (c *Cache) GroupInfoCacheIsEmpty() bool {
	return !hasRefreshed[entity.Group](c)
}

// RkeyInfoCacheIsEmpty RKey缓存是否为空
func (c *Cache) RkeyInfoCacheIsEmpty() bool {
	return !hasRefreshed[entity.RKeyInfo](c)
}

func (c *Cache) RkeyInfoCacheIsExpired() bool {
	return hasExpired[entity.RKeyInfo](c)
}
