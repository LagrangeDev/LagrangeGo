package client

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/events.go

import (
	"runtime/debug"
	"sync"
)

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
			client.error("event error: %v\n%s", pan, debug.Stack())
		}
	}()
	for _, handler := range handle.handlers {
		handler(client, event)
	}
}

type eventHandlers struct {
	subscribedEventHandlers     []any
	groupMessageReceiptHandlers sync.Map
}

func (c *QQClient) SubscribeEventHandler(handler any) {
	c.eventHandlers.subscribedEventHandlers = append(c.eventHandlers.subscribedEventHandlers, handler)
}

func (c *QQClient) onGroupMessageReceipt(id string, f ...func(*QQClient, *groupMessageReceiptEvent)) {
	if len(f) == 0 {
		c.eventHandlers.groupMessageReceiptHandlers.Delete(id)
		return
	}
	c.eventHandlers.groupMessageReceiptHandlers.LoadOrStore(id, f[0])
}
