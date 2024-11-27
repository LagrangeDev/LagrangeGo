package cache

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
)

func (c *Cache) RefreshAll(
	friendCache map[uint32]*entity.User,
	groupCache map[uint32]*entity.Group,
	groupMemberCache map[uint32]map[uint32]*entity.GroupMember,
	rkeyCache entity.RKeyMap,
) {
	c.RefreshAllFriend(friendCache)
	c.RefreshAllGroup(groupCache)
	c.RefreshAllGroupMembers(groupMemberCache)
	c.RefreshAllRKeyInfo(rkeyCache)
}

// RefreshFriend 刷新一个好友的缓存
func (c *Cache) RefreshFriend(friend *entity.User) {
	setCacheOf(c, friend.Uin, friend)
}

// RefreshAllFriend 刷新所有好友缓存
func (c *Cache) RefreshAllFriend(friendCache map[uint32]*entity.User) {
	refreshAllCacheOf(c, friendCache)
}

// RefreshGroupMember 刷新指定群的一个群成员缓存
func (c *Cache) RefreshGroupMember(groupUin uint32, groupMember *entity.GroupMember) {
	group, ok := getCacheOf[Cache](c, groupUin)
	if !ok {
		group = &Cache{}
		setCacheOf(c, groupUin, group)
	}
	setCacheOf(group, groupMember.Uin, groupMember)
}

// RefreshGroupMembers 刷新一个群内的所有群成员缓存
func (c *Cache) RefreshGroupMembers(groupUin uint32, groupMembers map[uint32]*entity.GroupMember) {
	newc := &Cache{}
	for k, v := range groupMembers {
		setCacheOf(newc, k, v)
	}
	setCacheOf(c, groupUin, newc)
}

// RefreshAllGroupMembers 刷新所有群的群员缓存
func (c *Cache) RefreshAllGroupMembers(groupMemberCache map[uint32]map[uint32]*entity.GroupMember) {
	newc := make(map[uint32]*Cache, len(groupMemberCache)*2)
	for groupUin, v := range groupMemberCache {
		group := &Cache{}
		refreshAllCacheOf(group, v)
		newc[groupUin] = group
	}
	refreshAllCacheOf(c, newc)
}

// RefreshGroup 刷新一个群的群信息缓存
func (c *Cache) RefreshGroup(group *entity.Group) {
	setCacheOf(c, group.GroupUin, group)
}

// RefreshAllGroup 刷新所有群的群信息缓存
func (c *Cache) RefreshAllGroup(groupCache map[uint32]*entity.Group) {
	refreshAllCacheOf(c, groupCache)
}

// RefreshAllRKeyInfo 刷新所有RKey缓存
func (c *Cache) RefreshAllRKeyInfo(rkeyCache entity.RKeyMap) {
	refreshAllCacheOf(c, rkeyCache)
}
