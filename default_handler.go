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

var _ Handler = (*DefaultHandler)(nil)

type DefaultHandler struct {
}

func (h *DefaultHandler) OnServe() {
}

func (h *DefaultHandler) OnConnect(timestamp uint32, cmd *message.NetConnectionConnect) error {
	return nil
}

func (h *DefaultHandler) OnCreateStream(timestamp uint32, cmd *message.NetConnectionCreateStream) error {
	return nil
}

func (h *DefaultHandler) OnReleaseStream(timestamp uint32, cmd *message.NetConnectionReleaseStream) error {
	return nil
}

func (h *DefaultHandler) OnDeleteStream(timestamp uint32, cmd *message.NetStreamDeleteStream) error {
	return nil
}

func (h *DefaultHandler) OnPublish(timestamp uint32, cmd *message.NetStreamPublish) error {
	return nil
}

func (h *DefaultHandler) OnPlay(timestamp uint32, cmd *message.NetStreamPlay) error {
	return nil
}

func (h *DefaultHandler) OnFCPublish(timestamp uint32, cmd *message.NetStreamFCPublish) error {
	return nil
}

func (h *DefaultHandler) OnFCUnpublish(timestamp uint32, cmd *message.NetStreamFCUnpublish) error {
	return nil
}

func (h *DefaultHandler) OnSetDataFrame(timestamp uint32, data *message.NetStreamSetDataFrame) error {
	return nil
}

func (h *DefaultHandler) OnAudio(timestamp uint32, payload []byte) error {
	return nil
}

func (h *DefaultHandler) OnVideo(timestamp uint32, payload []byte) error {
	return nil
}

func (h *DefaultHandler) OnUnknownMessage(timestamp uint32, cmd message.Message) error {
	return nil
}

func (h *DefaultHandler) OnUnknownCommandMessage(timestamp uint32, cmd *message.CommandMessage) error {
	return nil
}

func (h *DefaultHandler) OnUnknownDataMessage(timestamp uint32, cmd *message.DataMessage) error {
	return nil
}

func (h *DefaultHandler) OnClose() {
}
