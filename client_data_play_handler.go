//
// Copyright (c) 2023- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/yutopp/go-rtmp/internal"
	"github.com/yutopp/go-rtmp/message"
)

// clientDataPlayHandler Handle control messages from a server in flow of connected.
//
//	transitions:
//	  | _         -> self
type clientDataPlayHandler struct {
	sh *streamHandler
}

var _ stateHandler = (*clientDataPlayHandler)(nil)

func (h *clientDataPlayHandler) onMessage(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
) error {
	switch msg := msg.(type) {
	case *message.AudioMessage:
		return h.sh.stream.userHandler().OnAudio(timestamp, msg.Payload)

	case *message.VideoMessage:
		return h.sh.stream.userHandler().OnVideo(timestamp, msg.Payload)

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *clientDataPlayHandler) onData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	body interface{},
) error {
	switch data := body.(type) {
	case *message.NetStreamSetDataFrame:
		return h.sh.stream.userHandler().OnSetDataFrame(timestamp, data)

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *clientDataPlayHandler) onCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
	body interface{},
) error {
	return internal.ErrPassThroughMsg
}
