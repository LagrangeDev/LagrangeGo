package client

import (
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/entity"
)

// GetUID 获取缓存中对应uin的uid
func (c *QQClient) GetUID(uin uint32, groupUin ...uint32) string {
	if len(groupUin) == 0 && c.cache.FriendCacheIsEmpty() {
		if err := c.RefreshFriendCache(); err != nil {
			return ""
		}
	} else if len(groupUin) != 0 && c.cache.GroupMemberCacheIsEmpty(groupUin[0]) {
		if err := c.RefreshGroupMembersCache(groupUin[0]); err != nil {
			return ""
		}
	}
	return c.cache.GetUID(uin, groupUin...)
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
func (c *QQClient) GetCachedFriendInfo(uin uint32) *entity.Friend {
	if c.cache.FriendCacheIsEmpty() {
		if err := c.RefreshFriendCache(); err != nil {
			return nil
		}
	}
	return c.cache.GetFriend(uin)
}

// GetCachedAllFriendsInfo 获取所有好友信息(缓存)
func (c *QQClient) GetCachedAllFriendsInfo() map[uint32]*entity.Friend {
	if c.cache.FriendCacheIsEmpty() {
		if err := c.RefreshFriendCache(); err != nil {
			return nil
		}
	}
	return c.cache.GetAllFriends()
}

// GetCachedGroupInfo 获取群信息(缓存)
func (c *QQClient) GetCachedGroupInfo(groupUin uint32) *entity.Group {
	if c.cache.GroupInfoCacheIsEmpty() {
		if err := c.RefreshAllGroupsInfo(); err != nil {
			return nil
		}
	}
	return c.cache.GetGroupInfo(groupUin)
}

// GetCachedAllGroupsInfo 获取所有群信息(缓存)
func (c *QQClient) GetCachedAllGroupsInfo() map[uint32]*entity.Group {
	if c.cache.GroupInfoCacheIsEmpty() {
		if err := c.RefreshAllGroupsInfo(); err != nil {
			return nil
		}
	}
	return c.cache.GetAllGroupsInfo()
}

// GetCachedMemberInfo 获取群成员信息(缓存)
func (c *QQClient) GetCachedMemberInfo(uin, groupUin uint32) *entity.GroupMember {
	if c.cache.GroupMemberCacheIsEmpty(groupUin) {
		if err := c.RefreshGroupMemberCache(groupUin, uin); err != nil {
			return nil
		}
	}
	return c.cache.GetGroupMember(uin, groupUin)
}

// GetCachedMembersInfo 获取指定群所有群成员信息(缓存)
func (c *QQClient) GetCachedMembersInfo(groupUin uint32) map[uint32]*entity.GroupMember {
	if c.cache.GroupMemberCacheIsEmpty(groupUin) {
		if err := c.RefreshGroupMembersCache(groupUin); err != nil {
			return nil
		}
	}
	return c.cache.GetGroupMembers(groupUin)
}

// GetCachedRkeyInfo 获取指定类型的RKey信息(缓存)
func (c *QQClient) GetCachedRkeyInfo(rkeyType entity.RKeyType) *entity.RKeyInfo {
	refresh := c.cache.RkeyInfoCacheIsEmpty()
	for {
		if refresh {
			if err := c.RefreshAllRkeyInfoCache(); err != nil {
				return nil
			}
		}
		inf := c.cache.GetRKeyInfo(rkeyType)
		if inf.ExpireTime <= uint64(time.Now().Unix()) {
			refresh = true
			continue
		}
		return inf
	}
}

// GetCachedRkeyInfos 获取所有RKey信息(缓存)
func (c *QQClient) GetCachedRkeyInfos() map[entity.RKeyType]*entity.RKeyInfo {
	refresh := c.cache.RkeyInfoCacheIsEmpty()
	for {
		if refresh {
			if err := c.RefreshAllRkeyInfoCache(); err != nil {
				return nil
			}
			refresh = false
		}
		inf := c.cache.GetAllRkeyInfo()
		for _, v := range inf {
			if v.ExpireTime <= uint64(time.Now().Unix()) {
				refresh = true
				break
			}
		}
		if refresh {
			continue
		}
		return inf
	}
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

// RefreshGroupMemberCache 刷新一个群的指定群成员缓存
func (c *QQClient) RefreshGroupMemberCache(groupUin, memberUin uint32) error {
	member, err := c.FetchGroupMember(groupUin, memberUin)
	if err != nil {
		return err
	}
	c.cache.RefreshGroupMember(groupUin, member)
	return nil
}

// RefreshGroupMembersCache 刷新指定群的所有群成员缓存
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

// RefreshAllRkeyInfoCache 刷新RKey缓存
func (c *QQClient) RefreshAllRkeyInfoCache() error {
	rkeyInfo, err := c.FetchRkey()
	if err != nil {
		return err
	}
	c.cache.RefreshAllRKeyInfo(rkeyInfo)
	return nil
}

// GetFriendsData 获取好友列表数据
func (c *QQClient) GetFriendsData() (map[uint32]*entity.Friend, error) {
	friendsData := make(map[uint32]*entity.Friend)
	friends, token, err := c.FetchFriends(0)
	if err != nil {
		return friendsData, err
	}
	for _, friend := range friends {
		friendsData[friend.Uin] = friend
	}
	for token != 0 {
		friends, token, err = c.FetchFriends(token)
		if err != nil {
			return friendsData, err
		}
		for _, friend := range friends {
			friendsData[friend.Uin] = friend
		}
	}
	c.debug("获取%d个好友", len(friendsData))
	return friendsData, err
}

// GetGroupMembersData 获取指定群所有成员信息
func (c *QQClient) GetGroupMembersData(groupUin uint32) (map[uint32]*entity.GroupMember, error) {
	groupMembers := make(map[uint32]*entity.GroupMember)
	members, token, err := c.FetchGroupMembers(groupUin, "")
	if err != nil {
		return groupMembers, err
	}
	for _, member := range members {
		groupMembers[member.Uin] = member
	}
	for token != "" {
		members, token, err = c.FetchGroupMembers(groupUin, token)
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
	c.debug("获取%d个群的成员信息", len(groupsData))
	return groupsData, err
}

func (c *QQClient) GetAllGroupsInfo() (map[uint32]*entity.Group, error) {
	groupsInfo, err := c.FetchGroups()
	if err != nil {
		return nil, err
	}
	groupsData := make(map[uint32]*entity.Group, len(groupsInfo))
	for _, group := range groupsInfo {
		groupsData[group.GroupUin] = group
	}
	c.debug("获取%d个群信息", len(groupsData))
	return groupsData, err
}
