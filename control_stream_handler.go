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

var _ streamHandler = (*controlStreamHandler)(nil)

type controlStreamState uint8

const (
	controlStreamStateNotConnected controlStreamState = iota
	controlStreamStateConnected
)

func (s controlStreamState) String() string {
	switch s {
	case controlStreamStateNotConnected:
		return "NotConnected"
	case controlStreamStateConnected:
		return "Connected"
	default:
		return "<Unknown>"
	}
}

// controlStreamHandler Handle messages which are categorised as control messages.
//   transitions:
//     = controlStreamStateNotConnected
//       | "connect" -> controlStreamStateConnected
//       | _         -> self
//
//     = controlStreamStateConnected
//       | _ -> self
//
type controlStreamHandler struct {
	state    controlStreamState
	streamer *ChunkStreamer
	streams  *streams
	handler  Handler
	logger   logrus.FieldLogger
}

func (h *controlStreamHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	switch h.state {
	case controlStreamStateNotConnected:
		return h.handleInNotConnected(chunkStreamID, timestamp, msg, stream)

	case controlStreamStateConnected:
		return h.handleInConnected(chunkStreamID, timestamp, msg, stream)

	default:
		panic("Unreachable!")
	}
}

func (h *controlStreamHandler) handleInNotConnected(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	l := h.logger.WithFields(logrus.Fields{
		"stream_id": stream.streamID,
		"state":     h.state,
		"handler":   "control",
	})

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
		return h.handleCommonMessage(chunkStreamID, timestamp, msg, stream)
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetConnectionConnect:
		l.Info("Connect")

		if err := h.handler.OnConnect(timestamp, cmd); err != nil {
			cmdRespMsg := h.newConnectErrorMessage()

			l.Infof("Reject a connect request: Response = %#v", cmdRespMsg.Command)
			if writeErr := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncTy, cmdRespMsg); writeErr != nil {
				return errors.Wrapf(err, "Write failed: Err = %+v", writeErr)
			}

			return err
		}

		l.Infof("Set win ack size: Size = %+v", h.streamer.SelfState().AckWindowSize())
		if err := stream.WriteWinAckSize(ctrlMsgChunkStreamID, timestamp, &message.WinAckSize{
			Size: h.streamer.SelfState().AckWindowSize(),
		}); err != nil {
			return err
		}

		l.Infof("Set peer bandwidth: Size = %+v, Limit = %+v",
			h.streamer.SelfState().BandwidthWindowSize(),
			h.streamer.SelfState().BandwidthLimitType(),
		)
		if err := stream.WriteSetPeerBandwidth(ctrlMsgChunkStreamID, timestamp, &message.SetPeerBandwidth{
			Size:  h.streamer.SelfState().BandwidthWindowSize(),
			Limit: h.streamer.SelfState().BandwidthLimitType(),
		}); err != nil {
			return err
		}

		l.Infof("Stream Begin: ID = %d", 0)
		if err := stream.WriteUserCtrl(ctrlMsgChunkStreamID, timestamp, &message.UserCtrl{
			Event: &message.UserCtrlEventStreamBegin{
				StreamID: 0,
			},
		}); err != nil {
			return err
		}

		cmdRespMsg := h.newConnectSuccessMessage()
		l.Infof("Connect: Response = %#v", cmdRespMsg.Command)
		if err := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncTy, cmdRespMsg); err != nil {
			return err
		}
		l.Info("Connected")

		h.state = controlStreamStateConnected

		return nil

	default:
		return h.handler.OnUnknownCommandMessage(timestamp, cmdMsg)
	}
}

func (h *controlStreamHandler) handleInConnected(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	l := h.logger.WithFields(logrus.Fields{
		"stream_id": stream.streamID,
		"state":     h.state,
		"handler":   "control",
	})

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
		return h.handleCommonMessage(chunkStreamID, timestamp, msg, stream)
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetConnectionCreateStream:
		l.Infof("Stream creating...: %#v", cmd)

		if err := h.handler.OnCreateStream(timestamp, cmd); err != nil {
			cmdRespMsg := h.newCreateStreamErrorMessage(cmdMsg.TransactionID)
			l.Infof("Reject a CreateStream request: Response = %#v", cmdRespMsg.Command)
			if writeErr := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncTy, cmdRespMsg); writeErr != nil {
				return errors.Wrapf(err, "Write failed: Err = %+v", writeErr)
			}

			return err
		}

		// Create a stream which handles messages for data(play, publish, video, audio, etc...)
		streamID, err := h.streams.CreateIfAvailable(&dataStreamHandler{
			handler: h.handler,
			logger:  h.logger,
		})
		if err != nil {
			cmdRespMsg := h.newCreateStreamErrorMessage(cmdMsg.TransactionID)
			l.Errorf("Failed to create stream: Err = %+v, Response = %#v", err, cmdRespMsg.Command)
			if writeErr := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncTy, cmdRespMsg); writeErr != nil {
				return errors.Wrapf(err, "Write failed: Err = %+v", writeErr)
			}

			return nil
		}

		cmdRespMsg := h.newCreateStreamSuccessMessage(
			cmdMsg.TransactionID,
			streamID,
		)
		if err := stream.WriteCommandMessage(chunkStreamID, timestamp, cmdMsgEncTy, cmdRespMsg); err != nil {
			_ = h.streams.Delete(streamID) // TODO: error handling
			return err
		}

		l.Infof("Stream created...: NewStreamID = %d", streamID)

		return nil

	case *message.NetStreamDeleteStream:
		l.Infof("Stream deleting...: TargetStreamID = %d", cmd.StreamID)

		if err := h.handler.OnDeleteStream(timestamp, cmd); err != nil {
			return err
		}

		if err := h.streams.Delete(cmd.StreamID); err != nil {
			return err
		}

		// server does not send any response(7.2.2.3)

		l.Infof("Stream deleted: TargetStreamID = %d", cmd.StreamID)

		return nil

	case *message.NetConnectionReleaseStream:
		l.Infof("Release stream...: StreamName = %s", cmd.StreamName)

		if err := h.handler.OnReleaseStream(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	case *message.NetStreamFCPublish:
		l.Infof("FCPublish stream...: StreamName = %s", cmd.StreamName)

		if err := h.handler.OnFCPublish(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	case *message.NetStreamFCUnpublish:
		l.Infof("FCUnpublish stream...: StreamName = %s", cmd.StreamName)

		if err := h.handler.OnFCUnpublish(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	default:
		return h.handler.OnUnknownCommandMessage(timestamp, cmdMsg)
	}
}

func (h *controlStreamHandler) handleCommonMessage(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	l := h.logger.WithFields(logrus.Fields{
		"stream_id": stream.streamID,
		"state":     h.state,
		"handler":   "control",
	})

	switch msg := msg.(type) {
	case *message.SetChunkSize:
		l.Infof("Handle SetChunkSize: Msg = %#v", msg)
		return h.streamer.PeerState().SetChunkSize(msg.ChunkSize)

	case *message.WinAckSize:
		l.Infof("Handle WinAckSize: Msg = %#v", msg)

		return h.streamer.PeerState().SetAckWindowSize(msg.Size)

	default:
		return h.handler.OnUnknownMessage(timestamp, msg)
	}
}

func (h *controlStreamHandler) newConnectSuccessMessage() *message.CommandMessage {
	return &message.CommandMessage{
		CommandName:   "_result",
		TransactionID: 1, // 7.2.1.2, flow.6
		Command: &message.NetConnectionConnectResult{
			Properties: message.NetConnectionConnectResultProperties{
				FMSVer:       "GO-RTMP/0,0,0,0", // TODO: fix
				Capabilities: 31,                // TODO: fix
				Mode:         1,                 // TODO: fix
			},
			Information: message.NetConnectionConnectResultInformation{
				Level:       "status",
				Code:        message.NetConnectionConnectCodeSuccess,
				Description: "Connection succeeded.",
				Data: map[string]interface{}{
					"type":    "go-rtmp",
					"version": "master", // TODO: fix
				},
			},
		},
	}
}

func (h *controlStreamHandler) newConnectErrorMessage() *message.CommandMessage {
	return &message.CommandMessage{
		CommandName:   "_error",
		TransactionID: 1, // 7.2.1.2, flow.6
		Command: &message.NetConnectionConnectResult{
			Properties: message.NetConnectionConnectResultProperties{
				FMSVer:       "GO-RTMP/0,0,0,0", // TODO: fix
				Capabilities: 31,                // TODO: fix
				Mode:         1,                 // TODO: fix
			},
			Information: message.NetConnectionConnectResultInformation{
				Level:       "error",
				Code:        message.NetConnectionConnectCodeFailed,
				Description: "Connection failed.",
				Data: map[string]interface{}{
					"type":    "go-rtmp",
					"version": "master", // TODO: fix
				},
			},
		},
	}
}

func (h *controlStreamHandler) newCreateStreamSuccessMessage(
	transactionID int64,
	streamID uint32,
) *message.CommandMessage {
	return &message.CommandMessage{
		CommandName:   "_result",
		TransactionID: transactionID,
		Command: &message.NetConnectionCreateStreamResult{
			StreamID: streamID,
		},
	}
}

func (h *controlStreamHandler) newCreateStreamErrorMessage(
	transactionID int64,
) *message.CommandMessage {
	return &message.CommandMessage{
		CommandName:   "_error",
		TransactionID: transactionID,
		Command: &message.NetConnectionCreateStreamResult{
			StreamID: 0, // TODO: Change to error information object...
		},
	}
}
