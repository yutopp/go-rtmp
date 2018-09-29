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

func (h *serverControlConnectedHandler) HandleData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	body interface{},
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlConnectedHandler) HandleCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
	body interface{},
	stream *Stream,
) (err error) {
	l := h.entry.Logger()

	switch cmd := body.(type) {
	case *message.NetConnectionCreateStream:
		l.Infof("Stream creating...: %#v", cmd)
		defer func() {
			if err != nil {
				result := h.newCreateStreamErrorResult()

				l.Infof("CreateStream(Error): ResponseBody = %#v", result)
				if err1 := stream.ReplyCreateStream(chunkStreamID, timestamp, cmdMsg.TransactionID, result); err1 != nil {
					err = errors.Wrapf(err, "Failed to reply response: Err = %+v", err1)
				}
			}
		}()

		if err := h.entry.conn.handler.OnCreateStream(timestamp, cmd); err != nil {
			return err
		}

		// Create a stream which handles messages for data(play, publish, video, audio, etc...)
		eh := h.entry.Clone()
		eh.ChangeState(&serverDataInactiveHandler{entry: eh})
		streamID, err := h.entry.conn.streams.CreateIfAvailable(eh)
		if err != nil {
			l.Errorf("Failed to create stream: Err = %+v", err)
			result := h.newCreateStreamErrorResult()
			if err1 := stream.ReplyCreateStream(chunkStreamID, timestamp, cmdMsg.TransactionID, result); err1 != nil {
				return errors.Wrapf(err, "Failed to reply response: Err = %+v", err1)
			}

			return nil // Keep the connection
		}

		result := h.newCreateStreamSuccessResult(streamID)
		if err := stream.ReplyCreateStream(chunkStreamID, timestamp, cmdMsg.TransactionID, result); err != nil {
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

func (h *serverControlConnectedHandler) newCreateStreamSuccessResult(
	streamID uint32,
) *message.NetConnectionCreateStreamResult {
	return &message.NetConnectionCreateStreamResult{
		StreamID: streamID,
	}
}

func (h *serverControlConnectedHandler) newCreateStreamErrorResult() *message.NetConnectionCreateStreamResult {
	return nil
}
