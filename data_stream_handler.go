//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/pkg/errors"
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

	logger logrus.FieldLogger
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
	l := h.logger.WithFields(logrus.Fields{
		"stream_id": stream.streamID,
		"state":     h.state,
		"handler":   "data",
	})

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
		l.Infof("Message unhandled: Msg = %+v", msg)
		return h.defaultHandler.Handle(chunkStreamID, timestamp, msg, stream)
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetStreamPublish:
		l.Infof("Publisher is comming: %+v", cmd)

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
		l.Infof("Publisher accepted")

		h.state = dataStreamStateHasPublisher

		return nil

	default:
		l.Infof("Unexpected command: Command = %+v", cmdMsg)
		return nil
	}
}

func (h *dataStreamHandler) handlePublisher(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	l := h.logger.WithFields(logrus.Fields{
		"stream_id": stream.streamID,
		"state":     h.state,
		"handler":   "data",
	})

	var dataMsg *message.DataMessage
	switch msg := msg.(type) {
	case *message.AudioMessage:
		return h.conn.handler.OnAudio(timestamp, msg.Payload)

	case *message.VideoMessage:
		return h.conn.handler.OnVideo(timestamp, msg.Payload)

	case *message.DataMessageAMF0:
		dataMsg = &msg.DataMessage
		goto handleCommand

	case *message.DataMessageAMF3:
		dataMsg = &msg.DataMessage
		goto handleCommand

	default:
		l.Infof("Message unhandled: Msg = %+v", msg)
		return h.defaultHandler.Handle(chunkStreamID, timestamp, msg, stream)
	}

handleCommand:
	switch dataMsg.Name {
	case "@setDataFrame":
		df := dataMsg.Data.(*message.NetStreamSetDataFrame)
		if df == nil {
			return errors.New("setDataFrame has nil value")
		}
		return h.conn.handler.OnSetDataFrame(timestamp, df.Payload)

	default:
		l.Infof("Ignore unknown data message: Msg = %+v", dataMsg)
		return nil
	}
}
