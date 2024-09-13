package message

import (
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	message2 "github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

func BuildMultiMsgUploadReq(selfUid string, groupUin uint32, msg []*message.PushMsgBody) ([]byte, error) {
	longMsgResult := &message2.LongMsgResult{
		Action: &message2.LongMsgAction{
			ActionCommand: "MultiMsg",
			ActionData: &message2.LongMsgContent{
				MsgBody: msg,
			},
		},
	}
	longMsgResultData, _ := proto.Marshal(longMsgResult)
	payload := binary.GZipCompress(longMsgResultData)
	req := &message2.SendLongMsgReq{
		Info: &message2.SendLongMsgInfo{
			Type: utils.Ternary[uint32](groupUin == 0, 1, 3),
			Uid: &message2.LongMsgUid{
				Uid: utils.Ternary(groupUin == 0, proto.String(selfUid), proto.String(strconv.Itoa(int(groupUin)))),
			},
			GroupUin: proto.Uint32(groupUin),
			Payload:  payload,
		},
		Settings: &message2.LongMsgSettings{
			Field1: 4,
			Field2: 1,
			Field3: 7,
			Field4: 0,
		},
	}
	return proto.Marshal(req)
}

func ParseMultiMsgUploadResp(data []byte) (resp *message2.SendLongMsgResp, err error) {
	resp = &message2.SendLongMsgResp{}
	if err = proto.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
