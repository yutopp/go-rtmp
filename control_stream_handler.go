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
	conn           *Conn
	state          controlStreamState
	defaultHandler streamHandler

	logger *logrus.Entry
}

func (h *controlStreamHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	switch h.state {
	case controlStreamStateNotConnected:
		return h.handleConnect(chunkStreamID, timestamp, msg, stream)
	case controlStreamStateConnected:
		return h.handleCreateStream(chunkStreamID, timestamp, msg, stream)
	default:
		panic("Unreachable!")
	}
}

func (h *controlStreamHandler) handleConnect(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
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
		h.logger.Printf("Message unhandled(netConnection): Message = %+v, State = %d", msg, h.state)
		return h.defaultHandler.Handle(chunkStreamID, timestamp, msg, stream)
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetConnectionConnect:
		h.logger.Printf("connect: %+v", msg)

		if err := h.conn.handler.OnConnect(timestamp, cmd); err != nil {
			return err
		}

		// TODO: fix
		if err := stream.Write(chunkStreamID, timestamp, &message.WinAckSize{
			Size: h.conn.streamer.selfState.windowSize,
		}); err != nil {
			return err
		}

		// TODO: fix
		if err := stream.Write(chunkStreamID, timestamp, &message.SetPeerBandwidth{
			Size:  1 * 1024 * 1024,
			Limit: 1,
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
		h.logger.Printf("conn: %+v", m.(*message.CommandMessageAMF0).Command)
		if err := stream.Write(chunkStreamID, timestamp, m); err != nil {
			return err
		}
		h.logger.Printf("connected")

		h.state = controlStreamStateConnected

		return nil

	default:
		h.logger.Printf("Unexpected command(netConnection): Command = %+v, State = %d", cmdMsg, h.state)
		return nil
	}

}

func (h *controlStreamHandler) handleCreateStream(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
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
		h.logger.Printf("Message unhandled(netConnection): Message = %+v, State = %d", msg, h.state)
		return h.defaultHandler.Handle(chunkStreamID, timestamp, msg, stream)
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetConnectionCreateStream:
		h.logger.Printf("Stream creating...: %+v", cmd)

		// Create a stream which handles messages for data(play, publish, video, audio, etc...)
		handler := &dataStreamHandler{
			conn:           h.conn,
			defaultHandler: h.defaultHandler,
		}
		streamID, err := h.conn.createStreamIfAvailable(handler)
		if err != nil {
			// TODO: send failed _result
			h.logger.Printf("Failed to create stream: Err = %+v", err)
			return nil
		}

		handler.logger = h.logger.WithField("StreamID", streamID)

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

		h.logger.Printf("Stream created: StreamID: %d", streamID)

		return nil

	case *message.NetStreamDeleteStream:
		h.logger.Infof("Stream deleteing...: StreamID = %d", cmd.StreamID)
		if err := h.conn.deleteStream(cmd.StreamID); err != nil {
			return err
		}

		// server does not send any response(7.2.2.3)

		h.logger.Printf("Stream deleted: StreamID: %d", cmd.StreamID)

		return nil

	default:
		h.logger.Printf("Unexpected command(netConnection): Command = %+v, State = %d", cmdMsg, h.state)
		return nil
	}
}
