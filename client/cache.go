package client

import (
	"github.com/LagrangeDev/LagrangeGo/entity"
)

// GetUidFromFriends 获取缓存中对应qq的uid，仅限好友
func (c *QQClient) GetUidFromFriends(uin uint32) string {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	if friend, ok := c.friendCache[uin]; ok {
		return friend.Uid
	}
	return ""
}

func (c *QQClient) GetUidFromGroup(uin, groupUin uint32) string {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	if group, ok := c.groupCache[groupUin]; ok {
		if member, ok := group[uin]; ok {
			return member.Uid
		}
	}
	return ""
}

func (c *QQClient) GetMemberInfo(uin, groupUin uint32) *entity.GroupMember {
	c.refreshLock.RLock()
	defer c.refreshLock.RUnlock()
	if group, ok := c.groupCache[groupUin]; ok {
		if member, ok := group[uin]; ok {
			return member
		}
	}
	return nil
}

func (c *QQClient) RefreshFriendCache() {
	friendsData, err := c.GetFriendsData()
	if err != nil {
		return
	}
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.friendCache = friendsData
}

func (c *QQClient) RefreshGroupCache(groupUin uint32) {
	groupData, err := c.GetGroupMembersData(groupUin)
	if err != nil {
		return
	}
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.groupCache[groupUin] = groupData
}

func (c *QQClient) RefreshAllGroupCache() {
	groupsData, err := c.GetAllGroupsMembersData()
	if err != nil {
		return
	}
	c.refreshLock.Lock()
	defer c.refreshLock.Unlock()
	c.groupCache = groupsData
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
