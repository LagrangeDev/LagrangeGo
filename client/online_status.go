package client

import "github.com/LagrangeDev/LagrangeGo/client/packets/pb/action"

var (
	StatusOnline                = action.SetStatus{Status: 10}                  // 在线
	StatusAway                  = action.SetStatus{Status: 30}                  // 离开
	StatusInvisible             = action.SetStatus{Status: 40}                  // 隐身
	StatusBusy                  = action.SetStatus{Status: 50}                  // 忙碌
	StatusQme                   = action.SetStatus{Status: 60}                  // Q我吧
	StatusDoNotDisturb          = action.SetStatus{Status: 70}                  // 请勿打扰
	StatusBattery               = action.SetStatus{Status: 10, ExtStatus: 1000} // 当前电量
	StatusStatusSignal          = action.SetStatus{Status: 10, ExtStatus: 1011} // 信号弱
	StatusStatusSleep           = action.SetStatus{Status: 10, ExtStatus: 1016} // 睡觉中
	StatusStatusStudy           = action.SetStatus{Status: 10, ExtStatus: 1018} // 学习中
	StatusWatchingTV            = action.SetStatus{Status: 10, ExtStatus: 1021} // 追剧中
	StatusTimi                  = action.SetStatus{Status: 10, ExtStatus: 1027} // timi中
	StatusListening             = action.SetStatus{Status: 10, ExtStatus: 1028} // 听歌中
	StatusWeather               = action.SetStatus{Status: 10, ExtStatus: 1030} // 今日天气
	StatusStayUp                = action.SetStatus{Status: 10, ExtStatus: 1032} // 熬夜中
	StatusLoving                = action.SetStatus{Status: 10, ExtStatus: 1051} // 恋爱中
	StatusIMFine                = action.SetStatus{Status: 10, ExtStatus: 1052} // 我没事
	StatusHiToFly               = action.SetStatus{Status: 10, ExtStatus: 1056} // 嗨到飞起
	StatusFullOfEnergy          = action.SetStatus{Status: 10, ExtStatus: 1058} // 元气满满
	StatusLeisurely             = action.SetStatus{Status: 10, ExtStatus: 1059} // 悠哉哉
	StatusBoredom               = action.SetStatus{Status: 10, ExtStatus: 1060} // 无聊中
	StatusIWantToBeQuiet        = action.SetStatus{Status: 10, ExtStatus: 1061} // 想静静
	StatusItSTooHardForMe       = action.SetStatus{Status: 10, ExtStatus: 1062} // 我太难了
	StatusItSHardToPutIntoWords = action.SetStatus{Status: 10, ExtStatus: 1063} // 一言难尽
	StatusGoodLuckKoi           = action.SetStatus{Status: 10, ExtStatus: 1071} // 好运锦鲤
	StatusTheWaterRetreats      = action.SetStatus{Status: 10, ExtStatus: 1201} // 水逆退散
	StatusTouchingTheFish       = action.SetStatus{Status: 10, ExtStatus: 1300} // 摸鱼中
	StatusStatusEmo             = action.SetStatus{Status: 10, ExtStatus: 1401} // emo中
	StatusItSHardToGetConfused  = action.SetStatus{Status: 10, ExtStatus: 2001} // 难得糊涂
	StatusGetOutOnTheWaves      = action.SetStatus{Status: 10, ExtStatus: 2003} // 出去浪
	StatusLoveYou               = action.SetStatus{Status: 10, ExtStatus: 2006} // 爱你
	StatusLiverWork             = action.SetStatus{Status: 10, ExtStatus: 2012} // 肝作业
	StatusIWantToOpenIt         = action.SetStatus{Status: 10, ExtStatus: 2013} // 我想开了
	StatusHollowedOut           = action.SetStatus{Status: 10, ExtStatus: 2014} // 被掏空
	StatusGoOnATrip             = action.SetStatus{Status: 10, ExtStatus: 2015} // 去旅行
	StatusTodayStepCount        = action.SetStatus{Status: 10, ExtStatus: 2017} // 今日步数
	StatusCrushed               = action.SetStatus{Status: 10, ExtStatus: 2019} // 我crush了
	StatusMovingBricks          = action.SetStatus{Status: 10, ExtStatus: 2023} // 搬砖中
	StatusTotherStar            = action.SetStatus{Status: 10, ExtStatus: 2025} // 一起元梦
	StatusSeekPartner           = action.SetStatus{Status: 10, ExtStatus: 2026} // 求星搭子
	StatusDoGood                = action.SetStatus{Status: 10, ExtStatus: 2047} // 做好事
)
