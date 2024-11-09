package album

import (
	"errors"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/album"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildGetMediaListReq(selfUin uint32, groupUin uint32, albumID string, pageInfo string) ([]byte, error) {
	return proto.Marshal(&album.QzoneGetMediaList{
		Field1: 0,
		Field2: "h5_test",
		Field3: "h5_test",
		Field4: &album.QzoneGetMediaList_F4{
			GroupID:  strconv.Itoa(int(groupUin)),
			AlbumID:  albumID,
			Field3:   0,
			Field4:   "",
			PageInfo: pageInfo,
		},
		UinTimeStamp: utils.UinTimestamp(selfUin),
		Field10: &album.QzoneGetMediaList_F10{
			AppIdFlag:  "fc-appid",
			AppIdValue: "100",
		},
	})
}

func ParseGetMediaListResp(data []byte) (resp *album.QzoneGetMediaList_Response, err error) {
	resp = &album.QzoneGetMediaList_Response{}
	if err = proto.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	if resp.ErrorCode.IsSome() && resp.ErrorMsg.IsSome() {
		return nil, errors.New(resp.ErrorMsg.Unwrap())
	}
	return resp, nil
}
