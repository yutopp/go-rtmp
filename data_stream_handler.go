//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/sirupsen/logrus"

	"github.com/yutopp/go-rtmp/message"
)

var _ streamHandler = (*dataStreamHandler)(nil)

type dataStreamState uint8

const (
	dataStreamStateNotInAction dataStreamState = iota
	dataStreamStateHasPublisher
	dataStreamStateHasPlayer
)

// dataStreamHandler Handle messages which are categorised as NetStream.
//   transitions:
//     = dataStreamStateNotInAction
//       | "publish" -> dataStreamStateHasPublisher
//       | "play"    -> dataStreamStateHasPlayer (Not implemented)
//       | _         -> self
//
//     = dataStreamStateHasPublisher
//       | _ -> self
//
//     = dataStreamStateHasPlayer
//       | _ -> self
//
type dataStreamHandler struct {
	conn           *Conn
	state          dataStreamState
	defaultHandler streamHandler

	logger *logrus.Entry
}

func (h *dataStreamHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	switch h.state {
	case dataStreamStateNotInAction:
		return h.handleAction(chunkStreamID, timestamp, msg, stream)
	case dataStreamStateHasPublisher:
		return h.handlePublisher(chunkStreamID, timestamp, msg, stream)
	default:
		panic("Unreachable!")
	}
}

func (h *dataStreamHandler) handleAction(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
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
		h.logger.Printf("Message unhandled(netStream): Message = %+v, State = %d", msg, h.state)
		return h.defaultHandler.Handle(chunkStreamID, timestamp, msg, stream)
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetStreamPublish:
		h.logger.Printf("Publisher is comming: %+v", cmd)

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
		h.logger.Printf("Publisher accepted")

		h.state = dataStreamStateHasPublisher

		return nil

	default:
		h.logger.Printf("Unexpected command(netStream): Command = %+v, State = %d", cmdMsg, h.state)
		return nil
	}
}

func (h *dataStreamHandler) handlePublisher(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	switch msg := msg.(type) {
	case *message.AudioMessage:
		return h.conn.handler.OnAudio(timestamp, msg.Payload)
	case *message.VideoMessage:
		return h.conn.handler.OnVideo(timestamp, msg.Payload)
	default:
		h.logger.Printf("Message unhandled(netStream): Message = %+v, State = %d", msg, h.state)
		return h.defaultHandler.Handle(chunkStreamID, timestamp, msg, stream)
	}
}
