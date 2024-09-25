package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/entity"
	messagePkt "github.com/LagrangeDev/LagrangeGo/client/packets/message"
	oidb2 "github.com/LagrangeDev/LagrangeGo/client/packets/oidb"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	message2 "github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/tidwall/gjson"
)

// FetchFriends 获取好友列表信息，使用token可以获取下一页的群成员信息
func (c *QQClient) FetchFriends(token uint32) ([]*entity.Friend, uint32, error) {
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
	pkt, err := oidb2.BuildFetchMemberReq(groupUin, c.GetUid(memberUin, groupUin))
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

// GroupRemark 设置群聊备注
func (c *QQClient) GroupRemark(groupUin uint32, remark string) error {
	pkt, err := oidb2.BuildGroupRemarkReq(groupUin, remark)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupRemarkResp(resp)
}

// GroupRename 设置群聊名称
func (c *QQClient) GroupRename(groupUin uint32, name string) error {
	pkt, err := oidb2.BuildGroupRenameReq(groupUin, name)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupRenameResp(resp)
}

// GroupMuteGlobal 群全员禁言
func (c *QQClient) GroupMuteGlobal(groupUin uint32, isMute bool) error {
	pkt, err := oidb2.BuildGroupMuteGlobalReq(groupUin, isMute)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupMuteGlobalResp(resp)
}

// GroupMuteMember 禁言群成员
func (c *QQClient) GroupMuteMember(groupUin, uin, duration uint32) error {
	uid := c.GetUid(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildGroupMuteMemberReq(groupUin, duration, uid)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupMuteMemberResp(resp)
}

// GroupLeave 退出群聊
func (c *QQClient) GroupLeave(groupUin uint32) error {
	pkt, err := oidb2.BuildGroupLeaveReq(groupUin)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupLeaveResp(resp)
}

// GroupSetAdmin 设置群管理员
func (c *QQClient) GroupSetAdmin(groupUin, uin uint32, isAdmin bool) error {
	uid := c.GetUid(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildGroupSetAdminReq(groupUin, uid, isAdmin)
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
	if m := c.GetCachedMemberInfo(uin, groupUin); m != nil {
		m.Permission = entity.Admin
		c.cache.RefreshGroupMember(groupUin, m)
	}

	return nil
}

// GroupRenameMember 设置群成员昵称
func (c *QQClient) GroupRenameMember(groupUin, uin uint32, name string) error {
	uid := c.GetUid(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildGroupRenameMemberReq(groupUin, uid, name)
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
	if m := c.GetCachedMemberInfo(uin, groupUin); m != nil {
		m.MemberCard = name
		c.cache.RefreshGroupMember(groupUin, m)
	}

	return nil
}

// GroupKickMember 踢出群成员，可选是否拒绝加群请求
func (c *QQClient) GroupKickMember(groupUin, uin uint32, rejectAddRequest bool) error {
	uid := c.GetUid(uin, groupUin)
	if uid == "" {
		return errors.New("uid not found")
	}
	pkt, err := oidb2.BuildGroupKickMemberReq(groupUin, uid, rejectAddRequest)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupKickMemberResp(resp)
}

// GroupSetSpecialTitle 设置群成员专属头衔
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

func (c *QQClient) GroupSetReaction(groupUin, sequence uint32, code string, isAdd bool) error {
	pkt, err := oidb2.BuildGroupSetReactionReq(groupUin, sequence, code, isAdd)
	if err != nil {
		return err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return err
	}
	return oidb2.ParseGroupSetReactionResp(resp)
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

func (c *QQClient) RecallFriendMessage(uin, seq, random, clientSeq, timestamp uint32) error {
	packet := message.C2CRecallMsg{
		Type:      1,
		TargetUid: c.GetUid(uin),
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

// GetPrivateImageUrl 获取私聊图片下载url
func (c *QQClient) GetPrivateImageUrl(node *oidb.IndexNode) (string, error) {
	pkt, err := oidb2.BuildPrivateImageDownloadReq(c.GetUid(c.Uin), node)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParsePrivateImageDownloadResp(resp)
}

// GetGroupImageUrl 获取群聊图片下载url
func (c *QQClient) GetGroupImageUrl(groupUin uint32, node *oidb.IndexNode) (string, error) {
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

// GetPrivateRecordUrl 获取私聊语音下载url
func (c *QQClient) GetPrivateRecordUrl(node *oidb.IndexNode) (string, error) {
	pkt, err := oidb2.BuildPrivateRecordDownloadReq(c.GetUid(c.Uin), node)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParsePrivateRecordDownloadResp(resp)
}

// GetGroupRecordUrl 获取群聊语音下载url
func (c *QQClient) GetGroupRecordUrl(groupUin uint32, node *oidb.IndexNode) (string, error) {
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

func (c *QQClient) GetVideoUrl(isGroup bool, video *message2.ShortVideoElement) (string, error) {
	pkt, err := oidb2.BuildVideoDownloadReq(c.Sig().Uid, string(video.Uuid), video.Name, isGroup, video.Md5, video.Sha1)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParseVideoDownloadResp(resp)
}

func (c *QQClient) GetGroupFileUrl(groupUin uint32, fileID string) (string, error) {
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

func (c *QQClient) GetPrivateFileUrl(fileUUID string, fileHash string) (string, error) {
	pkt, err := oidb2.BuildPrivateFileDownloadReq(c.GetUid(c.Uin), fileUUID, fileHash)
	if err != nil {
		return "", err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return "", err
	}
	return oidb2.ParsePrivateFileDownloadResp(resp)
}

func (c *QQClient) queryImage(url string, method string) (*message2.ImageElement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			return
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("file not found")
	}
	return &message2.ImageElement{
		Url:  url,
		Size: uint32(resp.ContentLength),
	}, nil
}

func (c *QQClient) QueryGroupImage(md5 []byte, fileUUID string) (*message2.ImageElement, error) {
	var url string
	if fileUUID != "" {
		rkeyInfo := c.GetCachedRkeyInfo(entity.GroupRKey)
		url = fmt.Sprintf("https://multimedia.nt.qq.com.cn/download?appid=1407&fileid=%s&rkey=%s", fileUUID, rkeyInfo.RKey)
		return c.queryImage(url, http.MethodGet)
	} else if len(md5) == 16 {
		url = fmt.Sprintf("http://gchat.qpic.cn/gchatpic_new/0/0-0-%X/0", md5)
		return c.queryImage(url, http.MethodHead)
	} else {
		return nil, errors.New("invalid parameters")
	}
}

func (c *QQClient) QueryFriendImage(md5 []byte, fileUUID string) (*message2.ImageElement, error) {
	var url string
	if fileUUID != "" {
		rkeyInfo := c.GetCachedRkeyInfo(entity.FriendRKey)
		url = fmt.Sprintf("https://multimedia.nt.qq.com.cn/download?appid=1406&fileid=%s&rkey=%s", fileUUID, rkeyInfo.RKey)
		return c.queryImage(url, http.MethodGet)
	} else if len(md5) == 16 {
		url = fmt.Sprintf("http://gchat.qpic.cn/gchatpic_new/0/0-0-%X/0", md5)
		return c.queryImage(url, http.MethodHead)
	} else {
		return nil, errors.New("invalid parameters")
	}
}

// FetchUserInfo 获取用户信息
func (c *QQClient) FetchUserInfo(uid string) (*entity.Friend, error) {
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
func (c *QQClient) FetchUserInfoUin(uin uint32) (*entity.Friend, error) {
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

// GetGroupSystemMessages 获取加群请求信息
func (c *QQClient) GetGroupSystemMessages(isFiltered bool, count uint32, groupUin ...uint32) ([]*entity.GroupJoinRequest, error) {
	pkt, err := oidb2.BuildFetchGroupSystemMessagesReq(isFiltered, count)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendOidbPacketAndWait(pkt)
	if err != nil {
		return nil, err
	}
	return oidb2.ParseFetchGroupSystemMessagesReq(isFiltered, resp, groupUin...)
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
func (c *QQClient) SetFriendRequest(accept bool, targetUid string) error {
	pkt, err := oidb2.BuildSetFriendRequest(accept, targetUid)
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

// UploadPrivateFile 上传私聊文件
func (c *QQClient) UploadPrivateFile(targetUin uint32, localFilePath string) error {
	fileElement, err := message2.NewLocalFile(localFilePath)
	if err != nil {
		return err
	}
	uploadedFileElement, err := c.FileUploadPrivate(c.GetUid(targetUin), fileElement)
	if err != nil {
		return err
	}
	route := &message.RoutingHead{
		Trans0X211: &message.Trans0X211{
			CcCmd: proto.Uint32(4),
			Uid:   proto.String(c.GetUid(targetUin)),
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

// UploadGroupFile 上传群文件
func (c *QQClient) UploadGroupFile(groupUin uint32, localFilePath string, targetDirectory string) error {
	fileElement, err := message2.NewLocalFile(localFilePath)
	if err != nil {
		return err
	}
	if _, err = c.FileUploadGroup(groupUin, fileElement, targetDirectory); err != nil {
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
	var startIndex uint32 = 0
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
					FileId:        fe.FileInfo.FileId,
					FileName:      fe.FileInfo.FileName,
					BusId:         fe.FileInfo.BusId,
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
					FolderId:       fe.FolderInfo.FolderId,
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
func (c *QQClient) FetchForwardMsg(resId string) (msg *message2.ForwardMessage, err error) {
	if resId == "" {
		return msg, errors.New("empty resId")
	}
	forwardMsg := &message2.ForwardMessage{ResID: resId}
	pkt, err := messagePkt.BuildMultiMsgDownloadReq(c.GetUid(c.Uin), resId)
	if err != nil {
		return forwardMsg, err
	}
	resp, err := c.sendUniPacketAndWait("trpc.group.long_msg_interface.MsgService.SsoRecvLongMsg", pkt)
	if err != nil {
		return forwardMsg, err
	}
	pasted, err := messagePkt.ParseMultiMsgDownloadResp(resp)
	if err != nil {
		return forwardMsg, err
	}
	if pasted.Result == nil || pasted.Result.Payload == nil {
		return forwardMsg, errors.New("empty response data")
	}
	data := binary.GZipUncompress(pasted.Result.Payload)
	result := &message.LongMsgResult{}
	if err = proto.Unmarshal(data, result); err != nil {
		return forwardMsg, err
	}
	forwardMsg.Nodes = make([]*message2.ForwardNode, len(result.Action.ActionData.MsgBody))
	for idx, b := range result.Action.ActionData.MsgBody {
		isGroupMsg := b.ResponseHead.Grp != nil
		forwardMsg.Nodes[idx] = &message2.ForwardNode{
			SenderId:   int64(b.ResponseHead.FromUin),
			SenderName: b.ResponseHead.Forward.FriendName.Unwrap(),
			Time:       int32(b.ContentHead.TimeStamp.Unwrap()),
		}
		if isGroupMsg {
			forwardMsg.Nodes[idx].GroupId = int64(b.ResponseHead.Grp.GroupUin)
			grpMsg := message2.ParseGroupMessage(b)
			c.PreprocessGroupMessageEvent(grpMsg)
			forwardMsg.Nodes[idx].Message = grpMsg.Elements
		} else {
			prvMsg := message2.ParsePrivateMessage(b)
			c.PreprocessPrivateMessageEvent(prvMsg)
			forwardMsg.Nodes[idx].Message = prvMsg.Elements
		}
	}
	return forwardMsg, nil
}

// UploadForwardMsg 上传合并转发消息
// groupUin should be the group number where the uploader is located or 0 (c2c)
func (c *QQClient) UploadForwardMsg(forwardNodes []*message2.ForwardNode, groupUin uint32) (resId string, err error) {
	msgBody := c.BuildFakeMessage(forwardNodes)
	pkt, err := messagePkt.BuildMultiMsgUploadReq(c.GetUid(c.Uin), groupUin, msgBody)
	if err != nil {
		return "", err
	}
	resp, err := c.sendUniPacketAndWait("trpc.group.long_msg_interface.MsgService.SsoSendLongMsg", pkt)
	if err != nil {
		return "", err
	}
	pasted, err := messagePkt.ParseMultiMsgUploadResp(resp)
	if err != nil {
		return "", err
	}
	if pasted.Result == nil {
		return "", errors.New("empty response data")
	}
	return pasted.Result.ResId, nil
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
		reqUrl := fmt.Sprintf("https://qun.qq.com/cgi-bin/group_digest/digest_list?random=7800&X-CROSS-ORIGIN=fetch&group_code=%d&page_start=%d&page_limit=20&bkn=%d", groupUin, page, bkn)
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
		if err != nil {
			return essenceMsg, err
		}
		resp, err := c.SendRequestWithCookie(req, "qun.qq.com")
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
		respJson := gjson.ParseBytes(respData)
		if respJson.Get("retcode").Int() != 0 {
			return essenceMsg, fmt.Errorf("error code %d, %s", respJson.Get("retcode").Int(), respJson.Get("retmsg").String())
		}
		for _, v := range respJson.Get("data").Get("msg_list").Array() {
			var elements []message2.IMessageElement
			for _, e := range v.Get("msg_content").Array() {
				switch e.Get("msg_type").Int() {
				case 1:
					elements = append(elements, &message2.TextElement{Content: e.Get("text").String()})
				case 2:
					elements = append(elements, &message2.FaceElement{FaceID: uint16(e.Get("face_index").Int())})
				case 3:
					elements = append(elements, &message2.ImageElement{Url: e.Get("image_url").String()})
				case 4:
					elements = append(elements, &message2.FileElement{
						FileId:  e.Get("file_id").String(),
						FileUrl: e.Get("file_thumbnail_url").String(),
					})
				}
			}
			senderUin := uint32(v.Get("sender_uin").Int())
			senderInfo := c.GetCachedMemberInfo(senderUin, groupUin)
			essenceMsg = append(essenceMsg, &message2.GroupEssenceMessage{
				OperatorUin:  uint32(v.Get("add_digest_uin").Int()),
				OperatorUid:  c.GetUid(uint32(v.Get("add_digest_uin").Int())),
				OperatorTime: uint64(v.Get("add_digest_time").Int()),
				CanRemove:    v.Get("can_be_removed").Bool(),
				Message: &message2.GroupMessage{
					Id:         uint32(v.Get("msg_seq").Int()),
					InternalId: uint32(v.Get("msg_random").Int()),
					GroupUin:   grpInfo.GroupUin,
					GroupName:  grpInfo.GroupName,
					Sender: &message2.Sender{
						Uin:      senderUin,
						Uid:      c.GetUid(senderUin, groupUin),
						Nickname: senderInfo.MemberName,
						CardName: senderInfo.MemberCard,
					},
					Time:     uint32(v.Get("sender_time").Int()),
					Elements: elements,
				},
			})
		}
		if respJson.Get("data").Get("is_end").Bool() {
			break
		}
	}
	return essenceMsg, nil
}

// GetGroupHonorInfo 获取群荣誉信息
// reference https://github.com/Mrs4s/MiraiGo/blob/master/client/http_api.go
func (c *QQClient) GetGroupHonorInfo(groupCode int64, honorType entity.HonorType) (*entity.GroupHonorInfo, error) {
	ret := &entity.GroupHonorInfo{}
	honorRe := regexp.MustCompile(`window\.__INITIAL_STATE__\s*?=\s*?(\{.*\})`)
	reqUrl := fmt.Sprintf("https://qun.qq.com/interactive/honorlist?gc=%d&type=%d", groupCode, honorType)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return ret, err
	}
	resp, err := c.SendRequestWithCookie(req, "qun.qq.com")
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
func (c *QQClient) GetGroupNotice(groupCode int64) (l []*entity.GroupNoticeFeed, err error) {
	bkn, err := c.GetCsrfToken()
	if err != nil {
		return nil, err
	}
	v := url.Values{}
	v.Set("bkn", strconv.Itoa(bkn))
	v.Set("qid", strconv.FormatInt(groupCode, 10))
	v.Set("ft", "23")
	v.Set("ni", "1")
	v.Set("n", "1")
	v.Set("i", "1")
	v.Set("log_read", "1")
	v.Set("platform", "1")
	v.Set("s", "-1")
	v.Set("n", "20")
	req, _ := http.NewRequest(http.MethodGet, "https://web.qun.qq.com/cgi-bin/announce/get_t_list?"+v.Encode(), nil)
	resp, err := c.SendRequestWithCookie(req, "qun.qq.com")
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
	resp, err := c.SendRequestWithCookie(req, "qun.qq.com")
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
func (c *QQClient) AddGroupNoticeSimple(groupCode int64, text string) (noticeId string, err error) {
	bkn, err := c.GetCsrfToken()
	if err != nil {
		return "", err
	}
	body := fmt.Sprintf(`qid=%v&bkn=%v&text=%v&pinned=0&type=1&settings={"is_show_edit_card":0,"tip_window_type":1,"confirm_required":1}`, groupCode, bkn, url.QueryEscape(text))
	req, err := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/add_qun_notice?bkn="+strconv.Itoa(bkn), strings.NewReader(body))
	if err != nil {
		return "", err
	}
	resp, err := c.SendRequestWithCookie(req, "qun.qq.com")
	if err != nil {
		return "", err
	}
	var res entity.NoticeSendResp
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}
	_ = resp.Body.Close()
	return res.NoticeId, nil
}

// AddGroupNoticeWithPic 发群公告带图片
func (c *QQClient) AddGroupNoticeWithPic(groupCode int64, text string, pic []byte) (noticeId string, err error) {
	bkn, err := c.GetCsrfToken()
	if err != nil {
		return "", err
	}
	img, err := c.uploadGroupNoticePic(bkn, pic)
	if err != nil {
		return "", err
	}
	body := fmt.Sprintf(`qid=%v&bkn=%v&text=%v&pinned=0&type=1&settings={"is_show_edit_card":0,"tip_window_type":1,"confirm_required":1}&pic=%v&imgWidth=%v&imgHeight=%v`, groupCode, bkn, url.QueryEscape(text), img.ID, img.Width, img.Height)
	req, err := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/add_qun_notice?bkn="+strconv.Itoa(bkn), strings.NewReader(body))
	if err != nil {
		return "", err
	}
	resp, err := c.SendRequestWithCookie(req, "qun.qq.com")
	if err != nil {
		return "", err
	}
	var res entity.NoticeSendResp
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}
	_ = resp.Body.Close()
	return res.NoticeId, nil
}

// DelGroupNotice 删除群公告
func (c *QQClient) DelGroupNotice(groupCode int64, fid string) error {
	bkn, err := c.GetCsrfToken()
	if err != nil {
		return err
	}
	body := fmt.Sprintf(`fid=%s&qid=%v&bkn=%v&ft=23&op=1`, fid, groupCode, bkn)
	req, err := http.NewRequest(http.MethodPost, "https://web.qun.qq.com/cgi-bin/announce/del_feed", strings.NewReader(body))
	if err != nil {
		return err
	}
	_, err = c.SendRequestWithCookie(req, "qun.qq.com")
	if err != nil {
		return err
	}
	return nil
}
