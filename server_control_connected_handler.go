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

var _ messageHandler = (*serverControlConnectedHandler)(nil)

// serverControlConnectedHandler Handle control messages from a client at server side.
//   transitions:
//     | "createStream" -> spawn! serverDataInactiveHandler
//     | _              -> self
//
type serverControlConnectedHandler struct {
	entry *entryHandler
}

func (h *serverControlConnectedHandler) Handle(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlConnectedHandler) HandleCommand(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	cmdMsg *message.CommandMessage,
	body interface{},
	stream *Stream,
) error {
	l := h.entry.Logger()

	switch cmd := body.(type) {
	case *message.NetConnectionCreateStream:
		l.Infof("Stream creating...: %#v", cmd)

		if err := h.entry.conn.handler.OnCreateStream(timestamp, cmd); err != nil {
			cmdRespMsg := h.newCreateStreamErrorMessage(cmdMsg.TransactionID)
			l.Infof("Reject a CreateStream request: Response = %#v", cmdRespMsg.Encoder.Value)
			if writeErr := stream.WriteCommandMessage(chunkStreamID, timestamp, encTy, cmdRespMsg); writeErr != nil {
				return errors.Wrapf(err, "Write failed: Err = %+v", writeErr)
			}

			return err
		}

		// Create a stream which handles messages for data(play, publish, video, audio, etc...)
		eh := h.entry.Clone()
		eh.ChangeState(&serverDataInactiveHandler{entry: eh})
		streamID, err := h.entry.conn.streams.CreateIfAvailable(eh)
		if err != nil {
			cmdRespMsg := h.newCreateStreamErrorMessage(cmdMsg.TransactionID)
			l.Errorf("Failed to create stream: Err = %+v, Response = %#v", err, cmdRespMsg.Encoder.Value)
			if writeErr := stream.WriteCommandMessage(chunkStreamID, timestamp, encTy, cmdRespMsg); writeErr != nil {
				return errors.Wrapf(err, "Write failed: Err = %+v", writeErr)
			}

			return nil
		}

		cmdRespMsg := h.newCreateStreamSuccessMessage(
			cmdMsg.TransactionID,
			streamID,
		)
		if err := stream.WriteCommandMessage(chunkStreamID, timestamp, encTy, cmdRespMsg); err != nil {
			_ = h.entry.conn.streams.Delete(streamID) // TODO: error handling
			return err
		}

		l.Infof("Stream created...: NewStreamID = %d", streamID)

		return nil

	case *message.NetStreamDeleteStream:
		l.Infof("Stream deleting...: TargetStreamID = %d", cmd.StreamID)

		if err := h.entry.conn.handler.OnDeleteStream(timestamp, cmd); err != nil {
			return err
		}

		if err := h.entry.conn.streams.Delete(cmd.StreamID); err != nil {
			return err
		}

		// server does not send any response(7.2.2.3)

		l.Infof("Stream deleted: TargetStreamID = %d", cmd.StreamID)

		return nil

	case *message.NetConnectionReleaseStream:
		l.Infof("Release stream...: StreamName = %s", cmd.StreamName)

		if err := h.entry.conn.handler.OnReleaseStream(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	case *message.NetStreamFCPublish:
		l.Infof("FCPublish stream...: StreamName = %s", cmd.StreamName)

		if err := h.entry.conn.handler.OnFCPublish(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	case *message.NetStreamFCUnpublish:
		l.Infof("FCUnpublish stream...: StreamName = %s", cmd.StreamName)

		if err := h.entry.conn.handler.OnFCUnpublish(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *serverControlConnectedHandler) HandleData(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	dataMsg *message.DataMessage,
	body interface{},
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlConnectedHandler) newCreateStreamSuccessMessage(
	transactionID int64,
	streamID uint32,
) *message.CommandMessage {
	bodyEnc := &message.BodyEncoder{
		Value: &message.NetConnectionCreateStreamResult{
			StreamID: streamID,
		},
		MsgEncoder: message.EncodeBodyAnyValues,
	}
	return &message.CommandMessage{
		CommandName:   "_result",
		TransactionID: transactionID,
		Encoder:       bodyEnc,
	}
}

func (h *serverControlConnectedHandler) newCreateStreamErrorMessage(
	transactionID int64,
) *message.CommandMessage {
	bodyEnc := &message.BodyEncoder{
		Value: &message.NetConnectionCreateStreamResult{
			StreamID: 0, // TODO: Change to error information object...
		},
		MsgEncoder: message.EncodeBodyAnyValues,
	}
	return &message.CommandMessage{
		CommandName:   "_error",
		TransactionID: transactionID,
		Encoder:       bodyEnc,
	}
}
