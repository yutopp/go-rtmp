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

type commonMessageHandler struct {
	conn *Conn
}

func (h *commonMessageHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	switch msg := msg.(type) {
	case *message.SetChunkSize:
		return h.conn.streamer.SetReadChunkSize(msg.ChunkSize)

	default:
		log.Printf("Unexpected message(common): Message = %+v", msg)
		return nil
	}
}
