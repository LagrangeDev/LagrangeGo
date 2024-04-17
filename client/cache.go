package client

import (
	"github.com/LagrangeDev/LagrangeGo/entity"
)

// GetUid 获取缓存中对应uin的uid
func (c *QQClient) GetUid(uin uint32, groupUin ...uint32) string {
	if c.cache.FriendCacheIsEmpty() {
		c.RefreshFriendCache()
	}
	return c.cache.GetUid(uin, groupUin...)
}

// GetUin 获取缓存中对应的uin
func (c *QQClient) GetUin(uid string, groupUin ...uint32) uint32 {
	if c.cache.FriendCacheIsEmpty() {
		c.RefreshFriendCache()
	}
	if len(groupUin) != 0 && c.cache.GroupMemberCacheIsEmpty(groupUin[0]) {
		c.RefreshGroupMembersCache(groupUin[0])
	}
	return c.cache.GetUin(uid, groupUin...)
}

// GetFriendInfo 获取好友信息
func (c *QQClient) GetFriendInfo(uin uint32) *entity.Friend {
	if c.cache.FriendCacheIsEmpty() {
		c.RefreshFriendCache()
	}
	return c.cache.GetFriend(uin)
}

// GetMemberInfo 获取群成员信息
func (c *QQClient) GetMemberInfo(uin, groupUin uint32) *entity.GroupMember {
	if c.cache.GroupMemberCacheIsEmpty(groupUin) {
		c.RefreshGroupMembersCache(groupUin)
	}
	return c.cache.GetGroupMember(uin, groupUin)
}

func (c *QQClient) GetMembersInfo(groupUin uint32) map[uint32]*entity.GroupMember {
	if c.cache.GroupMemberCacheIsEmpty(groupUin) {
		c.RefreshGroupMembersCache(groupUin)
	}
	return c.cache.GetGroupMembers(groupUin)
}

// RefreshFriendCache 刷新好友缓存
func (c *QQClient) RefreshFriendCache() {
	friendsData, err := c.GetFriendsData()
	if err != nil {
		return
	}
	c.cache.RefreshAllFriend(friendsData)
}

// RefreshGroupMembersCache 刷新指定群的群成员员缓存
func (c *QQClient) RefreshGroupMembersCache(groupUin uint32) {
	groupData, err := c.GetGroupMembersData(groupUin)
	if err != nil {
		return
	}
	c.cache.RefreshGroupMembers(groupUin, groupData)
}

// RefreshAllGroupCache 刷新所有群的群成员缓存
func (c *QQClient) RefreshAllGroupCache() {
	groupsData, err := c.GetAllGroupsMembersData()
	if err != nil {
		return
	}
	c.cache.RefreshAllGroupMembers(groupsData)
}

// GetFriendsData 获取好友列表数据
func (c *QQClient) GetFriendsData() (map[uint32]*entity.Friend, error) {
	friends, err := c.FetchFriends()
	if err != nil {
		return nil, err
	}
	friendsData := make(map[uint32]*entity.Friend, len(friends))
	for _, friend := range friends {
		friendsData[friend.Uin] = friend
	}
	loginLogger.Infof("获取%d个好友", len(friendsData))
	return friendsData, err
}

// GetGroupMembersData 获取指定群所有成员信息
func (c *QQClient) GetGroupMembersData(groupUin uint32) (map[uint32]*entity.GroupMember, error) {
	groupMembers := make(map[uint32]*entity.GroupMember)
	members, token, err := c.FetchGroupMember(groupUin, "")
	if err != nil {
		return groupMembers, err
	}
	for _, member := range members {
		groupMembers[member.Uin] = member
	}
	for token != "" {
		members, token, err = c.FetchGroupMember(groupUin, token)
		if err != nil {
			return groupMembers, err
		}
		for _, member := range members {
			groupMembers[member.Uin] = member
		}
	}
	return groupMembers, err
}

// GetAllGroupsMembersData 获取所有群的群成员信息
func (c *QQClient) GetAllGroupsMembersData() (map[uint32]map[uint32]*entity.GroupMember, error) {
	groups, err := c.FetchGroups()
	if err != nil {
		return nil, err
	}
	groupsData := make(map[uint32]map[uint32]*entity.GroupMember, len(groups))
	for _, group := range groups {
		groupMembersData, err := c.GetGroupMembersData(group.GroupUin)
		if err != nil {
			return nil, err
		}
		groupsData[group.GroupUin] = groupMembersData
	}
	loginLogger.Infof("获取%d个群和成员信息", len(groupsData))
	return groupsData, err
}
