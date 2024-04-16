package client

import (
	"reflect"
	"runtime/debug"
	"sync"

	"github.com/LagrangeDev/LagrangeGo/event"
	"github.com/LagrangeDev/LagrangeGo/message"

	"github.com/LagrangeDev/LagrangeGo/utils"
)

var eventLogger = utils.GetLogger("event")

// protected all EventHandle, since write is very rare, use
// only one lock to save memory
var eventMu sync.RWMutex

type EventHandle[T any] struct {
	// QQClient?
	handlers []func(client *QQClient, event T)
}

func (handle *EventHandle[T]) Subscribe(handler func(client *QQClient, event T)) {
	eventMu.Lock()
	defer eventMu.Unlock()
	// shrink the slice
	newHandlers := make([]func(client *QQClient, event T), len(handle.handlers)+1)
	copy(newHandlers, handle.handlers)
	newHandlers[len(handle.handlers)] = handler
	handle.handlers = newHandlers
}

func (handle *EventHandle[T]) dispatch(client *QQClient, event T) {
	eventMu.RLock()
	defer func() {
		eventMu.RUnlock()
		if pan := recover(); pan != nil {
			eventLogger.Errorf("event error: %v\n%s", pan, debug.Stack())
		}
	}()
	for _, handler := range handle.handlers {
		handler(client, event)
	}
}

// OnEvent 事件响应，耗时操作，需提交协程处理
func OnEvent(client *QQClient, msg any) {
	switch msg := msg.(type) {
	case *message.PrivateMessage:
		client.PrivateMessageEvent.dispatch(client, msg)
	case *message.GroupMessage:
		client.GroupMessageEvent.dispatch(client, msg)
	case *message.TempMessage:
		client.TempMessageEvent.dispatch(client, msg)
	case *event.GroupMemberJoinRequest:
		client.GroupInvitedEvent.dispatch(client, msg)
	case *event.GroupMemberJoined:
		if client.uin == msg.Uin {
			client.GroupJoinEvent.dispatch(client, msg)
		} else {
			client.GroupMemberJoinEvent.dispatch(client, msg)
		}
	case *event.GroupMemberQuit:
		if client.uin == msg.Uin {
			client.GroupLeaveEvent.dispatch(client, msg)
		} else {
			client.GroupMemberLeaveEvent.dispatch(client, msg)
		}
	case nil:
		networkLogger.Errorf("nil event msg, ignore")
	default:
		networkLogger.Warningf("Unknown event type: %v, ignore", reflect.TypeOf(msg).String())
	}
}
