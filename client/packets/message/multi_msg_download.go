package message

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
)

func BuildMultiMsgDownloadReq(uid string, resId string) ([]byte, error) {
	return proto.Marshal(&message.RecvLongMsgReq{
		Info: &message.RecvLongMsgInfo{
			Uid: &message.LongMsgUid{
				Uid: proto.String(uid),
			},
			ResId:   proto.String(resId),
			Acquire: true,
		},
		Settings: &message.LongMsgSettings{
			Field1: 2,
			Field2: 0,
			Field3: 0,
			Field4: 0,
		},
	})
}

func ParseMultiMsgDownloadResp(data []byte) (resp *message.RecvLongMsgResp, err error) {
	resp = &message.RecvLongMsgResp{}
	if err = proto.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
