//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/pkg/errors"

	"github.com/yutopp/go-rtmp/internal"
	"github.com/yutopp/go-rtmp/message"
)

var _ messageHandler = (*serverDataPublishHandler)(nil)

// serverDataPublishHandler Handle data messages from a publiser at server side.
//   transitions:
//     | _ -> self
type serverDataPublishHandler struct {
	entry *entryHandler
}

func (h *serverDataPublishHandler) Handle(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
	stream *Stream,
) error {
	switch msg := msg.(type) {
	case *message.AudioMessage:
		return h.entry.conn.handler.OnAudio(timestamp, msg.Payload)

	case *message.VideoMessage:
		return h.entry.conn.handler.OnVideo(timestamp, msg.Payload)

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *serverDataPublishHandler) HandleCommand(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	cmdMsg *message.CommandMessage,
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverDataPublishHandler) HandleData(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	dataMsg *message.DataMessage,
	stream *Stream,
) error {
	switch dataMsg.Name {
	case "@setDataFrame":
		df := dataMsg.Data.(*message.NetStreamSetDataFrame)
		if df == nil {
			return errors.New("setDataFrame has nil value")
		}
		return h.entry.conn.handler.OnSetDataFrame(timestamp, df)

	default:
		return internal.ErrPassThroughMsg
	}
}
