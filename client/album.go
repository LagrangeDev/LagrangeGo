package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/LagrangeDev/LagrangeGo/utils"

	"github.com/LagrangeDev/LagrangeGo/utils/crypto"

	"github.com/tidwall/gjson"
)

const TimeLayout = "2006-01-02 15:04:05"

func (c *QQClient) GetGroupAlbum(groupUin uint32) ([]*GroupAlbum, error) {
	gtk, err := c.GetCsrfToken()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://h5.qzone.qq.com/proxy/domain/u.photo.qzone.qq.com/cgi-bin/upp/qun_list_album_v2?random=7570&g_tk=%d&format=json&inCharset=utf-8&outCharset=utf-8&qua=V1_IPH_SQ_6.2.0_0_HDBM_T&cmd=qunGetAlbumList&qunId=%d&qunid=%d&start=0&num=1000&uin=%d&getMemberRole=0",
			gtk, groupUin, groupUin, c.Uin), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error resp code %d", resp.StatusCode)
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	respJson := gjson.ParseBytes(respData)
	if respJson.Get("ret").Int() != 0 {
		return nil, fmt.Errorf("error: ret:%d, msg:%s", respJson.Get("ret").Int(), respJson.Get("msg").Str)
	}
	albumList := respJson.Get("data.album").Array()
	grpAlbumList := make([]*GroupAlbum, len(albumList))
	for i, v := range albumList {
		timeStr := v.Get("createtime").Str
		timeStamp, err := time.Parse(TimeLayout, timeStr)
		if err != nil {
			fmt.Println(err)
		}
		grpAlbumList[i] = &GroupAlbum{
			Name:           v.Get("title").Str,
			ID:             v.Get("id").Str,
			Description:    v.Get("desc").Str,
			CoverUrl:       v.Get("coverurl").Str,
			CreateNickname: v.Get("createnickname").Str,
			CreateUin:      uint32(v.Get("createuin").Int()),
			CreateTime:     timeStamp.Unix(),
		}
	}
	return grpAlbumList, nil
}

func (c *QQClient) getGroupAlbunUploadSession(groupUin uint32, fileName, albumId, albumName string, md5 []byte, size, gtk int) (*uploadOptions, error) {
	cookies, err := c.GetCookies("qzone.qq.com")
	if err != nil {
		return nil, err
	}
	timeStamp := time.Now().Unix()
	reqBody := &groupAlbunUploadReq{
		ControlReq: []controlReq{
			{
				Uin: strconv.Itoa(int(c.Uin)),
				Token: token{
					Type:  4,
					Data:  cookies.PsKey,
					Appid: 5,
				},
				Appid:    "qun",
				Checksum: fmt.Sprintf("%x", md5),
				FileLen:  size,
				Env: env{
					Refer:      "qzone",
					DeviceInfo: "h5",
				},
				BizReq: bizReq{
					SPicTitle:   fileName,
					SPicDesc:    "",
					SAlbumName:  albumName,
					SAlbumID:    albumId,
					IBatchID:    int(timeStamp),
					INeedFeeds:  1,
					IUploadTime: int(timeStamp),
					MapExt: mapExt{
						Appid:  "qun",
						Userid: strconv.Itoa(int(groupUin)),
					},
				},
				Cmd: "FileUpload",
			},
		},
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("https://h5.qzone.qq.com/webapp/json/sliceUpload/FileBatchControl/%x?g_tk=%d", md5, gtk),
		bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	respJson := gjson.ParseBytes(respData)
	if respJson.Get("ret").Int() != 0 {
		return nil, fmt.Errorf("error: ret:%d, msg:%s", respJson.Get("ret").Int(), respJson.Get("msg").Str)
	}
	return &uploadOptions{
		Session:   respJson.Get("data.session").Str,
		BlockSize: int(respJson.Get("data.slice_size").Int()),
	}, nil
}

func (c *QQClient) uploadGroupAlbumBlock(session string, seq, offset, chunkSize, totlaSize, gtk int, chunk []byte, latest bool) (*GroupPhoto, error) {
	c.debug("seq:%d,offset:%d,chunksize:%d,totalsize:%d", seq, offset, chunkSize, totlaSize)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("uin", strconv.Itoa(int(c.Uin)))
	_ = writer.WriteField("appid", "qun")
	_ = writer.WriteField("session", session)
	_ = writer.WriteField("offset", strconv.Itoa(offset))
	part, err := writer.CreateFormFile("data", "blob")
	if err != nil {
		return nil, err
	}
	_, _ = part.Write(chunk)
	_ = writer.WriteField("checksum", "")
	_ = writer.WriteField("check_type", "0")
	_ = writer.WriteField("retry", "0")
	_ = writer.WriteField("seq", strconv.Itoa(seq))
	_ = writer.WriteField("end", strconv.Itoa(offset+chunkSize))
	_ = writer.WriteField("cmd", "FileUpload")
	_ = writer.WriteField("slice_size", strconv.Itoa(chunkSize))
	_ = writer.WriteField("biz_req.iUploadType", "0")
	_ = writer.Close()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://h5.qzone.qq.com/webapp/json/sliceUpload/FileUpload?seq=%d&retry=0&offset=%d&end=%d&total=%d&type=form&g_tk=%d",
		seq, offset, offset+chunkSize, totlaSize, gtk), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.StatusCode)
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error resp code %d", resp.StatusCode)
	}
	respJson := gjson.ParseBytes(respData)
	if respJson.Get("ret").Int() != 0 {
		return nil, fmt.Errorf("error: ret:%d, msg:%s", respJson.Get("ret").Int(), respJson.Get("msg").Str)
	}
	if latest {
		return &GroupPhoto{
			ID:  respJson.Get("data.biz.sPhotoID").Str,
			Url: respJson.Get("data.biz.sBURL").Str,
		}, nil
	}
	return nil, nil
}

func (c *QQClient) UploadGroupAlbum(parms *GroupAlbumUploadParm) (*GroupPhoto, error) {
	if parms == nil {
		return nil, errors.New("upload parms is nil")
	}
	defer utils.CloseIO(parms.File)
	cookie, err := c.GetCookies("qzone.qq.com")
	if err != nil {
		return nil, err
	}
	gtk := GTK(cookie.PsKey)
	md5, size := crypto.ComputeMd5AndLength(parms.File)
	session, err := c.getGroupAlbunUploadSession(parms.GroupUin, parms.FileName, parms.AlbumId, parms.AlbumName, md5, int(size), gtk)
	if err != nil {
		return nil, err
	}
	c.debug("upload group album session %s", session.Session)
	offset := 0
	seq := 0
	latest := false
	chunk := make([]byte, session.BlockSize)
	for {
		chunkSize, err := io.ReadFull(parms.File, chunk)
		if chunkSize == 0 {
			break
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			chunk = chunk[:chunkSize]
			latest = true
		}
		photo, err := c.uploadGroupAlbumBlock(session.Session, seq, offset, chunkSize, int(size), gtk, chunk, latest)
		if err != nil {
			return nil, err
		}
		if latest {
			return photo, err
		}
		seq += 1
		offset += chunkSize
	}
	return nil, errors.New("upload group album failed: unkown error")
}

type (
	GroupAlbum struct {
		Name           string
		ID             string
		Description    string
		CoverUrl       string
		CreateNickname string
		CreateUin      uint32
		CreateTime     int64
	}

	GroupPhoto struct {
		ID  string
		Url string
	}

	GroupAlbumUploadParm struct {
		GroupUin                     uint32
		FileName, AlbumId, AlbumName string
		File                         io.ReadSeeker
	}

	uploadOptions struct {
		Session   string
		BlockSize int
	}

	// request upload session req
	groupAlbunUploadReq struct {
		ControlReq []controlReq `json:"control_req"`
	}

	controlReq struct {
		Uin       string `json:"uin"`
		Token     token  `json:"token"`
		Appid     string `json:"appid"`
		Checksum  string `json:"checksum"`
		CheckType int    `json:"check_type"`
		FileLen   int    `json:"file_len"`
		Env       env    `json:"env"`
		Model     int    `json:"model"`
		BizReq    bizReq `json:"biz_req"`
		Session   string `json:"session"`
		AsyUpload int    `json:"asy_upload"`
		Cmd       string `json:"cmd"`
	}

	token struct {
		Type  int    `json:"type"`
		Data  string `json:"data"`
		Appid int    `json:"appid"`
	}

	env struct {
		Refer      string `json:"refer"`
		DeviceInfo string `json:"deviceInfo"`
	}

	bizReq struct {
		SPicTitle    string `json:"sPicTitle"`
		SPicDesc     string `json:"sPicDesc"`
		SAlbumName   string `json:"sAlbumName"`
		SAlbumID     string `json:"sAlbumID"`
		IAlbumTypeID int    `json:"iAlbumTypeID"`
		IBitmap      int    `json:"iBitmap"`
		IUploadType  int    `json:"iUploadType"`
		IUpPicType   int    `json:"iUpPicType"`
		IBatchID     int    `json:"iBatchID"`
		SPicPath     string `json:"sPicPath"`
		IPicWidth    int    `json:"iPicWidth"`
		IPicHight    int    `json:"iPicHight"`
		IWaterType   int    `json:"iWaterType"`
		IDistinctUse int    `json:"iDistinctUse"`
		INeedFeeds   int    `json:"iNeedFeeds"`
		IUploadTime  int    `json:"iUploadTime"`
		MapExt       mapExt `json:"mapExt"`
	}

	mapExt struct {
		Appid  string `json:"appid"`
		Userid string `json:"userid"`
	}
)
