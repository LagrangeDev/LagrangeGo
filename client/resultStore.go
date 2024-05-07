package client

// 借鉴 https://github.com/nonebot/adapter-onebot/blob/master/nonebot/adapters/onebot/store.py

import (
	"errors"
	"time"

	"github.com/RomiChan/syncx"

	"github.com/LagrangeDev/LagrangeGo/packets/wtlogin"
)

// var resultLogger = utils.GetLogger("resultstore")

// ssofetcher 灵感来源于ddl的onebot适配器
type ssofetcher syncx.Map[uint32, chan *wtlogin.SSOPacket]

func newssofetcher() *ssofetcher {
	return &ssofetcher{}
}

// ContainSeq 判断这个seq是否存在
func (s *ssofetcher) ContainSeq(seq int) bool {
	_, ok := (*syncx.Map[uint32, chan *wtlogin.SSOPacket])(s).Load(uint32(seq))
	return ok
}

// AddSeq 发消息的时候调用，把seq加到map里面
func (s *ssofetcher) AddSeq(seq int) {
	resultChan := make(chan *wtlogin.SSOPacket, 1)
	(*syncx.Map[uint32, chan *wtlogin.SSOPacket])(s).Store(uint32(seq), resultChan)
}

// DeleteSeq 删除seq
func (s *ssofetcher) DeleteSeq(seq int) {
	(*syncx.Map[uint32, chan *wtlogin.SSOPacket])(s).Delete(uint32(seq))
}

// AddResult 收到消息的时候调用，返回此seq是否存在，如果存在则存储数据
func (s *ssofetcher) AddResult(seq int, data *wtlogin.SSOPacket) bool {
	if resultChan, ok := (*syncx.Map[uint32, chan *wtlogin.SSOPacket])(s).Load(uint32(seq)); ok {
		resultChan <- data
		return true
	}
	return false
}

// Fecth 等待获取数据直到超时，这里找不到对应的seq会直接返回错误，务必在发包之前调用 AddSeq，如果发包出错可以 DeleteSeq
func (s *ssofetcher) Fecth(seq, timeout int) (*wtlogin.SSOPacket, error) {
	if resultChan, ok := (*syncx.Map[uint32, chan *wtlogin.SSOPacket])(s).Load(uint32(seq)); ok {
		// 确保读取完删除这个结果
		defer (*syncx.Map[uint32, chan *wtlogin.SSOPacket])(s).Delete(uint32(seq))
		select {
		case <-time.After(time.Duration(timeout) * time.Second):
			return nil, errors.New("fetch timeout")
		case result := <-resultChan:
			return result, nil
		}
	}
	return nil, errors.New("unknown seq")
}
