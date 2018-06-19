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

var _ streamHandler = (*netStreamMessageHandler)(nil)

type netStreamState uint8

const (
	netStreamStateNotInAction netStreamState = iota
	netStreamStateHasPublisher
	netStreamStateHasPlayer
)

// netStreamMessageHandler Handle messages which are categorised as NetStream.
//   transitions:
//     = netStreamStateNotInAction
//       | "publish" -> netStreamStateHasPublisher
//       | "play"    -> netStreamStateHasPlayer (Not implemented)
//       | _         -> self
//
//     = netStreamStateHasPublisher
//       | _ -> self
//
//     = netStreamStateHasPlayer
//       | _ -> self
//
type netStreamMessageHandler struct {
	conn           *Conn
	state          netStreamState
	defaultHandler streamHandler
}

func (h *netStreamMessageHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	switch h.state {
	case netStreamStateNotInAction:
		return h.handleAction(chunkStreamID, timestamp, msg, stream)
	case netStreamStateHasPublisher:
		return h.handlePublisher(chunkStreamID, timestamp, msg, stream)
	default:
		panic("Unreachable!")
	}
}

func (h *netStreamMessageHandler) handleAction(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	var cmdMsgWrapper amfWrapperFunc
	var cmdMsg *message.CommandMessage
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		cmdMsgWrapper = amf0Wrapper
		cmdMsg = &msg.CommandMessage
		goto handleCommand

	case *message.CommandMessageAMF3:
		cmdMsgWrapper = amf0Wrapper
		cmdMsg = &msg.CommandMessage
		goto handleCommand

	default:
		log.Printf("Message unhandled(netStream): Message = %+v, State = %d", msg, h.state)
		return h.defaultHandler.Handle(chunkStreamID, timestamp, msg, stream)
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetStreamPublish:
		log.Printf("Publisher is comming: %+v", cmd)

		if err := h.conn.handler.OnPublish(timestamp, cmd); err != nil {
			return err
		}

		// TODO: fix
		m := cmdMsgWrapper(func(cmsg *message.CommandMessage) {
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
		if err := stream.Write(chunkStreamID, timestamp, m); err != nil {
			return err
		}
		log.Printf("Publisher accepted")

		h.state = netStreamStateHasPublisher

		return nil

	default:
		log.Printf("Unexpected command(netStream): Command = %+v, State = %d", cmdMsg, h.state)
		return nil
	}
}

func (h *netStreamMessageHandler) handlePublisher(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	switch msg := msg.(type) {
	case *message.AudioMessage:
		return h.conn.handler.OnAudio(timestamp, msg.Payload)
	case *message.VideoMessage:
		return h.conn.handler.OnVideo(timestamp, msg.Payload)
	default:
		log.Printf("Message unhandled(netStream): Message = %+v, State = %d", msg, h.state)
		return h.defaultHandler.Handle(chunkStreamID, timestamp, msg, stream)
	}
}
