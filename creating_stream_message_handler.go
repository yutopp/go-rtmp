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

// creatingStreamMessageHandler Handle message until "createStream" message arived.
//   transitions:
//     "createStream" -> controllingStreamMessageHandler
//     _              -> self
type creatingStreamMessageHandler struct {
	conn *Conn
}

func (h *creatingStreamMessageHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message) error {
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		return h.handleCommand(chunkStreamID, timestamp, &msg.CommandMessage, amf0Wrapper)
	case *message.CommandMessageAMF3:
		return h.handleCommand(chunkStreamID, timestamp, &msg.CommandMessage, amf3Wrapper)
	default:
		log.Printf("Unexpected message(createStream): %+v", msg)
		return nil
	}
}

func (h *creatingStreamMessageHandler) handleCommand(chunkStreamID int, timestamp uint32, msg *message.CommandMessage, wrapper amfWrapperFunc) error {
	switch cmd := msg.Command.(type) {
	case *message.NetConnectionCreateStream:
		log.Printf("Stream creating...: %+v", cmd)

		// TODO: fix
		m := wrapper(func(cmsg *message.CommandMessage) {
			*cmsg = message.CommandMessage{
				CommandName:   "_result",
				TransactionID: msg.TransactionID,
				Command: &message.NetConnectionCreateStreamResult{
					StreamID: 20, // TODO: fix
				},
			}
		})
		if err := h.conn.streamer.Write(chunkStreamID, timestamp, m); err != nil {
			return err
		}
		log.Printf("Stream created")

		// next to controllingStreamHandler
		h.conn.stateHandler = &controllingStreamMessageHandler{
			conn: h.conn,
		}

		return nil

	default:
		log.Printf("Unexpected command(createStream): %+v", cmd)
		return nil
	}
}
