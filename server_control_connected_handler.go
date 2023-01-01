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

var _ stateHandler = (*serverControlConnectedHandler)(nil)

// serverControlConnectedHandler Handle control messages from a client at server side.
//
//	transitions:
//	  | "createStream" -> spawn! serverDataInactiveHandler
//	  | _              -> self
type serverControlConnectedHandler struct {
	sh *streamHandler
}

func (h *serverControlConnectedHandler) onMessage(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlConnectedHandler) onData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	body interface{},
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlConnectedHandler) onCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
	body interface{},
) (err error) {
	l := h.sh.Logger()
	tID := cmdMsg.TransactionID

	switch cmd := body.(type) {
	case *message.NetConnectionCreateStream:
		l.Infof("Stream creating...: %#v", cmd)
		defer func() {
			if err != nil {
				result := h.newCreateStreamErrorResult()

				l.Infof("CreateStream(Error): ResponseBody = %#v, Err = %+v", result, err)
				if err1 := h.sh.stream.ReplyCreateStream(chunkStreamID, timestamp, tID, result); err1 != nil {
					err = errors.Wrapf(err, "Failed to reply response: Err = %+v", err1)
				}
			}
		}()

		if err := h.sh.stream.userHandler().OnCreateStream(timestamp, cmd); err != nil {
			return err
		}

		// Create a stream which handles messages for data(play, publish, video, audio, etc...)
		newStream, err := h.sh.stream.streams().conn.streams.CreateIfAvailable()
		if err != nil {
			l.Errorf("Failed to create stream: Err = %+v", err)

			result := h.newCreateStreamErrorResult()
			if err1 := h.sh.stream.ReplyCreateStream(chunkStreamID, timestamp, tID, result); err1 != nil {
				return errors.Wrapf(err, "Failed to reply response: Err = %+v", err1)
			}

			return nil // Keep the connection
		}
		newStream.handler.ChangeState(streamStateServerInactive)

		result := h.newCreateStreamSuccessResult(newStream.streamID)
		if err := h.sh.stream.ReplyCreateStream(chunkStreamID, timestamp, tID, result); err != nil {
			_ = h.sh.stream.streams().Delete(newStream.streamID) // TODO: error handling
			return err
		}

		l.Infof("Stream created...: NewStreamID = %d", newStream.streamID)

		return nil

	case *message.NetStreamDeleteStream:
		l.Infof("Stream deleting...: TargetStreamID = %d", cmd.StreamID)

		if err := h.sh.stream.userHandler().OnDeleteStream(timestamp, cmd); err != nil {
			return err
		}

		if err := h.sh.stream.streams().Delete(cmd.StreamID); err != nil {
			return err
		}

		// server does not send any response(7.2.2.3)

		l.Infof("Stream deleted: TargetStreamID = %d", cmd.StreamID)

		return nil

	case *message.NetConnectionReleaseStream:
		l.Infof("Release stream...: StreamName = %s", cmd.StreamName)

		if err := h.sh.stream.userHandler().OnReleaseStream(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	case *message.NetStreamFCPublish:
		l.Infof("FCPublish stream...: StreamName = %s", cmd.StreamName)

		if err := h.sh.stream.userHandler().OnFCPublish(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	case *message.NetStreamFCUnpublish:
		l.Infof("FCUnpublish stream...: StreamName = %s", cmd.StreamName)

		if err := h.sh.stream.userHandler().OnFCUnpublish(timestamp, cmd); err != nil {
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
