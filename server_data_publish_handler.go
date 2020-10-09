//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/edgeware/go-rtmp/internal"
	"github.com/edgeware/go-rtmp/message"
)

var _ stateHandler = (*serverDataPublishHandler)(nil)

// serverDataPublishHandler Handle data messages from a publisher at server side.
//   transitions:
//     | _ -> self
type serverDataPublishHandler struct {
	sh *streamHandler
}

func (h *serverDataPublishHandler) onMessage(
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

func (h *serverDataPublishHandler) onData(
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

func (h *serverDataPublishHandler) onCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
	body interface{},
) error {
	return internal.ErrPassThroughMsg
}
