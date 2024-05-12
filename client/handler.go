package client

import "github.com/LagrangeDev/LagrangeGo/client/internal/network"

// handlerInfo from https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/client.go#L137
type handlerInfo struct {
	fun     func(i any, err error)
	dynamic bool
	params  network.RequestParams
}

func (h *handlerInfo) getParams() network.RequestParams {
	if h == nil {
		return nil
	}
	return h.params
}
