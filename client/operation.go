package client

import (
	"github.com/LagrangeDev/LagrangeGo/entity"
	"github.com/LagrangeDev/LagrangeGo/packets/oidb"
)

// FetchFriends 获取好友列表信息
func (c *QQClient) FetchFriends() ([]*entity.Friend, error) {
	pkt, err := oidb.BuildFetchFriendsReq()
	if err != nil {
		return nil, err
	}
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	friends, err := oidb.ParseFetchFriendsResp(resp.Data)
	if err != nil {
		return nil, err
	}
	return friends, nil
}

// FetchGroups 获取所有已加入的群的信息
func (c *QQClient) FetchGroups() ([]*entity.Group, error) {
	pkt, err := oidb.BuildFetchGroupsReq()
	if err != nil {
		return nil, err
	}
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	groups, err := oidb.ParseFetchGroupsResp(resp.Data)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// FetchGroupMember 获取对应群的群成员信息，使用token可以获取下一页的群成员信息
func (c *QQClient) FetchGroupMember(groupID uint32, token string) ([]*entity.GroupMember, string, error) {
	pkt, err := oidb.BuildFetchMembersReq(groupID, token)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, "", err
	}
	members, newToken, err := oidb.ParseFetchMembersResp(resp.Data)
	if err != nil {
		return nil, "", err
	}
	return members, newToken, nil
}
