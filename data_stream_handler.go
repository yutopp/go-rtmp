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

	var cmdMsgEncTy message.EncodingType
	var cmdMsg *message.CommandMessage
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		cmdMsgEncTy = message.EncodingTypeAMF0
		cmdMsg = &msg.CommandMessage
		goto handleCommand

	case *message.CommandMessageAMF3:
		cmdMsgEncTy = message.EncodingTypeAMF3
		cmdMsg = &msg.CommandMessage
		goto handleCommand

	default:
		return h.handler.OnUnknownMessage(timestamp, msg)
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetStreamPublish:
		l.Infof("Publisher is comming: %#v", cmd)

		if err := h.handler.OnPublish(timestamp, cmd); err != nil {
			// TODO: Support message.NetStreamOnStatusCodePublishBadName
			cmdRespMsg := h.newOnStatusMessage(
				message.NetStreamOnStatusCodePublishFailed,
				"Publish failed.",
			)
			l.Infof("Reject a Publish request: Response = %#v", cmdRespMsg.Command)
			if writeErr := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncTy, cmdRespMsg); writeErr != nil {
				return errors.Wrapf(err, "Write failed: Err = %+v", writeErr)
			}

			return err
		}

		cmdRespMsg := h.newOnStatusMessage(
			message.NetStreamOnStatusCodePublishStart,
			"Publish succeeded.",
		)
		if err := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncTy, cmdRespMsg); err != nil {
			return err
		}
		l.Infof("Publisher accepted")

		h.state = dataStreamStateHasPublisher

		return nil

	case *message.NetStreamPlay:
		l.Infof("Player is comming: %#v", cmd)

		if err := h.handler.OnPlay(timestamp, cmd); err != nil {
			cmdRespMsg := h.newOnStatusMessage(
				message.NetStreamOnStatusCodePlayFailed,
				"Play failed.",
			)
			l.Infof("Reject a Play request: Response = %#v", cmdRespMsg.Command)
			if writeErr := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncTy, cmdRespMsg); writeErr != nil {
				return errors.Wrapf(err, "Write failed: Err = %+v", writeErr)
			}

			return err
		}

		cmdRespMsg := h.newOnStatusMessage(
			message.NetStreamOnStatusCodePlayStart,
			"Play succeeded.",
		)
		if err := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncTy, cmdRespMsg); err != nil {
			return err
		}
		l.Infof("Player accepted")

		h.state = dataStreamStateHasPlayer

		return nil

	default:
		return h.handler.OnUnknownCommandMessage(timestamp, cmdMsg)
	}
}

func (h *dataStreamHandler) handlePublisher(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
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
		return h.handler.OnUnknownMessage(timestamp, msg)
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
		return h.handler.OnUnknownDataMessage(timestamp, dataMsg)
	}
}

func (h *dataStreamHandler) handlePlayer(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	switch msg := msg.(type) {
	default:
		return h.handler.OnUnknownMessage(timestamp, msg)
	}
}

func (h *dataStreamHandler) newOnStatusMessage(
	code message.NetStreamOnStatusCode,
	description string,
) *message.CommandMessage {
	// https://helpx.adobe.com/adobe-media-server/ssaslr/netstream-class.html#netstream_onstatus
	level := message.NetStreamOnStatusLevelStatus
	switch code {
	case message.NetStreamOnStatusCodeConnectFailed:
		fallthrough
	case message.NetStreamOnStatusCodePlayFailed:
		fallthrough
	case message.NetStreamOnStatusCodePublishBadName, message.NetStreamOnStatusCodePublishFailed:
		level = message.NetStreamOnStatusLevelError
	}

	return &message.CommandMessage{
		CommandName:   "onStatus",
		TransactionID: 0, // 7.2.2
		Command: &message.NetStreamOnStatus{
			InfoObject: message.NetStreamOnStatusInfoObject{
				Level:       level,
				Code:        code,
				Description: description,
			},
		},
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
