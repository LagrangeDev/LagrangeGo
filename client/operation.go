package client

import (
	"errors"

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

func (c *QQClient) GroupRemark(groupID uint32, remark string) error {
	pkt, err := oidb.BuildGroupRemarkReq(groupID, remark)
	if err != nil {
		return err
	}
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParseGroupRemarkResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("remark failed")
	}

	return nil
}

func (c *QQClient) GroupRename(groupID uint32, name string) error {
	pkt, err := oidb.BuildGroupRenameReq(groupID, name)
	if err != nil {
		return err
	}
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParseGroupRenameResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("rename failed")
	}

	return nil
}

func (c *QQClient) GroupMuteGlobal(groupID uint32, isMute bool) error {
	pkt, err := oidb.BuildGroupMuteGlobalReq(groupID, isMute)
	if err != nil {
		return err
	}
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParseGroupMuteGlobalResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("mute failed")
	}

	return nil
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
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParseGroupMuteMemberResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("mute failed")
	}

	return nil
}

func (c *QQClient) GroupLeave(groupID uint32) error {
	pkt, err := oidb.BuildGroupLeaveReq(groupID)
	if err != nil {
		return err
	}
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParseGroupLeaveResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("leave failed")
	}

	return nil
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
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParseGroupSetAdminResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("set admin failed")
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
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParseGroupRenameMemberResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("rename member failed")
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
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParseGroupKickMemberResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("kick member failed")
	}

	return nil
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
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParseGroupSetSpecialTitleResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("set special title failed")
	}

	return nil
}

func (c *QQClient) GroupPoke(groupID, uin uint32) error {
	pkt, err := oidb.BuildGroupPokeReq(groupID, uin)
	if err != nil {
		return err
	}
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParsePokeResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("poke failed")
	}

	return nil
}

func (c *QQClient) FriendPoke(uin uint32) error {
	pkt, err := oidb.BuildFriendPokeReq(uin)
	if err != nil {
		return err
	}
	resp, err := c.SendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	ok, err := oidb.ParsePokeResp(resp.Data)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("poke failed")
	}

	return nil
}
