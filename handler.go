//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

// Command =
//   | *message.NetConnectionConnect
//   | *message.NetConnectionCreateStream
//   | *message.NetStreamPublish
type Command interface{}

// Data =
//   | *message.NetStreamSetDataFrame
type Data interface{}

type Handler interface {
	OnCommand(timestamp uint32, cmd Command) error
	OnData(timestamp uint32, data Data) error
	OnAudio(timestamp uint32, payload []byte) error
	OnVideo(timestamp uint32, payload []byte) error
	OnClose()
}

type HandlerFactory func(conn *Conn) Handler

var _ Handler = (*NopHandler)(nil)

type NopHandler struct {
}

func (h *NopHandler) OnCommand(timestamp uint32, cmd Command) error {
	return nil
}

func (h *NopHandler) OnData(timestamp uint32, cmd Data) error {
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
