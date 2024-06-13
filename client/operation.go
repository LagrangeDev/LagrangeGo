package client

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"

	"github.com/LagrangeDev/LagrangeGo/client/entity"
	oidb2 "github.com/LagrangeDev/LagrangeGo/client/packets/oidb"
)

// FetchFriends 获取好友列表信息
func (c *QQClient) FetchFriends() ([]*entity.Friend, error) {
	pkt, err := oidb2.BuildFetchFriendsReq()
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	friends, err := oidb2.ParseFetchFriendsResp(resp)
	if err != nil {
		return nil, err
	}
	return friends, nil
}

// FetchGroups 获取所有已加入的群的信息
func (c *QQClient) FetchGroups() ([]*entity.Group, error) {
	pkt, err := oidb2.BuildFetchGroupsReq()
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	groups, err := oidb2.ParseFetchGroupsResp(resp)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// FetchGroupMember 获取对应群的群成员信息，使用token可以获取下一页的群成员信息
func (c *QQClient) FetchGroupMember(groupID uint32, token string) ([]*entity.GroupMember, string, error) {
	pkt, err := oidb2.BuildFetchMembersReq(groupID, token)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, "", err
	}
	members, newToken, err := oidb2.ParseFetchMembersResp(resp)
	if err != nil {
		return nil, "", err
	}
	return members, newToken, nil
}

func (c *QQClient) GroupRemark(groupID uint32, remark string) error {
	pkt, err := oidb2.BuildGroupRemarkReq(groupID, remark)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupRemarkResp(resp)
}

func (c *QQClient) GroupRename(groupID uint32, name string) error {
	pkt, err := oidb2.BuildGroupRenameReq(groupID, name)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupRenameResp(resp)
}

func (c *QQClient) GroupMuteGlobal(groupID uint32, isMute bool) error {
	pkt, err := oidb2.BuildGroupMuteGlobalReq(groupID, isMute)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupMuteGlobalResp(resp)
}

func (c *QQClient) GroupMuteMember(groupID, duration, uin uint32) error {
	uid := c.GetUid(uin, groupID)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildGroupMuteMemberReq(groupID, duration, uid)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupMuteMemberResp(resp)
}

func (c *QQClient) GroupLeave(groupID uint32) error {
	pkt, err := oidb2.BuildGroupLeaveReq(groupID)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupLeaveResp(resp)
}

func (c *QQClient) GroupSetAdmin(groupID, uin uint32, isAdmin bool) error {
	uid := c.GetUid(uin, groupID)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildGroupSetAdminReq(groupID, uid, isAdmin)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	err = oidb2.ParseGroupSetAdminResp(resp)
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
	pkt, err := oidb2.BuildGroupRenameMemberReq(groupID, uid, name)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	err = oidb2.ParseGroupRenameMemberResp(resp)
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
	pkt, err := oidb2.BuildGroupKickMemberReq(groupID, uid, rejectAddRequest)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupKickMemberResp(resp)
}

func (c *QQClient) GroupSetSpecialTitle(groupUin, uin uint32, title string) error {
	uid := c.GetUid(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildGroupSetSpecialTitleReq(groupUin, uid, title)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupSetSpecialTitleResp(resp)
}

func (c *QQClient) GroupPoke(groupID, uin uint32) error {
	pkt, err := oidb2.BuildGroupPokeReq(groupID, uin)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParsePokeResp(resp)
}

func (c *QQClient) FriendPoke(uin uint32) error {
	pkt, err := oidb2.BuildFriendPokeReq(uin)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParsePokeResp(resp)
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

func (c *QQClient) GetRecordUrl(node *oidb.IndexNode) (string, error) {
	pkt, err := oidb2.BuildRecordDownloadReq(c.GetUid(c.Uin), node)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParseRecordDownloadResp(resp)
}

func (c *QQClient) GetGroupRecordUrl(groupUin uint32, node *oidb.IndexNode) (string, error) {
	pkt, err := oidb2.BuildGroupRecordDownloadReq(groupUin, node)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParseRecordGroupDownloadResp(resp)
}
