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

var _ messageHandler = (*serverDataInactiveHandler)(nil)

// serverDataInactiveHandler Handle data messages from a non operated client at server side.
//   transitions:
//     | "publish" -> serverDataPublishHandler
//     | "play"	   -> serverDataPlayHandler
//     | _         -> self
type serverDataInactiveHandler struct {
	entry *entryHandler
}

func (h *serverDataInactiveHandler) Handle(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverDataInactiveHandler) HandleCommand(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	cmdMsg *message.CommandMessage,
	body interface{},
	stream *Stream,
) error {
	l := h.entry.Logger()

	switch cmd := body.(type) {
	case *message.NetStreamPublish:
		l.Infof("Publisher is comming: %#v", cmd)

		if err := h.entry.conn.handler.OnPublish(timestamp, cmd); err != nil {
			// TODO: Support message.NetStreamOnStatusCodePublishBadName
			cmdRespMsg := h.newOnStatusMessage(
				message.NetStreamOnStatusCodePublishFailed,
				"Publish failed.",
			)
			l.Infof("Reject a Publish request: Response = %#v", cmdRespMsg.Encoder.Value)
			if writeErr := stream.WriteCommandMessage(chunkStreamID, timestamp, encTy, cmdRespMsg); writeErr != nil {
				return errors.Wrapf(err, "Write failed: Err = %+v", writeErr)
			}

			return err
		}

		cmdRespMsg := h.newOnStatusMessage(
			message.NetStreamOnStatusCodePublishStart,
			"Publish succeeded.",
		)
		if err := stream.WriteCommandMessage(chunkStreamID, timestamp, encTy, cmdRespMsg); err != nil {
			return err
		}
		l.Infof("Publisher accepted")

		h.entry.ChangeState(&serverDataPublishHandler{entry: h.entry})

		return nil

	case *message.NetStreamPlay:
		l.Infof("Player is comming: %#v", cmd)

		if err := h.entry.conn.handler.OnPlay(timestamp, cmd); err != nil {
			cmdRespMsg := h.newOnStatusMessage(
				message.NetStreamOnStatusCodePlayFailed,
				"Play failed.",
			)
			l.Infof("Reject a Play request: Response = %#v", cmdRespMsg.Encoder.Value)
			if writeErr := stream.WriteCommandMessage(chunkStreamID, timestamp, encTy, cmdRespMsg); writeErr != nil {
				return errors.Wrapf(err, "Write failed: Err = %+v", writeErr)
			}

			return err
		}

		cmdRespMsg := h.newOnStatusMessage(
			message.NetStreamOnStatusCodePlayStart,
			"Play succeeded.",
		)
		if err := stream.WriteCommandMessage(chunkStreamID, timestamp, encTy, cmdRespMsg); err != nil {
			return err
		}
		l.Infof("Player accepted")

		h.entry.ChangeState(&serverDataPlayHandler{entry: h.entry})

		return nil

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *serverDataInactiveHandler) newOnStatusMessage(
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

	bodyEnc := &message.BodyEncoder{
		Value: &message.NetStreamOnStatus{
			InfoObject: message.NetStreamOnStatusInfoObject{
				Level:       level,
				Code:        code,
				Description: description,
			},
		},
		MsgEncoder: message.EncodeBodyAnyValues,
	}
	return &message.CommandMessage{
		CommandName:   "onStatus",
		TransactionID: 0, // 7.2.2
		Encoder:       bodyEnc,
	}
}

func (h *serverDataInactiveHandler) HandleData(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	dataMsg *message.DataMessage,
	body interface{},
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}
