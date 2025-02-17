package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"

	"github.com/LagrangeDev/LagrangeGo/client/entity"
	messagePkt "github.com/LagrangeDev/LagrangeGo/client/packets/message"
	oidb2 "github.com/LagrangeDev/LagrangeGo/client/packets/oidb"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/action"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/highway"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	message2 "github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

// SetOnlineStatus 设置在线状态
func (c *QQClient) SetOnlineStatus(status action.SetStatus) error {
	pkt, _ := proto.Marshal(&status)
	resp, err := c.sendUniPacketAndWait("trpc.qq_new_tech.status_svc.StatusService.SetStatus", pkt)
	if err != nil {
		return err
	}
	setstatusResp := action.SetStatusResponse{}
	err = proto.Unmarshal(resp, &setstatusResp)
	if err != nil {
		return err
	}
	if setstatusResp.Message != "set status success" {
		return fmt.Errorf("set status failed: %s", setstatusResp.Message)
	}
	return nil
}

// FetchFriends 获取好友列表信息，使用token可以获取下一页的群成员信息
func (c *QQClient) FetchFriends(token uint32) ([]*entity.User, uint32, error) {
	pkt, err := oidb2.BuildFetchFriendsReq(token)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, 0, err
	}
	friends, token, err := oidb2.ParseFetchFriendsResp(resp)
	if err != nil {
		return nil, 0, err
	}
	return friends, token, nil
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

// FetchGroupMember 获取对应群的群成员信息
func (c *QQClient) FetchGroupMember(groupUin, memberUin uint32) (*entity.GroupMember, error) {
	pkt, err := oidb2.BuildFetchMemberReq(groupUin, c.GetUID(memberUin, groupUin))
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	members, err := oidb2.ParseFetchMemberResp(resp)
	if err != nil {
		return nil, err
	}
	return members, nil
}

// FetchGroupMembers 获取对应群的所有群成员信息，使用token可以获取下一页的群成员信息
func (c *QQClient) FetchGroupMembers(groupUin uint32, token string) ([]*entity.GroupMember, string, error) {
	pkt, err := oidb2.BuildFetchMembersReq(groupUin, token)
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

// SetGroupRemark 设置群聊备注
func (c *QQClient) SetGroupRemark(groupUin uint32, remark string) error {
	pkt, err := oidb2.BuildSetGroupRemarkReq(groupUin, remark)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseSetGroupRemarkResp(resp)
}

// SetGroupName 设置群聊名称
func (c *QQClient) SetGroupName(groupUin uint32, name string) error {
	pkt, err := oidb2.BuildSetGroupNameReq(groupUin, name)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseSetGroupNameResp(resp)
}

// SetGroupGlobalMute 群全员禁言
func (c *QQClient) SetGroupGlobalMute(groupUin uint32, isMute bool) error {
	pkt, err := oidb2.BuildSetGroupGlobalMuteReq(groupUin, isMute)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseSetGroupGlobalMuteResp(resp)
}

// SetGroupMemberMute 禁言群成员
func (c *QQClient) SetGroupMemberMute(groupUin, uin, duration uint32) error {
	uid := c.GetUID(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildSetGroupMemberMuteReq(groupUin, duration, uid)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseSetGroupMemberMuteResp(resp)
}

// SetGroupLeave 退出群聊
func (c *QQClient) SetGroupLeave(groupUin uint32) error {
	pkt, err := oidb2.BuildSetGroupLeaveReq(groupUin)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseSetGroupLeaveResp(resp)
}

// SetGroupAdmin 设置群管理员
func (c *QQClient) SetGroupAdmin(groupUin, uin uint32, isAdmin bool) error {
	uid := c.GetUID(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildSetGroupAdminReq(groupUin, uid, isAdmin)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	err = oidb2.ParseSetGroupAdminResp(resp)
	if err != nil {
		return err
	}
	if m := c.GetCachedMemberInfo(uin, groupUin); m != nil {
		m.Permission = entity.Admin
		c.cache.RefreshGroupMember(groupUin, m)
	}
	return nil
}

// SetGroupMemberName 设置群成员昵称
func (c *QQClient) SetGroupMemberName(groupUin, uin uint32, name string) error {
	uid := c.GetUID(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildSetGroupMemberNameReq(groupUin, uid, name)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	err = oidb2.ParseSetGroupMemberNameResp(resp)
	if err != nil {
		return err
	}
	if m := c.GetCachedMemberInfo(uin, groupUin); m != nil {
		m.MemberCard = name
		c.cache.RefreshGroupMember(groupUin, m)
	}
	return nil
}

// KickGroupMember 踢出群成员，可选是否拒绝加群请求
func (c *QQClient) KickGroupMember(groupUin, uin uint32, rejectAddRequest bool) error {
	uid := c.GetUID(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildKickGroupMemberReq(groupUin, uid, rejectAddRequest)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseKickGroupMemberResp(resp)
}

// SetGroupMemberSpecialTitle 设置群成员专属头衔
func (c *QQClient) SetGroupMemberSpecialTitle(groupUin, uin uint32, title string) error {
	uid := c.GetUID(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildSetGroupMemberSpecialTitleReq(groupUin, uid, title)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseSetGroupMemberSpecialTitleResp(resp)
}

// SetGroupReaction 设置群消息表态
func (c *QQClient) SetGroupReaction(groupUin, sequence uint32, code string, isAdd bool) error {
	pkt, err := oidb2.BuildSetGroupReactionReq(groupUin, sequence, code, isAdd)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseSetGroupReactionResp(resp)
}

// GroupPoke 戳一戳群友
func (c *QQClient) GroupPoke(groupUin, uin uint32) error {
	pkt, err := oidb2.BuildGroupPokeReq(groupUin, uin)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParsePokeResp(resp)
}

// FriendPoke 戳一戳好友
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

// DeleteFriend 删除好友
func (c *QQClient) DeleteFriend(uin uint32, block bool) error {
	uid := c.GetUID(uin)
	pkt, err := oidb2.BuildDeleteFriendReq(uid, block)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseDeleteFriendResp(resp)
}

// RecallFriendMessage 撤回私聊消息
func (c *QQClient) RecallFriendMessage(uin, seq, random, clientSeq, timestamp uint32) error {
	packet := message.C2CRecallMsg{
		Type:      1,
		TargetUid: c.GetUID(uin),
		Info: &message.C2CRecallMsgInfo{
			ClientSequence:  clientSeq,
			Random:          random,
			MessageId:       0x10000000<<32 | uint64(random),
			Timestamp:       timestamp,
			Field5:          0,
			MessageSequence: seq,
		},
		Settings: &message.C2CRecallMsgSettings{
			Field1: false,
			Field2: false,
		},
		Field6: false,
	}
	pkt, err := proto.Marshal(&packet)
	if err != nil {
		return err
	}
	_, err = c.sendUniPacketAndWait("trpc.msg.msg_svc.MsgService.SsoC2CRecallMsg", pkt)
	if err != nil {
		return err
	}
	return nil // sbtx不报错
}

// RecallGroupMessage 撤回群聊消息
func (c *QQClient) RecallGroupMessage(groupUin, seq uint32) error {
	packet := message.GroupRecallMsg{
		Type:     1,
		GroupUin: groupUin,
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

// MarkPrivateMessageReaded 标记私聊消息已读
func (c *QQClient) MarkPrivateMessageReaded(uin, timestamp, startSeq uint32) error {
	uid := c.GetUID(uin)
	pakcet := message.SsoReadedReport{
		C2C: &message.SsoReadedReportC2C{
			TargetUid:     proto.Some(uid),
			Time:          timestamp,
			StartSequence: startSeq,
		},
	}
	pktData, err := proto.Marshal(&pakcet)
	if err != nil {
		return err
	}
	resp, err := c.sendUniPacketAndWait("trpc.msg.msg_svc.MsgService.SsoReadedReport", pktData)
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return errors.New("empty response data")
	}
	return nil
}

// MarkGroupMessageReaded 标记群消息已读
func (c *QQClient) MarkGroupMessageReaded(groupUin, startSeq uint32) error {
	pakcet := message.SsoReadedReport{
		Group: &message.SsoReadedReportGroup{
			GroupUin:      groupUin,
			StartSequence: startSeq,
		},
	}
	pktData, err := proto.Marshal(&pakcet)
	if err != nil {
		return err
	}
	resp, err := c.sendUniPacketAndWait("trpc.msg.msg_svc.MsgService.SsoReadedReport", pktData)
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return errors.New("empty response data")
	}
	return nil
}

// GetPrivateImageURL 获取私聊图片下载url
func (c *QQClient) GetPrivateImageURL(node *oidb.IndexNode) (string, error) {
	pkt, err := oidb2.BuildPrivateImageDownloadReq(c.GetUID(c.Uin), node)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParsePrivateImageDownloadResp(resp)
}

// GetGroupImageURL 获取群聊图片下载url
func (c *QQClient) GetGroupImageURL(groupUin uint32, node *oidb.IndexNode) (string, error) {
	pkt, err := oidb2.BuildGroupImageDownloadReq(groupUin, node)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParseGroupImageDownloadResp(resp)
}

// GetPrivateRecordURL 获取私聊语音下载url
func (c *QQClient) GetPrivateRecordURL(node *oidb.IndexNode) (string, error) {
	pkt, err := oidb2.BuildPrivateRecordDownloadReq(c.GetUID(c.Uin), node)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParsePrivateRecordDownloadResp(resp)
}

// GetGroupRecordURL 获取群聊语音下载url
func (c *QQClient) GetGroupRecordURL(groupUin uint32, node *oidb.IndexNode) (string, error) {
	pkt, err := oidb2.BuildGroupRecordDownloadReq(groupUin, node)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParseGroupRecordDownloadResp(resp)
}

func (c *QQClient) GenFileNode(name, md5, sha1, uuid string, size uint32, isnt bool) *oidb.IndexNode {
	return &oidb.IndexNode{
		Info: &oidb.FileInfo{
			FileName: name,
			FileSize: size,
			FileSha1: sha1,
			FileHash: md5,
		},
		FileUuid: uuid,
		StoreId:  utils.Ternary[uint32](isnt, 1, 0), // 0旧服务器 1为nt服务器
	}
}

// GetPrivateVideoURL 获取私聊视频下载链接
func (c *QQClient) GetPrivateVideoURL(node *oidb.IndexNode) (string, error) {
	pkt, err := oidb2.BuildPrivateVideoDownloadReq(c.Sig().UID, node)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParseVideoDownloadResp(resp)
}

// GetGroupVideoURL 获取群聊视频下载链接
func (c *QQClient) GetGroupVideoURL(groupUin uint32, node *oidb.IndexNode) (string, error) {
	pkt, err := oidb2.BuildGroupVideoDownloadReq(groupUin, node)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParseVideoDownloadResp(resp)
}

// GetGroupFileURL 获取群文件下载链接
func (c *QQClient) GetGroupFileURL(groupUin uint32, fileID string) (string, error) {
	pkt, err := oidb2.BuildGroupFSDownloadReq(groupUin, fileID)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParseGroupFSDownloadResp(resp)
}

// GetPrivateFileURL 获取私聊文件下载链接
func (c *QQClient) GetPrivateFileURL(fileUUID string, fileHash string) (string, error) {
	pkt, err := oidb2.BuildPrivateFileDownloadReq(c.GetUID(c.Uin), fileUUID, fileHash)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParsePrivateFileDownloadResp(resp)
}

// QueryGroupImage 获取群图片
func (c *QQClient) QueryGroupImage(md5 []byte, fileUUID string) (*message2.ImageElement, error) {
	switch {
	case fileUUID != "":
		rkeyInfo := c.GetCachedRkeyInfo(entity.GroupRKey)
		return &message2.ImageElement{
			URL: fmt.Sprintf("https://multimedia.nt.qq.com.cn/download?appid=1407&fileid=%s&rkey=%s", fileUUID, rkeyInfo.RKey),
		}, nil
	case len(md5) == 16:
		return &message2.ImageElement{
			URL: fmt.Sprintf("http://gchat.qpic.cn/gchatpic_new/0/0-0-%X/0", md5),
		}, nil
	default:
		return nil, errors.New("invalid parameters")
	}
}

// QueryFriendImage 获取私聊图片
func (c *QQClient) QueryFriendImage(md5 []byte, fileUUID string) (*message2.ImageElement, error) {
	switch {
	case fileUUID != "":
		rkeyInfo := c.GetCachedRkeyInfo(entity.FriendRKey)
		return &message2.ImageElement{
			URL: fmt.Sprintf("https://multimedia.nt.qq.com.cn/download?appid=1406&fileid=%s&rkey=%s", fileUUID, rkeyInfo.RKey),
		}, nil
	case len(md5) == 16:
		return &message2.ImageElement{
			URL: fmt.Sprintf("http://gchat.qpic.cn/gchatpic_new/0/0-0-%X/0", md5),
		}, nil
	default:
		return nil, errors.New("invalid parameters")
	}
}

// FetchUserInfo 获取用户信息
func (c *QQClient) FetchUserInfo(uid string) (*entity.User, error) {
	pkt, err := oidb2.BuildFetchUserInfoReq(uid)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	return oidb2.ParseFetchUserInfoResp(resp)
}

// FetchUserInfoUin 通过uin获取用户信息
func (c *QQClient) FetchUserInfoUin(uin uint32) (*entity.User, error) {
	pkt, err := oidb2.BuildFetchUserInfoReq(uin)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	return oidb2.ParseFetchUserInfoResp(resp)
}

// FetchGroupInfo 获取群信息 isStrange是否陌生群聊
func (c *QQClient) FetchGroupInfo(groupUin uint32, isStrange bool) (*entity.Group, error) {
	pkt, err := oidb2.BuildFetchGroupReq(groupUin, isStrange)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	groupResp, err := oidb2.ParseFetchGroupResp(resp)
	if err != nil {
		return nil, err
	}
	return &entity.Group{
		GroupUin:        groupResp.GroupUin,
		GroupName:       groupResp.GroupName,
		GroupOwner:      c.GetUin(groupResp.GroupOwner, groupUin),
		GroupCreateTime: groupResp.GroupCreateTime,
		GroupMemo:       groupResp.GroupMemo,
		GroupLevel:      groupResp.GroupLevel,
		MemberCount:     groupResp.GroupMemberNum,
		MaxMember:       groupResp.GroupMemberMaxNum,
		LastMsgSeq:      groupResp.GroupCurMsgSeq,
	}, nil
}

// GetGroupSystemMessages 获取加群请求信息
func (c *QQClient) GetGroupSystemMessages(isFiltered bool, count uint32, groupUin ...uint32) (*entity.GroupSystemMessages, error) {
	pkt, err := oidb2.BuildFetchGroupSystemMessagesReq(isFiltered, count)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	msgs, err := oidb2.ParseFetchGroupSystemMessagesReq(isFiltered, resp, groupUin...)
	if err != nil {
		return nil, err
	}
	for _, req := range msgs.InvitedRequests {
		if g, err := c.FetchGroupInfo(req.GroupUin, true); err == nil {
			req.GroupName = g.GroupName
		}
		if u, err := c.FetchUserInfoUin(req.InvitorUin); err == nil {
			req.InvitorNick = u.Nickname
		}
	}
	for _, req := range msgs.JoinRequests {
		if g, err := c.FetchGroupInfo(req.GroupUin, false); err == nil {
			req.GroupName = g.GroupName
		}
		if u, err := c.FetchUserInfoUin(req.TargetUin); err == nil {
			req.TargetNick = u.Nickname
		}
	}
	return msgs, nil
}

// SetGroupRequest 处理加群请求
func (c *QQClient) SetGroupRequest(isFiltered bool, accept bool, sequence uint64, typ uint32, groupUin uint32, message string) error {
	pkt, err := oidb2.BuildSetGroupRequestReq(isFiltered, accept, sequence, typ, groupUin, message)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseSetGroupRequestResp(resp)
}

// SetFriendRequest 处理好友请求
func (c *QQClient) SetFriendRequest(accept bool, targetUID string) error {
	pkt, err := oidb2.BuildSetFriendRequest(accept, targetUID)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseSetFriendRequestResp(resp)
}

// FetchRkey 获取Rkey
func (c *QQClient) FetchRkey() (entity.RKeyMap, error) {
	pkt, err := oidb2.BuildFetchRKeyReq()
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	return oidb2.ParseFetchRKeyResp(resp)
}

// FetchClientKey 获取ClientKey
func (c *QQClient) FetchClientKey() (string, error) {
	pkt, err := oidb2.BuildFetchClientKeyReq()
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParseFetchClientKeyResp(resp)
}

// FetchCookies 获取cookies
func (c *QQClient) FetchCookies(domains []string) ([]string, error) {
	pkt, err := oidb2.BuildFetchCookieReq(domains)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	return oidb2.ParseFetchCookieResp(resp)
}

// SendPrivateFile 发送私聊文件
func (c *QQClient) SendPrivateFile(targetUin uint32, localFilePath, filename string) error {
	fileElement, err := message2.NewLocalFile(localFilePath, filename)
	if err != nil {
		return err
	}
	uploadedFileElement, err := c.UploadPrivateFile(c.GetUID(targetUin), fileElement)
	if err != nil {
		return err
	}
	route := &message.RoutingHead{
		Trans0X211: &message.Trans0X211{
			CcCmd: proto.Uint32(4),
			Uid:   proto.String(c.GetUID(targetUin)),
		},
	}
	body := message2.PackElementsToBody([]message2.IMessageElement{uploadedFileElement})
	mr := crypto.RandU32()
	ret, _, err := c.SendRawMessage(route, body, mr)
	if err != nil || ret.PrivateSequence == 0 {
		return err
	}
	return nil
}

// SendGroupFile 发送群文件
func (c *QQClient) SendGroupFile(groupUin uint32, localFilePath, filename, targetDirectory string) error {
	fileElement, err := message2.NewLocalFile(localFilePath, filename)
	if err != nil {
		return err
	}
	if _, err = c.UploadGroupFile(groupUin, fileElement, targetDirectory); err != nil {
		return err
	}
	return nil
}

// GetGroupFileSystemInfo 获取群文件系统信息
func (c *QQClient) GetGroupFileSystemInfo(groupUin uint32) (*entity.GroupFileSystemInfo, error) {
	pkt, err := oidb2.BuildGroupFileCountReq(groupUin)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	fileCount, limitCount, err := oidb2.ParseGroupFileCountResp(resp)
	if err != nil {
		return nil, err
	}
	pkt, err = oidb2.BuildGroupFileSpaceReq(groupUin)
	if err != nil {
		return nil, err
	}
	resp, err = c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	totalSpace, usedSpace, err := oidb2.ParseGroupFileSpaceResp(resp)
	if err != nil {
		return nil, err
	}
	return &entity.GroupFileSystemInfo{
		GroupUin:   groupUin,
		FileCount:  fileCount,
		LimitCount: limitCount,
		TotalSpace: totalSpace,
		UsedSpace:  usedSpace,
	}, nil
}

// ListGroupFilesByFolder 获取群目录指定文件夹列表
func (c *QQClient) ListGroupFilesByFolder(groupUin uint32, targetDirectory string) ([]*entity.GroupFile, []*entity.GroupFolder, error) {
	var startIndex uint32
	var fileCount uint32 = 20
	var files []*entity.GroupFile
	var folders []*entity.GroupFolder
	for {
		pkt, err := oidb2.BuildGroupFileListReq(groupUin, targetDirectory, startIndex, fileCount)
		if err != nil {
			return files, folders, err
		}
		p, err := c.sendOidbPacketAndWait(pkt)
		if err != nil {
			return files, folders, err
		}
		res, err := oidb2.ParseGroupFileListResp(p)
		if err != nil {
			return files, folders, err
		}
		if res.List.IsEnd {
			break
		}
		for _, fe := range res.List.Items {
			if fe.FileInfo != nil {
				files = append(files, &entity.GroupFile{
					GroupUin:      groupUin,
					FileID:        fe.FileInfo.FileId,
					FileName:      fe.FileInfo.FileName,
					BusID:         fe.FileInfo.BusId,
					FileSize:      fe.FileInfo.FileSize,
					UploadTime:    fe.FileInfo.UploadedTime,
					DeadTime:      fe.FileInfo.ExpireTime,
					ModifyTime:    fe.FileInfo.ModifiedTime,
					DownloadTimes: fe.FileInfo.DownloadedTimes,
					Uploader:      fe.FileInfo.UploaderUin,
					UploaderName:  fe.FileInfo.UploaderName,
				})
			}
			if fe.FolderInfo != nil {
				folders = append(folders, &entity.GroupFolder{
					GroupUin:       groupUin,
					FolderID:       fe.FolderInfo.FolderId,
					FolderName:     fe.FolderInfo.FolderName,
					CreateTime:     fe.FolderInfo.CreateTime,
					Creator:        fe.FolderInfo.CreatorUin,
					CreatorName:    fe.FolderInfo.CreatorName,
					TotalFileCount: fe.FolderInfo.TotalFileCount,
				})
			}
		}
		startIndex += fileCount
	}
	return files, folders, nil
}

// ListGroupRootFiles 获取群根目录文件列表
func (c *QQClient) ListGroupRootFiles(groupUin uint32) ([]*entity.GroupFile, []*entity.GroupFolder, error) {
	return c.ListGroupFilesByFolder(groupUin, "/")
}

// RenameGroupFile 重命名群文件
func (c *QQClient) RenameGroupFile(groupUin uint32, fileID string, parentFolder string, newFileName string) error {
	pkt, err := oidb2.BuildGroupFileRenameReq(groupUin, fileID, parentFolder, newFileName)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupFileRenameResp(resp)
}

// MoveGroupFile 移动群文件
func (c *QQClient) MoveGroupFile(groupUin uint32, fileID string, parentFolder string, targetFolderID string) error {
	pkt, err := oidb2.BuildGroupFileMoveReq(groupUin, fileID, parentFolder, targetFolderID)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupFileMoveResp(resp)
}

// DeleteGroupFile 删除群文件
func (c *QQClient) DeleteGroupFile(groupUin uint32, fileID string) error {
	pkt, err := oidb2.BuildGroupFileDeleteReq(groupUin, fileID)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupFileDeleteResp(resp)
}

// CreateGroupFolder 创建群文件夹
func (c *QQClient) CreateGroupFolder(groupUin uint32, targetDirectory string, folderName string) error {
	pkt, err := oidb2.BuildGroupFolderCreateReq(groupUin, targetDirectory, folderName)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupFolderCreateResp(resp)
}

// RenameGroupFolder 重命名群文件夹
func (c *QQClient) RenameGroupFolder(groupUin uint32, folderID string, newFolderName string) error {
	pkt, err := oidb2.BuildGroupFolderRenameReq(groupUin, folderID, newFolderName)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupFolderRenameResp(resp)
}

// DeleteGroupFolder 删除群文件夹
func (c *QQClient) DeleteGroupFolder(groupUin uint32, folderID string) error {
	pkt, err := oidb2.BuildGroupFolderDeleteReq(groupUin, folderID)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupFolderDeleteResp(resp)
}

// FetchForwardMsg 获取合并转发消息
func (c *QQClient) FetchForwardMsg(resID string) (msg *message2.ForwardMessage, err error) {
	if resID == "" {
		return msg, errors.New("empty resID")
	}
	pkt, err := messagePkt.BuildMultiMsgDownloadReq(c.GetUID(c.Uin), resID)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendUniPacketAndWait("trpc.group.long_msg_interface.MsgService.SsoRecvLongMsg", pkt)
	if err != nil {
		return nil, err
	}
	pasted, err := messagePkt.ParseMultiMsgDownloadResp(resp)
	if err != nil {
		return nil, err
	}
	if pasted.Result == nil || pasted.Result.Payload == nil {
		return nil, errors.New("empty response data")
	}
	data := binary.GZipUncompress(pasted.Result.Payload)
	result := &message.LongMsgResult{}
	if err = proto.Unmarshal(data, result); err != nil {
		return nil, err
	}

	forwardMsg := &message2.ForwardMessage{ResID: resID}
	for _, action := range result.Action {
		if action.ActionCommand == "MultiMsg" {
			forwardMsg.Nodes = make([]*message2.ForwardNode, len(action.ActionData.MsgBody))
			for idx, b := range action.ActionData.MsgBody {
				forwardMsg.Nodes[idx] = &message2.ForwardNode{
					SenderID: b.ResponseHead.FromUin,
					Time:     b.ContentHead.TimeStamp.Unwrap(),
				}
				if forwardMsg.IsGroup = b.ResponseHead.Grp != nil; forwardMsg.IsGroup {
					forwardMsg.Nodes[idx].GroupID = b.ResponseHead.Grp.GroupUin
					forwardMsg.Nodes[idx].SenderName = b.ResponseHead.Grp.MemberName
					grpMsg := message2.ParseGroupMessage(b)
					c.PreprocessGroupMessageEvent(grpMsg)
					forwardMsg.Nodes[idx].Message = grpMsg.Elements
				} else {
					forwardMsg.Nodes[idx].SenderName = b.ResponseHead.Forward.FriendName.Unwrap()
					prvMsg := message2.ParsePrivateMessage(b)
					c.PreprocessPrivateMessageEvent(prvMsg)
					forwardMsg.Nodes[idx].Message = prvMsg.Elements
				}
			}
		}
	}
	return forwardMsg, nil
}

// UploadForwardMsg 上传合并转发消息
// groupUin should be the group number where the uploader is located or 0 (c2c)
func (c *QQClient) UploadForwardMsg(forward *message2.ForwardMessage, groupUin uint32) (*message2.ForwardMessage, error) {
	msgBody := c.BuildFakeMessage(forward.Nodes)
	pkt, err := messagePkt.BuildMultiMsgUploadReq(c.GetUID(c.Uin), groupUin, msgBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendUniPacketAndWait("trpc.group.long_msg_interface.MsgService.SsoSendLongMsg", pkt)
	if err != nil {
		return nil, err
	}
	pasted, err := messagePkt.ParseMultiMsgUploadResp(resp)
	if err != nil {
		return nil, err
	}
	if pasted.Result == nil {
		return nil, errors.New("empty response data")
	}
	forward.ResID = pasted.Result.ResId
	return forward, nil
}

// FetchEssenceMessage 获取精华消息
func (c *QQClient) FetchEssenceMessage(groupUin uint32) ([]*message2.GroupEssenceMessage, error) {
	var essenceMsg []*message2.GroupEssenceMessage
	page := 0
	bkn, err := c.GetCsrfToken()
	if err != nil {
		return essenceMsg, err
	}
	grpInfo := c.GetCachedGroupInfo(groupUin)
	for {
		reqURL := fmt.Sprintf("https://qun.qq.com/cgi-bin/group_digest/digest_list?random=7800&X-CROSS-ORIGIN=fetch&group_code=%d&page_start=%d&page_limit=20&bkn=%d", groupUin, page, bkn)
		req, err := http.NewRequest(http.MethodGet, reqURL, nil)
		if err != nil {
			return essenceMsg, err
		}
		resp, err := c.SendRequestWithCookie(req)
		if err != nil {
			return essenceMsg, err
		}
		respData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return essenceMsg, fmt.Errorf("error resp code %d", resp.StatusCode)
		}
		respJSON := gjson.ParseBytes(respData)
		if respJSON.Get("retcode").Int() != 0 {
			return essenceMsg, fmt.Errorf("error code %d, %s", respJSON.Get("retcode").Int(), respJSON.Get("retmsg").String())
		}
		for _, v := range respJSON.Get("data").Get("msg_list").Array() {
			var elements []message2.IMessageElement
			for _, e := range v.Get("msg_content").Array() {
				switch e.Get("msg_type").Int() {
				case 1:
					elements = append(elements, &message2.TextElement{Content: e.Get("text").String()})
				case 2:
					elements = append(elements, &message2.FaceElement{FaceID: uint32(e.Get("face_index").Int())})
				case 3:
					elements = append(elements, &message2.ImageElement{URL: e.Get("image_url").String()})
				case 4:
					elements = append(elements, &message2.FileElement{
						FileID:  e.Get("file_id").String(),
						FileURL: e.Get("file_thumbnail_url").String(),
					})
				}
			}
			senderUin := uint32(v.Get("sender_uin").Int())
			senderInfo := c.GetCachedMemberInfo(senderUin, groupUin)
			essenceMsg = append(essenceMsg, &message2.GroupEssenceMessage{
				OperatorUin:  uint32(v.Get("add_digest_uin").Int()),
				OperatorUID:  c.GetUID(uint32(v.Get("add_digest_uin").Int())),
				OperatorTime: uint64(v.Get("add_digest_time").Int()),
				CanRemove:    v.Get("can_be_removed").Bool(),
				Message: &message2.GroupMessage{
					ID:         uint32(v.Get("msg_seq").Int()),
					InternalID: uint32(v.Get("msg_random").Int()),
					GroupUin:   grpInfo.GroupUin,
					GroupName:  grpInfo.GroupName,
					Sender: &message2.Sender{
						Uin:      senderUin,
						UID:      c.GetUID(senderUin, groupUin),
						Nickname: senderInfo.Nickname,
						CardName: senderInfo.MemberCard,
					},
					Time:     uint32(v.Get("sender_time").Int()),
					Elements: elements,
				},
			})
		}
		if respJSON.Get("data").Get("is_end").Bool() {
			break
		}
	}
	return essenceMsg, nil
}

// GetGroupHonorInfo 获取群荣誉信息
// reference https://github.com/Mrs4s/MiraiGo/blob/master/client/http_api.go
func (c *QQClient) GetGroupHonorInfo(groupUin uint32, honorType entity.HonorType) (*entity.GroupHonorInfo, error) {
	ret := &entity.GroupHonorInfo{}
	honorRe := regexp.MustCompile(`window\.__INITIAL_STATE__\s*?=\s*?(\{.*})`)
	reqURL := fmt.Sprintf("https://qun.qq.com/interactive/honorlist?gc=%d&type=%d", groupUin, honorType)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return ret, err
	}
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return ret, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ret, fmt.Errorf("error resp code %d", resp.StatusCode)
	}
	matched := honorRe.FindSubmatch(respData)
	if len(matched) == 0 {
		return nil, errors.New("no matched data")
	}
	err = json.NewDecoder(bytes.NewReader(matched[1])).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// GetGroupNotice 获取群公告
func (c *QQClient) GetGroupNotice(groupUin uint32) (l []*entity.GroupNoticeFeed, err error) {
	bkn, err := c.GetCsrfToken()
	if err != nil {
		return nil, err
	}
	v := url.Values{}
	v.Set("bkn", strconv.Itoa(bkn))
	v.Set("qid", strconv.FormatInt(int64(groupUin), 10))
	v.Set("ft", "23")
	v.Set("ni", "1")
	v.Set("n", "1")
	v.Set("i", "1")
	v.Set("log_read", "1")
	v.Set("platform", "1")
	v.Set("s", "-1")
	v.Set("n", "20")
	req, _ := http.NewRequest(http.MethodGet, "https://web.qun.qq.com/cgi-bin/announce/get_t_list?"+v.Encode(), nil)
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error resp code %d", resp.StatusCode)
	}
	r := entity.GroupNoticeRsp{}
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return
	}
	_ = resp.Body.Close()
	o := make([]*entity.GroupNoticeFeed, 0, len(r.Feeds)+len(r.Inst))
	o = append(o, r.Feeds...)
	o = append(o, r.Inst...)
	return o, nil
}

func (c *QQClient) uploadGroupNoticePic(bkn int, img []byte) (*entity.NoticeImage, error) {
	ret := &entity.NoticeImage{}
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	_ = w.WriteField("bkn", strconv.Itoa(bkn))
	_ = w.WriteField("source", "troopNotice")
	_ = w.WriteField("m", "0")
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="pic_up"; filename="temp_uploadFile.png"`)
	h.Set("Content-Type", "image/png")
	fw, _ := w.CreatePart(h)
	_, _ = fw.Write(img)
	_ = w.Close()
	req, err := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/upload_img", buf)
	if err != nil {
		return ret, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return ret, err
	}
	var res entity.NoticePicUpResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return ret, err
	}
	_ = resp.Body.Close()
	if res.ErrorCode != 0 {
		return ret, errors.New(res.ErrorMessage)
	}
	err = json.Unmarshal([]byte(html.UnescapeString(res.ID)), &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

// AddGroupNoticeSimple 发群公告
func (c *QQClient) AddGroupNoticeSimple(groupUin uint32, text string) (noticeID string, err error) {
	bkn, err := c.GetCsrfToken()
	if err != nil {
		return "", err
	}
	body := fmt.Sprintf(`qid=%v&bkn=%v&text=%v&pinned=0&type=1&settings={"is_show_edit_card":0,"tip_window_type":1,"confirm_required":1}`, groupUin, bkn, url.QueryEscape(text))
	req, err := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/add_qun_notice?bkn="+strconv.Itoa(bkn), strings.NewReader(body))
	if err != nil {
		return "", err
	}
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return "", err
	}
	var res entity.NoticeSendResp
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}
	_ = resp.Body.Close()
	return res.NoticeID, nil
}

// AddGroupNoticeWithPic 发群公告带图片
func (c *QQClient) AddGroupNoticeWithPic(groupUin uint32, text string, pic []byte) (noticeID string, err error) {
	bkn, err := c.GetCsrfToken()
	if err != nil {
		return "", err
	}
	img, err := c.uploadGroupNoticePic(bkn, pic)
	if err != nil {
		return "", err
	}
	body := fmt.Sprintf(`qid=%v&bkn=%v&text=%v&pinned=0&type=1&settings={"is_show_edit_card":0,"tip_window_type":1,"confirm_required":1}&pic=%v&imgWidth=%v&imgHeight=%v`, groupUin, bkn, url.QueryEscape(text), img.ID, img.Width, img.Height)
	req, err := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/add_qun_notice?bkn="+strconv.Itoa(bkn), strings.NewReader(body))
	if err != nil {
		return "", err
	}
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return "", err
	}
	var res entity.NoticeSendResp
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}
	_ = resp.Body.Close()
	return res.NoticeID, nil
}

// DelGroupNotice 删除群公告
func (c *QQClient) DelGroupNotice(groupUin uint32, fid string) error {
	bkn, err := c.GetCsrfToken()
	if err != nil {
		return err
	}
	body := fmt.Sprintf(`fid=%s&qid=%v&bkn=%v&ft=23&op=1`, fid, groupUin, bkn)
	req, err := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/del_feed", strings.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// SetAvatar 设置头像
func (c *QQClient) SetAvatar(avatar io.ReadSeeker) error {
	if avatar == nil {
		return errors.New("avatar is nil")
	}
	md5, size := crypto.ComputeMd5AndLength(avatar)
	return c.highwayUpload(90, avatar, uint64(size), md5, nil)
}

// SetGroupAvatar 设置群头像
func (c *QQClient) SetGroupAvatar(groupUin uint32, avatar io.ReadSeeker) error {
	if avatar == nil {
		return errors.New("avatar is nil")
	}
	extra := highway.GroupAvatarExtra{
		Type:     101,
		GroupUin: groupUin,
		Field3:   &highway.GroupAvatarExtraField3{Field1: 1},
		Field5:   3,
		Field6:   1,
	}
	extStream, err := proto.Marshal(&extra)
	if err != nil {
		return err
	}
	md5, size := crypto.ComputeMd5AndLength(avatar)
	return c.highwayUpload(3000, avatar, uint64(size), md5, extStream)
}

// SetEssenceMessage 设置群聊精华消息
func (c *QQClient) SetEssenceMessage(groupUin, seq, random uint32, isSet bool) error {
	pkt, err := oidb2.BuildSetEssenceMessageReq(groupUin, seq, random, isSet)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseSetEssenceMessageResp(resp)
}

// SendFriendLike 给好友点赞
func (c *QQClient) SendFriendLike(uin uint32, count uint32) error {
	if count > 20 {
		count = 20
	} else if count < 1 {
		count = 1
	}
	pkt, err := oidb2.BuildFriendLikeReq(c.GetUID(uin), count)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseFriendLikeResp(resp)
}

func (c *QQClient) GetPrivateMessages(uin, timestamp, count uint32) ([]*message2.PrivateMessage, error) {
	uid := c.GetUID(uin)
	pkt, err := proto.Marshal(&message.SsoGetRoamMsg{
		FriendUid: proto.Some(uid),
		Time:      timestamp,
		Random:    0,
		Count:     count,
		Direction: true,
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.sendUniPacketAndWait("trpc.msg.register_proxy.RegisterProxy.SsoGetRoamMsg", pkt)
	if err != nil {
		return nil, err
	}
	roamMsg := message.SsoGetRoamMsgResponse{}
	err = proto.Unmarshal(resp, &roamMsg)
	if err != nil {
		return nil, err
	}

	ret := make([]*message2.PrivateMessage, 0, len(roamMsg.Messages))
	for _, msg := range roamMsg.Messages {
		m := message2.ParsePrivateMessage(msg)
		c.PreprocessPrivateMessageEvent(m)
		ret = append(ret, m)
	}
	return ret, nil
}

// GetGroupMessages 获取群聊历史消息
func (c *QQClient) GetGroupMessages(groupUin, startSeq, endSeq uint32) ([]*message2.GroupMessage, error) {
	pkt, err := proto.Marshal(&message.SsoGetGroupMsg{
		Info: &message.SsoGetGroupMsgInfo{
			GroupUin:      groupUin,
			StartSequence: startSeq,
			EndSequence:   endSeq,
		},
		Direction: true,
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.sendUniPacketAndWait("trpc.msg.register_proxy.RegisterProxy.SsoGetGroupMsg", pkt)
	if err != nil {
		return nil, err
	}

	var groupMsg message.SsoGetGroupMsgResponse
	err = proto.Unmarshal(resp, &groupMsg)
	if err != nil {
		return nil, err
	}

	ret := make([]*message2.GroupMessage, 0, len(groupMsg.Body.Messages))
	for _, msg := range groupMsg.Body.Messages {
		m := message2.ParseGroupMessage(msg)
		c.PreprocessGroupMessageEvent(m)
		ret = append(ret, m)
	}

	return ret, nil
}

// ImageOcr 图片识别 有些域名的图可能无法识别，需要重新上传到tx服务器并获取图片下载链接
func (c *QQClient) ImageOcr(url string) (*oidb2.OcrResponse, error) {
	if url != "" {
		pkt, err := oidb2.BuildImageOcrRequestPacket(url)
		if err != nil {
			return nil, err
		}
		resp, err := c.sendOidbPacketAndWait(pkt)
		if err != nil {
			return nil, err
		}
		return oidb2.ParseImageOcrResp(resp)
	}
	return nil, errors.New("image error")
}

// SendGroupSign 发送群聊打卡消息
func (c *QQClient) SendGroupSign(groupUin uint32) (*oidb2.BotGroupClockInResult, error) {
	pkt, err := oidb2.BuildGroupSignPacket(c.Uin, groupUin, c.Version().CurrentVersion)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	return oidb2.ParseGroupSignResp(resp)
}

// GetUnidirectionalFriendList 获取单向好友列表
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/web.go#L23
func (c *QQClient) GetUnidirectionalFriendList() ([]*entity.User, error) {
	webRsp := &struct {
		BlockList []struct {
			Uin         uint32 `json:"uint64_uin"`
			NickBytes   string `json:"bytes_nick"`
			Age         uint32 `json:"uint32_age"`
			Sex         uint32 `json:"uint32_sex"`
			SourceBytes string `json:"bytes_source"`
			UID         string `json:"str_uid"`
		} `json:"rpt_block_list"`
		ErrorCode int32 `json:"ErrorCode"`
	}{}
	rsp, err := c.webSsoRequest("ti.qq.com", "OidbSvc.0xe17_0", fmt.Sprintf(`{"uint64_uin":%v,"uint64_top":0,"uint32_req_num":99,"bytes_cookies":""}`, c.Uin))
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(utils.S2B(rsp), webRsp); err != nil {
		return nil, errors.Wrap(err, "unmarshal json error")
	}
	if webRsp.ErrorCode != 0 {
		return nil, fmt.Errorf("web sso request error: %v", webRsp.ErrorCode)
	}
	ret := make([]*entity.User, 0, len(webRsp.BlockList))
	for _, block := range webRsp.BlockList {
		decodeBase64String := func(str string) string {
			b, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				return ""
			}
			return utils.B2S(b)
		}
		ret = append(ret, &entity.User{
			Uin:      block.Uin,
			UID:      block.UID,
			Nickname: decodeBase64String(block.NickBytes),
			Age:      block.Age,
			Source:   decodeBase64String(block.SourceBytes),
		})
	}
	return ret, err
}

// DeleteUnidirectionalFriend 删除单向好友
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/web.go#L62
func (c *QQClient) DeleteUnidirectionalFriend(uin uint32) error {
	webRsp := &struct {
		ErrorCode int32 `json:"ErrorCode"`
	}{}
	rsp, err := c.webSsoRequest("ti.qq.com", "OidbSvc.0x5d4_0", fmt.Sprintf(`{"uin_list":[%v]}`, uin))
	if err != nil {
		return err
	}
	if err = json.Unmarshal(utils.S2B(rsp), webRsp); err != nil {
		return errors.Wrap(err, "unmarshal json error")
	}
	if webRsp.ErrorCode != 0 {
		return fmt.Errorf("web sso request error: %v", webRsp.ErrorCode)
	}
	return nil
}

// CheckURLSafely 通过TX服务器检查URL安全性
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/security.go#L24
func (c *QQClient) CheckURLSafely(url string) (oidb2.URLSecurityLevel, error) {
	pkt, err := oidb2.BuildURLCheckRequest(c.Uin, url)
	if err != nil {
		return oidb2.URLSecurityLevelUnknown, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return oidb2.URLSecurityLevelUnknown, err
	}
	return oidb2.ParseURLCheckResponse(resp)
}

// GetAtAllRemain 获取剩余@全员次数
// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/group_msg.go#L68
func (c *QQClient) GetAtAllRemain(uin, groupUin uint32) (*oidb2.AtAllRemainInfo, error) {
	pkt, err := oidb2.BuildGetAtAllRemainRequest(uin, groupUin)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	return oidb2.ParseGetAtAllRemainResponse(resp)
}

// GetAiCharacters 获取AI语音角色列表
func (c *QQClient) GetAiCharacters(groupUin uint32, chatType entity.ChatType) (*entity.AiCharacterList, error) {
	if groupUin == 0 {
		groupUin = 42
	}
	pkt, err := oidb2.BuildAiCharacterListService(groupUin, chatType)
	if err != nil {
		return nil, err
	}
	rsp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	result, err := oidb2.ParseAiCharacterListService(rsp)
	if err != nil {
		return nil, err
	}
	result.Type = chatType
	return result, nil
}

// SendGroupAiRecord 发送群AI语音
func (c *QQClient) SendGroupAiRecord(groupUin uint32, chatType entity.ChatType, voiceID, text string) (*message2.VoiceElement, error) {
	pkt, err := oidb2.BuildGroupAiRecordService(groupUin, voiceID, text, chatType, crypto.RandU32())
	if err != nil {
		return nil, err
	}
	rsp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	return oidb2.ParseGroupAiRecordService(rsp)
}

// FetchMarketFaceKey 获取魔法表情key
func (c *QQClient) FetchMarketFaceKey(faceIDs ...string) ([]string, error) {
	for i, v := range faceIDs {
		faceIDs[i] = strings.ToLower(v)
	}
	pkt, err := proto.Marshal(&message.MarketFaceKeyReq{
		Field1: 3,
		Info:   &message.MarketFaceKeyReqInfo{FaceIds: faceIDs},
	})
	if err != nil {
		return nil, err
	}
	rsp, err := c.sendUniPacketAndWait("BQMallSvc.TabOpReq", pkt)
	if err != nil {
		return nil, err
	}
	var info message.MarketFaceKeyRsp
	if err = proto.Unmarshal(rsp, &info); err != nil {
		return nil, err
	}
	if info.Info == nil {
		return nil, errors.New("valid ids")
	}
	return info.Info.Keys, nil
}
