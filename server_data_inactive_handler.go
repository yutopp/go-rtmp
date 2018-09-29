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

func (h *serverDataInactiveHandler) HandleData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	body interface{},
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverDataInactiveHandler) HandleCommand(
	chunkStreamID int,
	timestamp uint32,
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
			result := h.newOnStatus(message.NetStreamOnStatusCodePublishFailed, "Publish failed.")

			l.Infof("Reject a Publish request: Response = %#v", result)
			if err1 := stream.NotifyStatus(chunkStreamID, timestamp, result); err1 != nil {
				return errors.Wrapf(err, "Failed to reply response: Err = %+v", err1)
			}

			return err
		}

		result := h.newOnStatus(message.NetStreamOnStatusCodePublishStart, "Publish succeeded.")
		if err := stream.NotifyStatus(chunkStreamID, timestamp, result); err != nil {
			return err
		}
		l.Infof("Publisher accepted")

		h.entry.ChangeState(&serverDataPublishHandler{entry: h.entry})

		return nil

	case *message.NetStreamPlay:
		l.Infof("Player is comming: %#v", cmd)

		if err := h.entry.conn.handler.OnPlay(timestamp, cmd); err != nil {
			result := h.newOnStatus(message.NetStreamOnStatusCodePlayFailed, "Play failed.")

			l.Infof("Reject a Play request: Response = %#v", result)
			if err1 := stream.NotifyStatus(chunkStreamID, timestamp, result); err1 != nil {
				return errors.Wrapf(err, "Failed to reply response: Err = %+v", err1)
			}

			return err
		}

		result := h.newOnStatus(message.NetStreamOnStatusCodePlayStart, "Play succeeded.")
		if err := stream.NotifyStatus(chunkStreamID, timestamp, result); err != nil {
			return err
		}
		l.Infof("Player accepted")

		h.entry.ChangeState(&serverDataPlayHandler{entry: h.entry})

		return nil

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *serverDataInactiveHandler) newOnStatus(
	code message.NetStreamOnStatusCode,
	description string,
) *message.NetStreamOnStatus {
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

	return &message.NetStreamOnStatus{
		InfoObject: message.NetStreamOnStatusInfoObject{
			Level:       level,
			Code:        code,
			Description: description,
		},
	}
}
