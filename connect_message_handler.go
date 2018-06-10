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

type stateHandler interface {
	Handle(chunkStreamID int, timestamp uint32, msg message.Message) error
}

var _ stateHandler = (*connectMessageHandler)(nil)

// connectMessageHandler Handle message until "connect" message arived.
//   transitions:
//     "connect" -> creatingStreamMessageHandler
//     _         -> self
type connectMessageHandler struct {
	conn *Conn
}

func (h *connectMessageHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message) error {
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		return h.handleCommand(chunkStreamID, timestamp, &msg.CommandMessage, amf0Wrapper)
	case *message.CommandMessageAMF3:
		return h.handleCommand(chunkStreamID, timestamp, &msg.CommandMessage, amf3Wrapper)
	default:
		log.Printf("Unexpected message(connect): %+v", msg)
		return nil
	}
}

func (h *connectMessageHandler) handleCommand(chunkStreamID int, timestamp uint32, msg *message.CommandMessage, wrapper amfWrapperFunc) error {
	switch cmd := msg.Command.(type) {
	case *message.NetConnectionConnect:
		log.Printf("connect: %+v", msg)

		if err := h.conn.handler.OnConnect(timestamp, cmd); err != nil {
			return err
		}

		// TODO: fix
		if err := h.conn.streamer.Write(chunkStreamID, timestamp, &message.WinAckSize{
			Size: h.conn.streamer.windowSize,
		}); err != nil {
			return err
		}

		// TODO: fix
		if err := h.conn.streamer.Write(chunkStreamID, timestamp, &message.SetPeerBandwidth{
			Size:  1 * 1024 * 1024,
			Limit: 1,
		}); err != nil {
			return err
		}

		// TODO: fix
		m := wrapper(func(cmsg *message.CommandMessage) {
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
		if err := h.conn.streamer.Write(chunkStreamID, timestamp, m); err != nil {
			return err
		}
		log.Printf("connected")

		// next to creatingStreamMessageHandler
		h.conn.stateHandler = &creatingStreamMessageHandler{
			conn: h.conn,
		}

		return nil

	default:
		log.Printf("Unexpected command(connect): %+v", cmd)
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
