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

	"github.com/tidwall/gjson"

	"github.com/LagrangeDev/LagrangeGo/client/packets/album"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	lgrio "github.com/LagrangeDev/LagrangeGo/utils/io"
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
	respJSON := gjson.ParseBytes(respData)
	if respJSON.Get("ret").Int() != 0 {
		return nil, fmt.Errorf("error: ret:%d, msg:%s", respJSON.Get("ret").Int(), respJSON.Get("msg").Str)
	}
	albumList := respJSON.Get("data.album").Array()
	grpAlbumList := make([]*GroupAlbum, len(albumList))
	for i, v := range albumList {
		timeStr := v.Get("createtime").Str
		timeStamp, err := time.Parse(TimeLayout, timeStr)
		if err != nil {
			c.errorln(err)
			continue
		}
		grpAlbumList[i] = &GroupAlbum{
			GroupUin:       groupUin,
			Name:           v.Get("title").Str,
			ID:             v.Get("id").Str,
			Description:    v.Get("desc").Str,
			CoverURL:       v.Get("coverurl").Str,
			CreateNickname: v.Get("createnickname").Str,
			CreateUin:      uint32(v.Get("createuin").Int()),
			CreateTime:     timeStamp.Unix(),
		}
	}
	return grpAlbumList, nil
}

func (c *QQClient) GetGroupAlbumElem(ab *GroupAlbum) ([]*GroupAlbumElem, error) {
	var elem []*GroupAlbumElem
	pageInfo := ""
	for {
		pkt, err := album.BuildGetMediaListReq(c.Uin, ab.GroupUin, ab.ID, pageInfo)
		if err != nil {
			return nil, err
		}
		payload, err := c.sendUniPacketAndWait("QunAlbum.trpc.qzone.webapp_qun_media.QunMedia.GetMediaList", pkt)
		if err != nil {
			return nil, err
		}
		resp, err := album.ParseGetMediaListResp(payload)
		if err != nil {
			return nil, err
		}
		for _, v := range resp.Body.ElemInfo {
			if v.ImgInfo != nil {
				elem = append(elem, &GroupAlbumElem{
					photo: &GroupPhoto{
						ID:  v.ImgInfo.ImageID,
						URL: v.ImgInfo.ImgLinkInfo.ImageURL,
					},
					operatorUserInfo: &GroupAlbumElemUserInfo{
						UserNickName: v.UploaderInfo.UserNickName,
						UserUin:      v.UploaderInfo.UserUin,
					},
				})
			}
			if v.VideoInfo != nil {
				elem = append(elem, &GroupAlbumElem{
					video: &GroupVideo{
						ID:  v.VideoInfo.VideoID,
						URL: v.VideoInfo.VideoURL,
					},
					operatorUserInfo: &GroupAlbumElemUserInfo{
						UserNickName: v.UploaderInfo.UserNickName,
						UserUin:      v.UploaderInfo.UserUin,
					},
				})
			}
		}
		if resp.Body.PageInfo.IsNone() || gjson.Get(resp.Body.PageInfo.Unwrap(), "Loc.return_num").Int() == 0 {
			break
		}
		if resp.Body.ElemInfo == nil && resp.Body.ElemMetaInfo == nil {
			break
		}
		pageInfo = resp.Body.PageInfo.Unwrap()
	}
	return elem, nil
}

func (c *QQClient) buildUploadSessionReq(param *uploadSessionParam) (*groupAlbumUploadReq, int64, error) {
	timeStamp := time.Now().Unix()
	cookies, err := c.GetCookies("qzone.qq.com")
	if err != nil {
		return nil, timeStamp, err
	}
	reqBody := &groupAlbumUploadReq{
		ControlReq: []controlReq{
			{
				Uin: strconv.Itoa(int(c.Uin)),
				Token: token{
					Type:  4,
					Data:  cookies.PsKey,
					Appid: 5,
				},
				Appid:     param.UploadType.ReqSessionAppID,
				Checksum:  fmt.Sprintf("%x", param.CheckSum),
				CheckType: param.UploadType.ReqCheckType,
				FileLen:   param.Size,
				Env: env{
					Refer:      param.UploadType.ReqRefer,
					DeviceInfo: "h5",
				},
				Model: 0,
				Cmd:   param.UploadType.ReqSessionCmd,
			},
		},
	}
	//nolint
	switch param.UploadType.ResourceType {
	case ResourceTypePhoto:
		reqBody.ControlReq[0].BizReq = imgBizReq{
			commonBizReq: commonBizReq{
				SPicTitle:  param.FileName,
				SAlbumName: param.AlbumName,
				SAlbumID:   param.AlbumID,
				IBatchID:   int(timeStamp),
			},
			INeedFeeds:  1,
			IUploadTime: int(timeStamp),
			MapExt: mapExt{
				Appid:  "qun",
				Userid: strconv.Itoa(int(param.GroupUin)),
			},
		}
	case ResourceTypeVideo:
		reqBody.ControlReq[0].BizReq = videoBizReq{
			commonBizReq: commonBizReq{
				SPicTitle:   param.FileName,
				IUploadType: 3,
			},
			STitle:      param.FileName,
			IUploadTime: int(timeStamp),
			IPlayTime:   6077.000, // TODO: do we really need a real video length?
			IIsNew:      111,
			VideoExtInfo: videoExtInfo{
				VideoType: "3",
				DomainID:  "5",
				PhotoNum:  "0",
				VideoNum:  "1",
				QunID:     strconv.Itoa(int(param.GroupUin)),
			},
		}
	case ResourceTypeVideoThumbPhoto:
		reqBody.ControlReq[0].BizReq = videoThumbImgBizReq{
			imgBizReq: imgBizReq{
				commonBizReq: commonBizReq{
					SPicTitle:   param.FileName,
					SAlbumName:  param.AlbumName,
					SAlbumID:    param.AlbumID,
					IUploadType: 2,
					IBatchID:    int(param.VidTimeStamp), // parent video upload timestamp
				},
				INeedFeeds:  1,
				IUploadTime: int(timeStamp),
				MapExt: mapExt{
					Appid:  param.UploadType.ReqSessionAppID,
					Userid: strconv.Itoa(int(param.GroupUin)),
				},
			},
			MultiPicInfo: multiPicInfo{
				IBatUploadNum: 1,
			},
			STExtendInfo: extendInfo{
				MapParams: mapParams{
					Vid:      param.Vid, // parent video vid
					PhotoNum: "0",
					VideoNum: "1",
				},
			},
			STExternalMapExt: externalMapExt{
				IsClientUploadCover: "1",
				IsPicVideoMixFeeds:  "1",
			},
		}
	default:
		return nil, timeStamp, errors.New("unkown upload type")
	}
	return reqBody, timeStamp, nil
}

func (c *QQClient) getGroupAlbumUploadSession(param *uploadSessionParam) (*uploadOptions, int64, error) {
	reqBody, timeStamp, err := c.buildUploadSessionReq(param)
	if err != nil {
		return nil, timeStamp, err
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, timeStamp, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("https://h5.qzone.qq.com/webapp/json/sliceUpload/FileBatchControl/%x?g_tk=%d", param.CheckSum, param.GTK),
		bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, timeStamp, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return nil, timeStamp, err
	}
	respData, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, timeStamp, err
	}
	respJSON := gjson.ParseBytes(respData)
	if respJSON.Get("ret").Int() != 0 {
		return nil, timeStamp, fmt.Errorf("error: ret:%d, msg:%s", respJSON.Get("ret").Int(), respJSON.Get("msg").Str)
	}
	return &uploadOptions{
		Session:   respJSON.Get("data.session").Str,
		BlockSize: int(respJSON.Get("data.slice_size").Int()),
	}, timeStamp, nil
}

func (c *QQClient) uploadGroupAlbumBlock(typ uploadTypeParam, session string, seq, offset, chunkSize, totalSize, gtk int, chunk []byte, latest bool) (rsp *uploadBlockRsp, err error) {
	uploadURLCmd := lgrio.Ternary[string](typ.ResourceType == ResourceTypeVideo, "FileUploadVideo", "FileUpload")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("uin", strconv.Itoa(int(c.Uin)))
	_ = writer.WriteField("appid", typ.ReqSessionAppID)
	_ = writer.WriteField("session", session)
	_ = writer.WriteField("offset", strconv.Itoa(offset))
	part, err := writer.CreateFormFile("data", "blob")
	if err != nil {
		return nil, err
	}
	_, _ = part.Write(chunk)
	_ = writer.WriteField("checksum", "")
	_ = writer.WriteField("check_type", strconv.Itoa(typ.ReqCheckType))
	_ = writer.WriteField("retry", "0")
	_ = writer.WriteField("seq", strconv.Itoa(seq))
	_ = writer.WriteField("end", strconv.Itoa(offset+chunkSize))
	_ = writer.WriteField("cmd", "FileUpload")
	_ = writer.WriteField("slice_size", strconv.Itoa(chunkSize))
	_ = writer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://h5.qzone.qq.com/webapp/json/sliceUpload/%s?seq=%d&retry=0&offset=%d&end=%d&total=%d&type=form&g_tk=%d",
		uploadURLCmd, seq, offset, offset+chunkSize, totalSize, gtk), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.SendRequestWithCookie(req)
	if err != nil {
		return nil, err
	}
	c.debug("uploadGroupAlbumBlock %d | %d | %d | %d | %d", seq, offset, totalSize, chunkSize, resp.StatusCode)
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error resp code %d", resp.StatusCode)
	}
	respJSON := gjson.ParseBytes(respData)
	if respJSON.Get("ret").Int() != 0 {
		return nil, fmt.Errorf("error: ret:%d, msg:%s", respJSON.Get("ret").Int(), respJSON.Get("msg").Str)
	}
	if respJSON.Get("data.biz.sVid").String() != "" {
		c.debug("fetched vid %s", respJSON.Get("data.biz.sVid").String())
		return &uploadBlockRsp{
			VID: respJSON.Get("data.biz.sVid").Str,
		}, nil
	}
	if latest {
		return &uploadBlockRsp{
			SPhotoID: respJSON.Get("data.biz.sPhotoID").Str,
			SBURL:    respJSON.Get("data.biz.sBURL").Str,
		}, nil
	}
	return nil, nil
}

func (c *QQClient) doUploadGroupAlbumBlock(uos *uploadOptions, usp *uploadSessionParam, file io.ReadSeeker) (rsp *uploadBlockRsp, err error) {
	defer lgrio.CloseIO(file)
	offset, seq, latest := 0, 0, false
	chunk := make([]byte, uos.BlockSize)
	for {
		chunkSize, err := io.ReadFull(file, chunk)
		if chunkSize == 0 {
			break
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			chunk = chunk[:chunkSize]
			latest = true
		}
		rsp, err = c.uploadGroupAlbumBlock(usp.UploadType, uos.Session, seq, offset, chunkSize, usp.Size, usp.GTK, chunk, latest)
		if err != nil {
			return nil, err
		}
		if latest {
			return rsp, nil
		}
		seq++
		offset += chunkSize
	}
	return nil, errors.New("upload group album failed: unkown error")
}

func (c *QQClient) UploadGroupAlbumPhoto(parms *GroupAlbumUploadParam) (*GroupPhoto, error) {
	if parms == nil {
		return nil, errors.New("upload parms is nil")
	}
	cookie, err := c.GetCookies("qzone.qq.com")
	if err != nil {
		return nil, err
	}
	gtk := GTK(cookie.PsKey)
	md5, size := crypto.ComputeMd5AndLength(parms.Image)
	st := uploadTypeParam{ResourceTypePhoto, "qzone", "FileUpload", "qun", 0}
	usp := &uploadSessionParam{
		UploadType: st,
		GroupUin:   parms.GroupUin,
		FileName:   parms.FileName,
		CheckSum:   md5,
		Size:       int(size),
		AlbumID:    parms.AlbumID,
		AlbumName:  parms.AlbumName,
		GTK:        gtk,
	}
	session, _, err := c.getGroupAlbumUploadSession(usp)
	if err != nil {
		return nil, err
	}
	c.debug("upload group album photo start, session %s", session.Session)
	ubRsp, err := c.doUploadGroupAlbumBlock(session, usp, parms.Image)
	if err != nil {
		return nil, err
	}
	if ubRsp.SPhotoID == "" || ubRsp.SBURL == "" {
		return nil, errors.New("upload group album failed because ubRsp missing fields")
	}
	return &GroupPhoto{
		ID:  ubRsp.SPhotoID,
		URL: ubRsp.SBURL,
	}, nil
}

func (c *QQClient) UploadGroupAlbumVideo(parms *GroupAlbumUploadParam) (*GroupVideo, error) {
	if parms == nil {
		return nil, errors.New("upload parms is nil")
	}
	cookie, err := c.GetCookies("qzone.qq.com")
	if err != nil {
		return nil, err
	}
	gtk := GTK(cookie.PsKey)
	// upload video
	sha1, size := crypto.ComputeSha1AndLength(parms.Video)
	st := uploadTypeParam{ResourceTypeVideo, "qzone", "FileUploadVideo", "video_qun", 1}
	usp := &uploadSessionParam{
		UploadType: st,
		GroupUin:   parms.GroupUin,
		FileName:   parms.FileName,
		CheckSum:   sha1,
		Size:       int(size),
		AlbumID:    parms.AlbumID,
		AlbumName:  parms.AlbumName,
		GTK:        gtk,
	}
	session, timeStamp, err := c.getGroupAlbumUploadSession(usp)
	if err != nil {
		return nil, err
	}
	c.debug("upload group album video start, session %s", session.Session)
	uvbRsp, err := c.doUploadGroupAlbumBlock(session, usp, parms.Video)
	if err != nil {
		return nil, err
	}
	if uvbRsp.VID == "" {
		return nil, errors.New("upload failed because the vid is missing in the upload group video response")
	}
	// upload video thumbnail
	md5, size := crypto.ComputeMd5AndLength(parms.Thumbnail)
	st = uploadTypeParam{ResourceTypeVideoThumbPhoto, "huodong", "", "qun", 0}
	usp = &uploadSessionParam{
		UploadType:   st,
		GroupUin:     parms.GroupUin,
		FileName:     parms.FileName,
		CheckSum:     md5,
		Size:         int(size),
		AlbumID:      parms.AlbumID,
		AlbumName:    parms.AlbumName,
		VidTimeStamp: timeStamp,
		Vid:          uvbRsp.VID,
		GTK:          gtk,
	}
	session, _, err = c.getGroupAlbumUploadSession(usp)
	if err != nil {
		return nil, err
	}
	c.debug("upload group album video thumb start, session %s", session.Session)
	utbRsp, err := c.doUploadGroupAlbumBlock(session, usp, parms.Thumbnail)
	if err != nil {
		return nil, err
	}
	if utbRsp.SPhotoID == "" || utbRsp.SBURL == "" {
		return nil, errors.New("upload group album failed because missing video thumbnail fields")
	}
	// get video url
	updList, err := c.GetGroupAlbumElem(&GroupAlbum{
		GroupUin: parms.GroupUin,
		ID:       parms.AlbumID,
	})
	if err != nil {
		return nil, err
	}
	// simple loop, enough
	for _, elem := range updList {
		if elem.video != nil && elem.video.ID == uvbRsp.VID {
			return &GroupVideo{
				ID:  elem.video.ID,
				URL: elem.video.URL,
			}, nil
		}
	}
	return &GroupVideo{}, errors.New("upload success but cannot get video url")
}

type ResourceType int

const (
	ResourceTypeUnknown ResourceType = iota
	ResourceTypePhoto
	ResourceTypeVideoThumbPhoto
	ResourceTypeVideo
)

type (
	uploadSessionParam struct {
		UploadType   uploadTypeParam
		GroupUin     uint32
		FileName     string
		CheckSum     []byte
		Size         int
		AlbumID      string
		AlbumName    string
		VidTimeStamp int64
		Vid          string
		GTK          int
	}

	uploadTypeParam struct {
		ResourceType
		ReqRefer        string
		ReqSessionCmd   string
		ReqSessionAppID string
		ReqCheckType    int
	}
)

type (
	GroupAlbum struct {
		GroupUin       uint32
		Name           string
		ID             string
		Description    string
		CoverURL       string
		CreateNickname string
		CreateUin      uint32
		CreateTime     int64
	}

	GroupAlbumElem struct {
		photo            *GroupPhoto
		video            *GroupVideo
		operatorUserInfo *GroupAlbumElemUserInfo
	}

	GroupPhoto struct {
		ID  string
		URL string
	}

	GroupVideo struct {
		ID  string
		URL string
	}

	GroupAlbumElemUserInfo struct {
		UserNickName string
		UserUin      string
	}

	ImageFile struct {
		Image io.ReadSeeker
	}

	VideoFile struct {
		Thumbnail io.ReadSeeker
		Video     io.ReadSeeker
	}

	GroupAlbumUploadParam struct {
		ResourceType
		GroupUin                     uint32
		FileName, AlbumID, AlbumName string
		ImageFile
		VideoFile
	}

	uploadOptions struct {
		Session   string
		BlockSize int
	}

	// request upload session req
	groupAlbumUploadReq struct {
		ControlReq []controlReq `json:"control_req"`
	}

	controlReq struct {
		Uin       string      `json:"uin"`
		Token     token       `json:"token"`
		Appid     string      `json:"appid"`
		Checksum  string      `json:"checksum"`
		CheckType int         `json:"check_type"`
		FileLen   int         `json:"file_len"`
		Env       env         `json:"env"`
		Model     int         `json:"model"`
		BizReq    interface{} `json:"biz_req"`
		Session   string      `json:"session"`
		AsyUpload int         `json:"asy_upload"`
		Cmd       string      `json:"cmd"`
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

	commonBizReq struct {
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
	}

	imgBizReq struct {
		commonBizReq
		INeedFeeds  int    `json:"iNeedFeeds"`
		IUploadTime int    `json:"iUploadTime"`
		MapExt      mapExt `json:"mapExt"`
	}

	videoBizReq struct {
		commonBizReq
		STitle           string       `json:"sTitle"`
		SDesc            string       `json:"sDesc"`
		IFlag            int          `json:"iFlag"`
		IUploadTime      int          `json:"iUploadTime"`
		IPlayTime        float64      `json:"iPlayTime"`
		SCoverURL        string       `json:"sCoverUrl"`
		IIsNew           int          `json:"iIsNew"`
		IIsOriginalVideo int          `json:"iIsOriginalVideo"`
		IIsFormatF20     int          `json:"iIsFormatF20"`
		VideoExtInfo     videoExtInfo `json:"extend_info"`
	}

	videoThumbImgBizReq struct {
		imgBizReq
		MultiPicInfo     multiPicInfo   `json:"mutliPicInfo"` // 没错，tx拼错了
		STExtendInfo     extendInfo     `json:"stExtendInfo"`
		STExternalMapExt externalMapExt `json:"stExternalMapExt"`
		CameraMaker      string         `json:"sExif_CameraMaker"`
		CameraModel      string         `json:"sExif_CameraModel"`
		Time             string         `json:"sExif_Time"`
		LatitudeRef      string         `json:"sExif_LatitudeRef"`
		Latitude         string         `json:"sExif_Latitude"`
		LongitudeRef     string         `json:"sExif_LongitudeRef"`
		Longitude        string         `json:"sExif_Longitude"`
	}

	mapExt struct {
		Appid  string `json:"appid"`
		Userid string `json:"userid"`
	}

	videoExtInfo struct {
		VideoType string `json:"video_type"`
		DomainID  string `json:"domainid"`
		PhotoNum  string `json:"photo_num"`
		VideoNum  string `json:"video_num"`
		QunID     string `json:"qun_id"`
	}

	multiPicInfo struct {
		IBatUploadNum int `json:"iBatUploadNum"`
		ICurUpload    int `json:"iCurUpload"`
		ISuccNum      int `json:"iSuccNum"`
		IFailNum      int `json:"iFailNum"`
	}

	extendInfo struct {
		MapParams mapParams `json:"mapParams"`
	}

	mapParams struct {
		Vid      string `json:"vid"`
		PhotoNum string `json:"photo_num"`
		VideoNum string `json:"video_num"`
	}

	externalMapExt struct {
		IsClientUploadCover string `json:"is_client_upload_cover"`
		IsPicVideoMixFeeds  string `json:"is_pic_video_mix_feeds"`
	}

	uploadBlockRsp struct {
		SPhotoID string
		SBURL    string
		VID      string
	}
)
