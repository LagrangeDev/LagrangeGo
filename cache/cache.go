package cache

import "github.com/LagrangeDev/LagrangeGo/entity"

// GetUid 根据uin获取uid
func (c *Cache) GetUid(uin uint32, groupUin ...uint32) string {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	if len(groupUin) == 0 {
		if friend, ok := c.FriendCache[uin]; ok {
			return friend.Uid
		}
	} else {
		if group, ok := c.GroupMemberCache[groupUin[0]]; ok {
			if member, ok := group[uin]; ok {
				return member.Uid
			}
		}
	}
	return ""
}

// GetUin 根据uid获取uin
func (c *Cache) GetUin(uid string, groupUin ...uint32) uint32 {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	if len(groupUin) == 0 {
		for uin, friend := range c.FriendCache {
			if friend.Uid == uid {
				return uin
			}
		}
	} else {
		if group, ok := c.GroupMemberCache[groupUin[0]]; ok {
			for uin, member := range group {
				if member.Uid == uid {
					return uin
				}
			}
		}
	}
	return 0
}

// GetFriend 获取好友信息
func (c *Cache) GetFriend(uin uint32) *entity.Friend {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	return c.FriendCache[uin]
}

// GetGroupInfo 获取群信息
func (c *Cache) GetGroupInfo(groupUin uint32) *entity.Group {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	return c.GroupInfoCache[groupUin]
}

// GetAllGroupsInfo 获取所有群信息
func (c *Cache) GetAllGroupsInfo() map[uint32]*entity.Group {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	groups := make(map[uint32]*entity.Group, len(c.GroupInfoCache))
	for group, grpInfo := range c.GroupInfoCache {
		groups[group] = grpInfo
	}
	return groups
}

// GetGroupMember 获取群成员信息
func (c *Cache) GetGroupMember(uin, groupUin uint32) *entity.GroupMember {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	if group, ok := c.GroupMemberCache[groupUin]; ok {
		return group[uin]
	}
	return nil
}

// GetGroupMembers 获取指定群所有群成员信息
func (c *Cache) GetGroupMembers(groupUin uint32) map[uint32]*entity.GroupMember {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	members := make(map[uint32]*entity.GroupMember, len(c.GroupMemberCache))
	for _, member := range c.GroupMemberCache[groupUin] {
		members[member.Uin] = member
	}
	return members
}

// FriendCacheIsEmpty 好友信息缓存是否为空
func (c *Cache) FriendCacheIsEmpty() bool {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	return len(c.FriendCache) == 0
}

// GroupMembersCacheIsEmpty 群成员缓存是否为空
func (c *Cache) GroupMembersCacheIsEmpty() bool {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	return len(c.GroupMemberCache) == 0
}

// GroupMemberCacheIsEmpty 指定群的群成员缓存是否为空
func (c *Cache) GroupMemberCacheIsEmpty(groupUin uint32) bool {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	return len(c.GroupMemberCache[groupUin]) == 0
}

// GroupInfoCacheIsEmpty 群信息缓存是否为空
func (c *Cache) GroupInfoCacheIsEmpty() bool {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	return len(c.GroupInfoCache) == 0
}
