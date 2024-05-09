package client

import (
	entity2 "github.com/LagrangeDev/LagrangeGo/internal/entity"
)

// GetUid 获取缓存中对应uin的uid
func (c *QQClient) GetUid(uin uint32, groupUin ...uint32) string {
	if c.cache.FriendCacheIsEmpty() {
		if err := c.RefreshFriendCache(); err != nil {
			return ""
		}
	}
	return c.cache.GetUid(uin, groupUin...)
}

// GetUin 获取缓存中对应的uin
func (c *QQClient) GetUin(uid string, groupUin ...uint32) uint32 {
	if len(groupUin) == 0 && c.cache.FriendCacheIsEmpty() {
		if err := c.RefreshFriendCache(); err != nil {
			return 0
		}
	}
	if len(groupUin) != 0 && c.cache.GroupMemberCacheIsEmpty(groupUin[0]) {
		if err := c.RefreshGroupMembersCache(groupUin[0]); err != nil {
			return 0
		}
	}
	return c.cache.GetUin(uid, groupUin...)
}

// GetCachedFriendInfo 获取好友信息(缓存)
func (c *QQClient) GetCachedFriendInfo(uin uint32) (*entity2.Friend, error) {
	if c.cache.FriendCacheIsEmpty() {
		if err := c.RefreshFriendCache(); err != nil {
			return nil, err
		}
	}
	return c.cache.GetFriend(uin), nil
}

// GetCachedGroupInfo 获取群信息(缓存)
func (c *QQClient) GetCachedGroupInfo(groupUin uint32) (*entity2.Group, error) {
	if c.cache.GroupInfoCacheIsEmpty() {
		if err := c.RefreshAllGroupsInfo(); err != nil {
			return nil, err
		}
	}
	return c.cache.GetGroupInfo(groupUin), nil
}

// GetCachedAllGroupsInfo 获取所有群信息(缓存)
func (c *QQClient) GetCachedAllGroupsInfo() (map[uint32]*entity2.Group, error) {
	if c.cache.GroupInfoCacheIsEmpty() {
		if err := c.RefreshAllGroupsInfo(); err != nil {
			return nil, err
		}
	}
	return c.cache.GetAllGroupsInfo(), nil
}

// GetCachedMemberInfo 获取群成员信息(缓存)
func (c *QQClient) GetCachedMemberInfo(uin, groupUin uint32) (*entity2.GroupMember, error) {
	if c.cache.GroupMemberCacheIsEmpty(groupUin) {
		if err := c.RefreshGroupMembersCache(groupUin); err != nil {
			return nil, err
		}
	}
	return c.cache.GetGroupMember(uin, groupUin), nil
}

// GetCachedMembersInfo 获取指定群所有群成员信息(缓存)
func (c *QQClient) GetCachedMembersInfo(groupUin uint32) (map[uint32]*entity2.GroupMember, error) {
	if c.cache.GroupMemberCacheIsEmpty(groupUin) {
		if err := c.RefreshGroupMembersCache(groupUin); err != nil {
			return nil, err
		}
	}
	return c.cache.GetGroupMembers(groupUin), nil
}

// RefreshFriendCache 刷新好友缓存
func (c *QQClient) RefreshFriendCache() error {
	friendsData, err := c.GetFriendsData()
	if err != nil {
		return err
	}
	c.cache.RefreshAllFriend(friendsData)
	return nil
}

// RefreshGroupMembersCache 刷新指定群的群成员员缓存
func (c *QQClient) RefreshGroupMembersCache(groupUin uint32) error {
	groupData, err := c.GetGroupMembersData(groupUin)
	if err != nil {
		return err
	}
	c.cache.RefreshGroupMembers(groupUin, groupData)
	return nil
}

// RefreshAllGroupMembersCache 刷新所有群的群成员缓存
func (c *QQClient) RefreshAllGroupMembersCache() error {
	groupsData, err := c.GetAllGroupsMembersData()
	if err != nil {
		return err
	}
	c.cache.RefreshAllGroupMembers(groupsData)
	return nil
}

// RefreshAllGroupsInfo 刷新所有群信息缓存
func (c *QQClient) RefreshAllGroupsInfo() error {
	groupsData, err := c.GetAllGroupsInfo()
	if err != nil {
		return err
	}
	c.cache.RefreshAllGroup(groupsData)
	return nil
}

// GetFriendsData 获取好友列表数据
func (c *QQClient) GetFriendsData() (map[uint32]*entity2.Friend, error) {
	friends, err := c.FetchFriends()
	if err != nil {
		return nil, err
	}
	friendsData := make(map[uint32]*entity2.Friend, len(friends))
	for _, friend := range friends {
		friendsData[friend.Uin] = friend
	}
	loginLogger.Infof("获取%d个好友", len(friendsData))
	return friendsData, err
}

// GetGroupMembersData 获取指定群所有成员信息
func (c *QQClient) GetGroupMembersData(groupUin uint32) (map[uint32]*entity2.GroupMember, error) {
	groupMembers := make(map[uint32]*entity2.GroupMember)
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
func (c *QQClient) GetAllGroupsMembersData() (map[uint32]map[uint32]*entity2.GroupMember, error) {
	groups, err := c.FetchGroups()
	if err != nil {
		return nil, err
	}
	groupsData := make(map[uint32]map[uint32]*entity2.GroupMember, len(groups))
	for _, group := range groups {
		groupMembersData, err := c.GetGroupMembersData(group.GroupUin)
		if err != nil {
			return nil, err
		}
		groupsData[group.GroupUin] = groupMembersData
	}
	loginLogger.Infof("获取%d个群的成员信息", len(groupsData))
	return groupsData, err
}

func (c *QQClient) GetAllGroupsInfo() (map[uint32]*entity2.Group, error) {
	groupsInfo, err := c.FetchGroups()
	if err != nil {
		return nil, err
	}
	groupsData := make(map[uint32]*entity2.Group, len(groupsInfo))
	for _, group := range groupsInfo {
		groupsData[group.GroupUin] = group
	}
	loginLogger.Infof("获取%d个群信息", len(groupsData))
	return groupsData, err
}
