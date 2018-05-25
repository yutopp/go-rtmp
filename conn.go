//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bufio"
	"net"

	"log"

	"github.com/yutopp/rtmp-go/handshake"
	"github.com/yutopp/rtmp-go/message"
)

type stateID int

const (
	stateIDInvalid stateID = iota
	stateIDConnecting
	stateIDCreateingStream
	stateIDControllingStream
	stateIDPublishing
)

type ConnHandler func(message.Message, Stream) error

// Server Connection
// TODO: rename or add prefix (Server/Client)
type Conn struct {
	rwc       net.Conn
	bufr      *bufio.Reader
	bufw      *bufio.Writer
	transport *ChunkStreamLayer
	writer    *ChunkStreamWriter
	stateID   stateID

	userHandler ConnHandler
	userData    interface{}
}

func NewConn(rwc net.Conn, handler ConnHandler) *Conn {
	return &Conn{
		rwc:         rwc,
		userHandler: handler,
	}
}

func (c *Conn) Serve() error {
	defer func() {
		if r := recover(); r != nil {
			// TODO: fix
			log.Printf("Panic: %+v", r)
		}
	}()
	defer c.rwc.Close()

	if err := handshake.HandshakeWithClient(c.rwc, c.rwc); err != nil {
		return err
	}

	c.bufr = bufio.NewReaderSize(c.rwc, 4*1024) // TODO: fix buffer size
	c.bufw = bufio.NewWriterSize(c.rwc, 4*1024) // TODO: fix buffer size

	handler := &Handler{
		OnMessage: c.handleMessage,
	}
	c.transport = NewChunkStreamLayer(c.bufr, c.bufw, handler)

	c.stateID = stateIDConnecting // nextState: wait for "connect"

	return c.transport.Serve()
}

func (c *Conn) handleMessage(msg message.Message, w Stream) {
	var err error

	switch c.stateID {
	case stateIDConnecting:
		err = c.handleConnectMessage(msg, w)
	case stateIDCreateingStream:
		err = c.handleCreateStreamMessage(msg, w)
	case stateIDControllingStream:
		err = c.handleControllingMessage(msg, w)
	case stateIDPublishing:
		err = c.handlePublishStreamMessage(msg, w)
	default:
		panic("unexpected state") // TODO: fix
	}
	if err != nil {
		// TODO: handle error
		panic(err)
	}

	c.bufw.Flush()
}

func (c *Conn) handleConnectMessage(msg message.Message, w Stream) error {
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		switch msg.CommandName {
		case "connect":
			log.Printf("connect")

			// TODO: fix
			if err := w.Write(&message.CtrlWinAckSize{Size: 1 * 1024 * 1024}); err != nil {
				return err
			}

			// TODO: fix
			if err := w.Write(&message.SetPeerBandwidth{Size: 1 * 1024 * 1024, Limit: 1}); err != nil {
				return err
			}

			// TODO: fix
			m := &message.CommandMessageAMF0{
				CommandName:   "_result",
				TransactionID: msg.TransactionID,
				Command: &message.NetConnectionResult{
					Objects: []interface{}{
						map[string]interface{}{
							"fmsVer":       "rtmp/testing",
							"capabilities": 250,
							"mode":         1,
						},
						map[string]interface{}{
							"level": "status",
							"code":  "NetConnection.Connect.Success",
							"data": []struct {
								Key   string `amf0:"ecma"`
								Value interface{}
							}{
								{"version", "testing"},
							},
							"application": nil,
						},
					},
				},
			}
			if err := w.Write(m); err != nil {
				return err
			}
			log.Printf("connected")

			c.stateID = stateIDCreateingStream

			return nil

		default:
			log.Printf("unexpected command: %+v", msg)
			return nil
		}

	default:
		log.Printf("unexpected message: %+v", msg)
		return nil
	}
}

func (c *Conn) handleCreateStreamMessage(msg message.Message, w Stream) error {
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		switch msg.CommandName {
		case "createStream":
			m := &message.CommandMessageAMF0{
				CommandName:   "_result",
				TransactionID: msg.TransactionID,
				Command: &message.NetConnectionResult{
					Objects: []interface{}{
						nil,
						20, // TODO: fix
					},
				},
			}

			if err := w.Write(m); err != nil {
				return err
			}
			log.Printf("streamCreated")

			c.stateID = stateIDControllingStream

			return nil

		default:
			log.Printf("unexpected command: %+v", msg)
			return nil
		}

	default:
		log.Printf("unexpected message: %+v", msg)
		return nil
	}
}

func (c *Conn) handleControllingMessage(msg message.Message, w Stream) error {
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		switch msg.CommandName {
		case "publish":
			m := &message.CommandMessageAMF0{
				CommandName:   "onStatus",
				TransactionID: 0,
				Command: &message.NetStreamOnStatus{
					InfoObject: map[string]interface{}{
						"level":       "status",
						"code":        "NetStream.Publish.Start",
						"description": "yoyo",
					},
				},
			}
			if err := w.Write(m); err != nil {
				return err
			}

			c.stateID = stateIDPublishing

			return nil

		default:
			log.Printf("unexpected command: %+v", msg)
			return nil
		}

	default:
		log.Printf("unexpected message: %+v", msg)
		return nil
	}
}

func (c *Conn) handlePublishStreamMessage(msg message.Message, w Stream) error {
	switch msg := msg.(type) {
	case *message.AudioMessage:
		return c.userHandler(msg, w)
	case *message.VideoMessage:
		return c.userHandler(msg, w)
	default:
		log.Printf("unexpected message: %+v", msg)
		return nil
	}
}
