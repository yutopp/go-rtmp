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

var _ stateHandler = (*controllingStreamMessageHandler)(nil)

// controllingStreamMessageHandler Handle message until "publish" or "play" message is arived.
//   transitions:
//     "publish" -> publisherMessageHandler
//     "play"    -> playerMessageHandler (Not implemented)
//     _         -> self
type controllingStreamMessageHandler struct {
	conn *Conn
}

func (h *controllingStreamMessageHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message) error {
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		return h.handleCommand(chunkStreamID, timestamp, &msg.CommandMessage, amf0Wrapper)
	case *message.CommandMessageAMF3:
		return h.handleCommand(chunkStreamID, timestamp, &msg.CommandMessage, amf3Wrapper)
	default:
		log.Printf("Unexpected message: %+v", msg)
		return nil
	}
}

func (h *controllingStreamMessageHandler) handleCommand(chunkStreamID int, timestamp uint32, msg *message.CommandMessage, wrapper amfWrapperFunc) error {
	switch cmd := msg.Command.(type) {
	case *message.NetStreamPublish:
		log.Printf("Publisher is comming: %+v", cmd)

		if err := h.conn.handler.OnPublish(timestamp, cmd); err != nil {
			return err
		}

		// TODO: fix
		m := wrapper(func(cmsg *message.CommandMessage) {
			*cmsg = message.CommandMessage{
				CommandName:   "onStatus",
				TransactionID: 0,
				Command: &message.NetStreamOnStatus{
					InfoObject: message.NetStreamOnStatusInfoObject{
						Level:       "status",
						Code:        "NetStream.Publish.Start",
						Description: "yoyo",
					},
				},
			}
		})
		if err := h.conn.streamer.Write(chunkStreamID, timestamp, m); err != nil {
			return err
		}
		log.Printf("Publisher accepted")

		// next to publisherStreamHandler
		h.conn.stateHandler = &publisherMessageHandler{
			conn: h.conn,
		}

		return nil

	default:
		log.Printf("Unexpected command: %+v", cmd)
		return nil
	}
}
