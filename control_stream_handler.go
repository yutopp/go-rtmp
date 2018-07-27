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
	conn  *Conn
	state controlStreamState

	logger logrus.FieldLogger
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
		return h.handleCommonMessage(chunkStreamID, timestamp, msg, stream)
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetConnectionConnect:
		l.Info("Connect")

		if err := h.conn.handler.OnConnect(timestamp, cmd); err != nil {
			return err
		}

		// TODO: fix
		l.Infof("Set win ack size: Size = %+v", h.conn.streamer.SelfState().AckWindowSize())
		if err := stream.Write(chunkStreamID, timestamp, &message.WinAckSize{
			Size: h.conn.streamer.SelfState().AckWindowSize(),
		}); err != nil {
			return err
		}

		// TODO: fix
		l.Infof("Set peer bandwidth: Size = %+v, Limit = %+v",
			h.conn.streamer.SelfState().BandwidthWindowSize(),
			h.conn.streamer.SelfState().BandwidthLimitType(),
		)
		if err := stream.Write(chunkStreamID, timestamp, &message.SetPeerBandwidth{
			Size:  h.conn.streamer.SelfState().BandwidthWindowSize(),
			Limit: h.conn.streamer.SelfState().BandwidthLimitType(),
		}); err != nil {
			return err
		}

		// TODO: fix
		m := cmdMsgWrapper(func(cmsg *message.CommandMessage) {
			*cmsg = message.CommandMessage{
				CommandName:   "_result",
				TransactionID: 1, // 7.2.1.2, flow.6
				Command: &message.NetConnectionConnectResult{
					Properties: message.NetConnectionConnectResultProperties{
						FMSVer:       "rtmp/testing",
						Capabilities: 250,
						Mode:         1,
					},
					Information: message.NetConnectionConnectResultInformation{
						Level: "status",
						Code:  "NetConnection.Connect.Success",
						Data: map[string]interface{}{
							"version": "testing",
						},
						Application: nil,
					},
				},
			}
		})
		l.Infof("Connect: Response = %#v", m.(*message.CommandMessageAMF0).Command)

		if err := stream.Write(chunkStreamID, timestamp, m); err != nil {
			return err
		}
		l.Info("Connected")

		h.state = controlStreamStateConnected

		return nil

	default:
		l.Warnf("Unexpected command: Command = %#v", cmdMsg)

		return nil
	}
}

func (h *controlStreamHandler) handleInConnected(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	l := h.logger.WithFields(logrus.Fields{
		"stream_id": stream.streamID,
		"state":     h.state,
		"handler":   "control",
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
		return h.handleCommonMessage(chunkStreamID, timestamp, msg, stream)
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetConnectionCreateStream:
		l.Infof("Stream creating...: %#v", cmd)

		if err := h.conn.handler.OnCreateStream(timestamp, cmd); err != nil {
			return err
		}

		// Create a stream which handles messages for data(play, publish, video, audio, etc...)
		streamID, err := h.conn.createStreamIfAvailable(&dataStreamHandler{
			conn:   h.conn,
			logger: h.logger,
		})
		if err != nil {
			// TODO: send failed _result
			l.Errorf("Stream creating...: Err = %#v", err)

			return nil
		}

		// TODO: fix
		m := cmdMsgWrapper(func(cmsg *message.CommandMessage) {
			*cmsg = message.CommandMessage{
				CommandName:   "_result",
				TransactionID: cmdMsg.TransactionID,
				Command: &message.NetConnectionCreateStreamResult{
					StreamID: streamID,
				},
			}
		})
		if err := stream.Write(chunkStreamID, timestamp, m); err != nil {
			_ = h.conn.deleteStream(streamID) // TODO: error handling
			return err
		}

		l.Infof("Stream created...: NewStreamID = %d", streamID)

		return nil

	case *message.NetStreamDeleteStream:
		l.Infof("Stream deleting...: TargetStreamID = %d", cmd.StreamID)

		if err := h.conn.handler.OnDeleteStream(timestamp, cmd); err != nil {
			return err
		}

		if err := h.conn.deleteStream(cmd.StreamID); err != nil {
			return err
		}

		// server does not send any response(7.2.2.3)

		l.Infof("Stream deleted: TargetStreamID = %d", cmd.StreamID)

		return nil

	case *message.NetConnectionReleaseStream:
		l.Infof("Release stream...: StreamName = %s", cmd.StreamName)

		if err := h.conn.handler.OnReleaseStream(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	case *message.NetStreamFCPublish:
		l.Infof("FCPublish stream...: StreamName = %s", cmd.StreamName)

		if err := h.conn.handler.OnFCPublish(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	case *message.NetStreamFCUnpublish:
		l.Infof("FCUnpublish stream...: StreamName = %s", cmd.StreamName)

		if err := h.conn.handler.OnFCUnpublish(timestamp, cmd); err != nil {
			return err
		}

		// TODO: send _result?

		return nil

	default:
		l.Warnf("Unexpected command: Command = %#v", cmdMsg)

		return nil
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
		return h.conn.streamer.PeerState().SetChunkSize(msg.ChunkSize)

	case *message.WinAckSize:
		l.Infof("Handle WinAckSize: Msg = %#v", msg)

		return h.conn.streamer.PeerState().SetAckWindowSize(msg.Size)

	default:
		l.Warnf("Message unhandled: Msg = %#v", msg)

		return nil
	}
}
