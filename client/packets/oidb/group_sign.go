package oidb

import (
	"errors"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

// 群打卡

type BotGroupClockInResult struct {
	Title          string // 今日已成功打卡
	KeepDayText    string // 已打卡N天
	GroupRankText  string // 群内排名第N位
	ClockInUtcTime int64  // 打卡时间
	DetailUrl      string // Detail info url https://qun.qq.com/v2/signin/detail?...
}

func BuildGroupSignPacket(botUin, groupUin uint32, appVersion string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0XEB7_1_ReqBody{
		SignInWriteReq: &oidb.StSignInWriteReq{
			Uin:        strconv.Itoa(int(botUin)),
			GroupUin:   strconv.Itoa(int(groupUin)),
			AppVersion: appVersion,
		},
	}
	return BuildOidbPacket(0xEB7, 1, body, false, false)
}

func ParseGroupSignResp(data []byte) (*BotGroupClockInResult, error) {
	var rsp oidb.OidbSvcTrpcTcp0XEB7_1_RspBody
	_, err := ParseOidbPacket(data, &rsp)
	if err != nil {
		return nil, err
	}

	if rsp.SignInWriteRsp == nil || rsp.SignInWriteRsp.DoneInfo == nil {
		return nil, errors.New("SignInWriteRsp or SignInWriteRsp.DoneInfo nil")
	}

	ret := &BotGroupClockInResult{
		Title:       rsp.SignInWriteRsp.DoneInfo.Title,
		KeepDayText: rsp.SignInWriteRsp.DoneInfo.KeepDayText,
		DetailUrl:   rsp.SignInWriteRsp.DoneInfo.DetailUrl,
	}
	if size := len(rsp.SignInWriteRsp.DoneInfo.ClockInInfo); size > 0 {
		ret.GroupRankText = rsp.SignInWriteRsp.DoneInfo.ClockInInfo[0]
		if size > 1 {
			ret.ClockInUtcTime, _ = strconv.ParseInt(rsp.SignInWriteRsp.DoneInfo.ClockInInfo[1], 10, 64)
		}
	}
	return ret, nil
}
