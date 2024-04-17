package cache

import (
	"github.com/LagrangeDev/LagrangeGo/entity"
)

func (c *Cache) RefreshAll(friendCache map[uint32]*entity.Friend, groupCache map[uint32]*entity.Group, groupMemberCache map[uint32]map[uint32]*entity.GroupMember) {
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.FriendCache = friendCache
	c.GroupMemberCache = groupMemberCache
	c.GroupCache = groupCache
}

// RefreshFriend 刷新一个好友的缓存
func (c *Cache) RefreshFriend(friend *entity.Friend) {
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.FriendCache[friend.Uin] = friend
}

// RefreshAllFriend 刷新所有好友缓存
func (c *Cache) RefreshAllFriend(friendCache map[uint32]*entity.Friend) {
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.FriendCache = friendCache
}

// RefreshGroupMember 刷新指定群的一个群成员缓存
func (c *Cache) RefreshGroupMember(groupUin uint32, groupMember *entity.GroupMember) {
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.GroupMemberCache[groupUin][groupMember.Uin] = groupMember
}

// RefreshGroupMembers 刷新一个群内的所有群成员缓存
func (c *Cache) RefreshGroupMembers(groupUin uint32, groupMembers map[uint32]*entity.GroupMember) {
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.GroupMemberCache[groupUin] = groupMembers
}

// RefreshAllGroupMembers 刷新所有群的群员缓存
func (c *Cache) RefreshAllGroupMembers(groupMemberCache map[uint32]map[uint32]*entity.GroupMember) {
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.GroupMemberCache = groupMemberCache
}

// RefreshGroup 刷新一个群的群信息缓存
func (c *Cache) RefreshGroup(group *entity.Group) {
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.GroupCache[group.GroupUin] = group
}

// RefreshAllGroup 刷新所有群的群信息缓存
func (c *Cache) RefreshAllGroup(groupCache map[uint32]*entity.Group) {
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.GroupCache = groupCache
}
