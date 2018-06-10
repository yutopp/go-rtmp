//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"log"

	"github.com/yutopp/go-rtmp/message"
)

var _ stateHandler = (*connectMessageHandler)(nil)

// publisherMessageHandler Handle message publisher messages.
//   transitions:
//     _              -> self
type publisherMessageHandler struct {
	conn *Conn
}

func (h *publisherMessageHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message) error {
	switch msg := msg.(type) {
	case *message.AudioMessage:
		return h.conn.handler.OnAudio(timestamp, msg.Payload)
	case *message.VideoMessage:
		return h.conn.handler.OnVideo(timestamp, msg.Payload)
	default:
		log.Printf("Unexpected message: %+v", msg)
		return nil
	}
}
