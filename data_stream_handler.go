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

func (s dataStreamState) String() string {
	switch s {
	case dataStreamStateNotInAction:
		return "NotInAction"
	case dataStreamStateHasPublisher:
		return "HasPublisher"
	case dataStreamStateHasPlayer:
		return "HasPlayer"
	default:
		return "<Unknown>"
	}
}

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
	state   dataStreamState
	handler Handler

	logger      logrus.FieldLogger
	loggerEntry *logrus.Entry
}

func (h *dataStreamHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	switch h.state {
	case dataStreamStateNotInAction:
		return h.handleAction(chunkStreamID, timestamp, msg, stream)

	case dataStreamStateHasPublisher:
		return h.handlePublisher(chunkStreamID, timestamp, msg, stream)

	case dataStreamStateHasPlayer:
		return h.handlePlayer(chunkStreamID, timestamp, msg, stream)

	default:
		panic("Unreachable!")
	}
}

func (h *dataStreamHandler) handleAction(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	l := h.loggerInstance(stream)

	var cmdMsgEncodingType message.EncodingType
	var cmdMsg *message.CommandMessage
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		cmdMsgEncodingType = message.EncodingTypeAMF0
		cmdMsg = &msg.CommandMessage
		goto handleCommand

	case *message.CommandMessageAMF3:
		cmdMsgEncodingType = message.EncodingTypeAMF3
		cmdMsg = &msg.CommandMessage
		goto handleCommand

	default:
		l.Warnf("Message unhandled: Msg = %#v", msg)

		return nil
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetStreamPublish:
		l.Infof("Publisher is comming: %#v", cmd)

		if err := h.handler.OnPublish(timestamp, cmd); err != nil {
			return err
		}

		// TODO: fix
		cmdRespMsg := &message.CommandMessage{
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
		if err := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncodingType, cmdRespMsg); err != nil {
			return err
		}
		l.Infof("Publisher accepted")

		h.state = dataStreamStateHasPublisher

		return nil

	case *message.NetStreamPlay:
		l.Infof("Player is comming: %#v", cmd)

		if err := h.handler.OnPlay(timestamp, cmd); err != nil {
			return err
		}

		// TODO: fix
		cmdRespMsg := &message.CommandMessage{
			CommandName:   "onStatus",
			TransactionID: 0,
			Command: &message.NetStreamOnStatus{
				InfoObject: message.NetStreamOnStatusInfoObject{
					Level:       "status",
					Code:        "NetStream.Play.Start",
					Description: "yoyo",
				},
			},
		}
		if err := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncodingType, cmdRespMsg); err != nil {
			return err
		}
		l.Infof("Player accepted")

		h.state = dataStreamStateHasPlayer

		return nil

	default:
		l.Warnf("Unexpected command: Command = %#v", cmdMsg)

		return nil
	}
}

func (h *dataStreamHandler) handlePublisher(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	l := h.loggerInstance(stream)

	var dataMsg *message.DataMessage
	switch msg := msg.(type) {
	case *message.AudioMessage:
		return h.handler.OnAudio(timestamp, msg.Payload)

	case *message.VideoMessage:
		return h.handler.OnVideo(timestamp, msg.Payload)

	case *message.DataMessageAMF0:
		dataMsg = &msg.DataMessage
		goto handleCommand

	case *message.DataMessageAMF3:
		dataMsg = &msg.DataMessage
		goto handleCommand

	default:
		l.Warnf("Message unhandled: Msg = %#v", msg)

		return nil
	}

handleCommand:
	switch dataMsg.Name {
	case "@setDataFrame":
		df := dataMsg.Data.(*message.NetStreamSetDataFrame)
		if df == nil {
			return errors.New("setDataFrame has nil value")
		}
		return h.handler.OnSetDataFrame(timestamp, df)

	default:
		l.Warnf("Ignore unknown data message: Msg = %#v", dataMsg)

		return nil
	}
}

func (h *dataStreamHandler) handlePlayer(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	l := h.loggerInstance(stream)

	switch msg := msg.(type) {
	default:
		l.Warnf("Message unhandled: Msg = %#v", msg)

		return nil
	}
}

func (h *dataStreamHandler) loggerInstance(stream *Stream) *logrus.Entry {
	if h.loggerEntry == nil {
		h.loggerEntry = h.logger.WithField("handler", "data")
	}

	h.loggerEntry.Data["state"] = h.state
	h.loggerEntry.Data["stream_id"] = stream.streamID

	return h.loggerEntry
}
