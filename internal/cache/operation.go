package cache

import (
	entity2 "github.com/LagrangeDev/LagrangeGo/internal/entity"
)

func (c *Cache) RefreshAll(friendCache map[uint32]*entity2.Friend, groupCache map[uint32]*entity2.Group, groupMemberCache map[uint32]map[uint32]*entity2.GroupMember) {
	c.RefreshAllFriend(friendCache)
	c.RefreshAllGroup(groupCache)
	c.RefreshAllGroupMembers(groupMemberCache)
}

// RefreshFriend 刷新一个好友的缓存
func (c *Cache) RefreshFriend(friend *entity2.Friend) {
	setCacheOf(c, friend.Uin, friend)
}

// RefreshAllFriend 刷新所有好友缓存
func (c *Cache) RefreshAllFriend(friendCache map[uint32]*entity2.Friend) {
	refreshAllCacheOf(c, friendCache)
}

// RefreshGroupMember 刷新指定群的一个群成员缓存
func (c *Cache) RefreshGroupMember(groupUin uint32, groupMember *entity2.GroupMember) {
	group, ok := getCacheOf[Cache](c, groupUin)
	if !ok {
		group = &Cache{}
		setCacheOf(c, groupUin, group)
	}
	setCacheOf(group, groupMember.Uin, groupMember)
}

// RefreshGroupMembers 刷新一个群内的所有群成员缓存
func (c *Cache) RefreshGroupMembers(groupUin uint32, groupMembers map[uint32]*entity2.GroupMember) {
	newc := &Cache{}
	for k, v := range groupMembers {
		setCacheOf(newc, k, v)
	}
	setCacheOf(c, groupUin, newc)
}

// RefreshAllGroupMembers 刷新所有群的群员缓存
func (c *Cache) RefreshAllGroupMembers(groupMemberCache map[uint32]map[uint32]*entity2.GroupMember) {
	newc := make(map[uint32]*Cache, len(groupMemberCache)*2)
	for groupUin, v := range groupMemberCache {
		group := &Cache{}
		refreshAllCacheOf(group, v)
		newc[groupUin] = group
	}
	refreshAllCacheOf(c, newc)
}

// RefreshGroup 刷新一个群的群信息缓存
func (c *Cache) RefreshGroup(group *entity2.Group) {
	setCacheOf(c, group.GroupUin, group)
}

// RefreshAllGroup 刷新所有群的群信息缓存
func (c *Cache) RefreshAllGroup(groupCache map[uint32]*entity2.Group) {
	refreshAllCacheOf(c, groupCache)
}
