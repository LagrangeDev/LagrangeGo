package cache

import (
	entity2 "github.com/LagrangeDev/LagrangeGo/internal/entity"
)

// GetUid 根据uin获取uid
func (c *Cache) GetUid(uin uint32, groupUin ...uint32) string {
	if len(groupUin) == 0 {
		if friend, ok := getCacheOf[entity2.Friend](c, uin); ok {
			return friend.Uid
		}
		return ""
	}
	if group, ok := getCacheOf[Cache](c, groupUin[0]); ok {
		if member, ok := getCacheOf[entity2.GroupMember](group, uin); ok {
			return member.Uid
		}
	}
	return ""
}

// GetUin 根据uid获取uin
func (c *Cache) GetUin(uid string, groupUin ...uint32) (uin uint32) {
	if len(groupUin) == 0 {
		rangeCacheOf[entity2.Friend](c, func(k uint32, friend *entity2.Friend) bool {
			if friend.Uid == uid {
				uin = k
				return false
			}
			return true
		})
		return uin
	}
	if group, ok := getCacheOf[Cache](c, groupUin[0]); ok {
		rangeCacheOf[entity2.GroupMember](group, func(k uint32, member *entity2.GroupMember) bool {
			if member.Uid == uid {
				uin = k
				return false
			}
			return true
		})
	}
	return 0
}

// GetFriend 获取好友信息
func (c *Cache) GetFriend(uin uint32) *entity2.Friend {
	v, _ := getCacheOf[entity2.Friend](c, uin)
	return v
}

// GetGroupInfo 获取群信息
func (c *Cache) GetGroupInfo(groupUin uint32) *entity2.Group {
	v, _ := getCacheOf[entity2.Group](c, groupUin)
	return v
}

// GetAllGroupsInfo 获取所有群信息
func (c *Cache) GetAllGroupsInfo() map[uint32]*entity2.Group {
	groups := make(map[uint32]*entity2.Group, 64)
	rangeCacheOf[entity2.Group](c, func(k uint32, v *entity2.Group) bool {
		groups[k] = v
		return true
	})
	return groups
}

// GetGroupMember 获取群成员信息
func (c *Cache) GetGroupMember(uin, groupUin uint32) *entity2.GroupMember {
	if group, ok := getCacheOf[Cache](c, groupUin); ok {
		v, _ := getCacheOf[entity2.GroupMember](group, uin)
		return v
	}
	return nil
}

// GetGroupMembers 获取指定群所有群成员信息
func (c *Cache) GetGroupMembers(groupUin uint32) map[uint32]*entity2.GroupMember {
	members := make(map[uint32]*entity2.GroupMember, 64)
	if group, ok := getCacheOf[Cache](c, groupUin); ok {
		rangeCacheOf[entity2.GroupMember](group, func(k uint32, member *entity2.GroupMember) bool {
			members[member.Uin] = member
			return true
		})
	}
	return members
}

// FriendCacheIsEmpty 好友信息缓存是否为空
func (c *Cache) FriendCacheIsEmpty() bool {
	return !hasRefreshed[entity2.Friend](c)
}

// GroupMembersCacheIsEmpty 群成员缓存是否为空
func (c *Cache) GroupMembersCacheIsEmpty() bool {
	return !hasRefreshed[Cache](c)
}

// GroupMemberCacheIsEmpty 指定群的群成员缓存是否为空
func (c *Cache) GroupMemberCacheIsEmpty(groupUin uint32) bool {
	if group, ok := getCacheOf[Cache](c, groupUin); ok {
		return !hasRefreshed[entity2.GroupMember](group)
	}
	return true
}

// GroupInfoCacheIsEmpty 群信息缓存是否为空
func (c *Cache) GroupInfoCacheIsEmpty() bool {
	return !hasRefreshed[entity2.Group](c)
}
