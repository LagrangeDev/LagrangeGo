package entity

// reference https://github.com/Mrs4s/MiraiGo/blob/master/client/http_api.go

type (
	GroupNoticeRsp struct {
		Feeds []*GroupNoticeFeed `json:"feeds"`
		Inst  []*GroupNoticeFeed `json:"inst"`
	}

	GroupNoticeFeed struct {
		NoticeID    string `json:"fid"`
		SenderID    uint32 `json:"u"`
		PublishTime uint64 `json:"pubt"`
		Message     struct {
			Text   string        `json:"text"`
			Images []NoticeImage `json:"pics"`
		} `json:"msg"`
	}

	NoticePicUpResponse struct {
		ErrorCode    int    `json:"ec"`
		ErrorMessage string `json:"em"`
		ID           string `json:"id"`
	}

	NoticeImage struct {
		Height string `json:"h"`
		Width  string `json:"w"`
		ID     string `json:"id"`
	}

	NoticeSendResp struct {
		NoticeID string `json:"new_fid"`
	}
)
