//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"log"

	"github.com/yutopp/go-rtmp/message"
)

type streamHandler interface {
	Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error
}

var _ streamHandler = (*netConnectionMessageHandler)(nil)

type netConnectionState uint8

const (
	netConnectionStateNotConnected netConnectionState = iota
	netConnectionStateConnected
)

// netConnectionMessageHandler Handle messages which are categorised as NetConnection.
//   transitions:
//     = netConnectionStateNotConnected
//       | "connect" -> netConnectionStateConnected
//       | _         -> self
//
//     = netConnectionStateConnected
//       | _ -> self
//
type netConnectionMessageHandler struct {
	conn  *Conn
	state netConnectionState
}

func (h *netConnectionMessageHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	switch h.state {
	case netConnectionStateNotConnected:
		return h.handleConnect(chunkStreamID, timestamp, msg, stream)
	case netConnectionStateConnected:
		return h.handleCreateStream(chunkStreamID, timestamp, msg, stream)
	default:
		panic("Unreachable!")
	}
}

func (h *netConnectionMessageHandler) handleConnect(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
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
		log.Printf("Unexpected message(netConnection): Message = %+v, State = %d", msg, h.state)
		return nil
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetConnectionConnect:
		log.Printf("connect: %+v", msg)

		if err := h.conn.handler.OnConnect(timestamp, cmd); err != nil {
			return err
		}

		// TODO: fix
		if err := stream.Write(chunkStreamID, timestamp, &message.WinAckSize{
			Size: h.conn.streamer.windowSize,
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
		log.Printf("conn: %+v", m.(*message.CommandMessageAMF0).Command)
		if err := stream.Write(chunkStreamID, timestamp, m); err != nil {
			return err
		}
		log.Printf("connected")

		h.state = netConnectionStateConnected

		return nil

	default:
		log.Printf("Unexpected command(netConnection): Command = %+v, State = %d", cmdMsg, h.state)
		return nil
	}

}

func (h *netConnectionMessageHandler) handleCreateStream(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
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
		log.Printf("Unexpected message(netConnection): Message = %+v, State = %d", msg, h.state)
		return nil
	}

handleCommand:
	switch cmd := cmdMsg.Command.(type) {
	case *message.NetConnectionCreateStream:
		log.Printf("Stream creating...: %+v", cmd)

		// Create a stream which handles NetStream(publish, play, etc...)
		streamID, err := h.conn.createStreamIfAvailable(&netStreamMessageHandler{
			conn: h.conn,
		})
		if err != nil {
			// TODO: send failed _result
			log.Printf("Failed to create stream: Err = %+v", err)
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

		log.Printf("Stream created: StreamID: %d", streamID)

		return nil

	default:
		log.Printf("Unexpected command(netConnection): Command = %+v, State = %d", cmdMsg, h.state)
		return nil
	}
}

type amfWrapperFunc func(callback func(cmd *message.CommandMessage)) message.Message

func amf0Wrapper(callback func(cmd *message.CommandMessage)) message.Message {
	var m message.CommandMessageAMF0
	callback(&m.CommandMessage)
	return &m
}

func amf3Wrapper(callback func(cmd *message.CommandMessage)) message.Message {
	var m message.CommandMessageAMF3
	callback(&m.CommandMessage)
	return &m
}
