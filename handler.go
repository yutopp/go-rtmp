//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/yutopp/go-rtmp/message"
)

type Handler interface {
	OnConnect(timestamp uint32, cmd *message.NetConnectionConnect) error
	OnPublish(timestamp uint32, cmd *message.NetStreamPublish) error
	OnPlay(timestamp uint32, args []interface{}) error
	OnSetDataFrame(timestamp uint32, payload []byte) error
	OnAudio(timestamp uint32, payload []byte) error
	OnVideo(timestamp uint32, payload []byte) error
	OnClose()
}

type HandlerFactory func() Handler

var _ Handler = (*NopHandler)(nil)

type NopHandler struct {
}

func (h *NopHandler) OnConnect(timestamp uint32, cmd *message.NetConnectionConnect) error {
	return nil
}

func (h *NopHandler) OnPublish(timestamp uint32, cmd *message.NetStreamPublish) error {
	return nil
}

func (h *NopHandler) OnPlay(timestamp uint32, args []interface{}) error {
	return nil
}

func (h *NopHandler) OnSetDataFrame(timestamp uint32, payload []byte) error {
	return nil
}

func (h *NopHandler) OnAudio(timestamp uint32, payload []byte) error {
	return nil
}

func (h *NopHandler) OnVideo(timestamp uint32, payload []byte) error {
	return nil
}

func (h *NopHandler) OnClose() {
}
