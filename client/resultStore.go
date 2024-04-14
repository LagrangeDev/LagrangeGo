package client

import (
	"errors"
	"sync"
	"time"

	"github.com/LagrangeDev/LagrangeGo/packets/wtlogin"

	"github.com/LagrangeDev/LagrangeGo/utils"
)

var resultLogger = utils.GetLogger("resultstore")

// ResultStore 灵感来源于ddl的onebot适配器
type ResultStore struct {
	result sync.Map
}

func NewResultStore() *ResultStore {
	return &ResultStore{}
}

// ContainSeq 判断这个seq是否存在
func (s *ResultStore) ContainSeq(seq int) bool {
	_, ok := s.result.Load(seq)
	return ok
}

// AddSeq 发消息的时候调用，把seq加到map里面
func (s *ResultStore) AddSeq(seq int) {
	resultChan := make(chan *wtlogin.SSOPacket, 1)
	s.result.Store(seq, resultChan)
}

// DeleteSeq 删除seq
func (s *ResultStore) DeleteSeq(seq int) {
	s.result.Delete(seq)
}

// AddResult 收到消息的时候调用，返回此seq是否存在，如果存在则存储数据
func (s *ResultStore) AddResult(seq int, data *wtlogin.SSOPacket) bool {
	if resultChan, ok := s.result.Load(seq); ok {
		resultChan.(chan *wtlogin.SSOPacket) <- data
		return true
	}
	return false
}

// Fecth 等待获取数据直到超时，这里找不到对应的seq会直接返回错误，务必在发包之前调用 AddSeq，如果发包出错可以 DeleteSeq
func (s *ResultStore) Fecth(seq, timeout int) (*wtlogin.SSOPacket, error) {
	if resultChan, ok := s.result.Load(seq); ok {
		// 确保读取完删除这个结果
		defer s.result.Delete(seq)
		select {
		case <-time.After(time.Duration(timeout) * time.Second):
			return nil, errors.New("fetch timeout")
		case result := <-(resultChan.(chan *wtlogin.SSOPacket)):
			return result, nil
		}
	}
	return nil, errors.New("unknown seq")
}
