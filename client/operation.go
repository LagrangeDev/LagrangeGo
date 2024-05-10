package client

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"

	"github.com/LagrangeDev/LagrangeGo/entity"
	"github.com/LagrangeDev/LagrangeGo/packets/oidb"
)

// FetchFriends 获取好友列表信息
func (c *QQClient) FetchFriends() ([]*entity.Friend, error) {
	pkt, err := oidb.BuildFetchFriendsReq()
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	friends, err := oidb.ParseFetchFriendsResp(resp)
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
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	groups, err := oidb.ParseFetchGroupsResp(resp)
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
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, "", err
	}
	members, newToken, err := oidb.ParseFetchMembersResp(resp)
	if err != nil {
		return nil, "", err
	}
	return members, newToken, nil
}

func (c *QQClient) GroupRemark(groupID uint32, remark string) error {
	pkt, err := oidb.BuildGroupRemarkReq(groupID, remark)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb.ParseGroupRemarkResp(resp)
}

func (c *QQClient) GroupRename(groupID uint32, name string) error {
	pkt, err := oidb.BuildGroupRenameReq(groupID, name)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb.ParseGroupRenameResp(resp)
}

func (c *QQClient) GroupMuteGlobal(groupID uint32, isMute bool) error {
	pkt, err := oidb.BuildGroupMuteGlobalReq(groupID, isMute)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb.ParseGroupMuteGlobalResp(resp)
}

func (c *QQClient) GroupMuteMember(groupID, duration, uin uint32) error {
	uid := c.GetUid(uin, groupID)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb.BuildGroupMuteMemberReq(groupID, duration, uid)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb.ParseGroupMuteMemberResp(resp)
}

func (c *QQClient) GroupLeave(groupID uint32) error {
	pkt, err := oidb.BuildGroupLeaveReq(groupID)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb.ParseGroupLeaveResp(resp)
}

func (c *QQClient) GroupSetAdmin(groupID, uin uint32, isAdmin bool) error {
	uid := c.GetUid(uin, groupID)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb.BuildGroupSetAdminReq(groupID, uid, isAdmin)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	err = oidb.ParseGroupSetAdminResp(resp)
	if err != nil {
		return err
	}
	if m, _ := c.GetCachedMemberInfo(uin, groupID); m != nil {
		m.Permission = entity.Admin
		c.cache.RefreshGroupMember(groupID, m)
	}

	return nil
}

func (c *QQClient) GroupRenameMember(groupID, uin uint32, name string) error {
	uid := c.GetUid(uin, groupID)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb.BuildGroupRenameMemberReq(groupID, uid, name)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	err = oidb.ParseGroupRenameMemberResp(resp)
	if err != nil {
		return err
	}
	if m, _ := c.GetCachedMemberInfo(uin, groupID); m != nil {
		m.MemberCard = name
		c.cache.RefreshGroupMember(groupID, m)
	}

	return nil
}

func (c *QQClient) GroupKickMember(groupID, uin uint32, rejectAddRequest bool) error {
	uid := c.GetUid(uin, groupID)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb.BuildGroupKickMemberReq(groupID, uid, rejectAddRequest)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb.ParseGroupKickMemberResp(resp)
}

func (c *QQClient) GroupSetSpecialTitle(groupUin, uin uint32, title string) error {
	uid := c.GetUid(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb.BuildGroupSetSpecialTitleReq(groupUin, uid, title)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb.ParseGroupSetSpecialTitleResp(resp)
}

func (c *QQClient) GroupPoke(groupID, uin uint32) error {
	pkt, err := oidb.BuildGroupPokeReq(groupID, uin)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb.ParsePokeResp(resp)
}

func (c *QQClient) FriendPoke(uin uint32) error {
	pkt, err := oidb.BuildFriendPokeReq(uin)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb.ParsePokeResp(resp)
}

func (c *QQClient) RecallGroupMessage(GrpUin, seq uint32) error {
	packet := message.GroupRecallMsg{
		Type:     1,
		GroupUin: GrpUin,
		Field3: &message.GroupRecallMsgField3{
			Sequence: seq,
			Field3:   0,
		},
		Field4: &message.GroupRecallMsgField4{Field1: 0},
	}
	pktData, err := proto.Marshal(&packet)
	if err != nil {
		return err
	}
	resp, err := c.sendUniPacketAndWait("trpc.msg.msg_svc.MsgService.SsoGroupRecallMsg", pktData)
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return errors.New("empty response data")
	}
	return nil
}
