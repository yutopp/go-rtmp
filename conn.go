//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bufio"
	"github.com/pkg/errors"
	"log"
	"net"

	"github.com/yutopp/go-rtmp/handshake"
	"github.com/yutopp/go-rtmp/message"
)

type stateID int

const (
	stateIDInvalid stateID = iota
	stateIDConnecting
	stateIDCreateingStream
	stateIDControllingStream
	stateIDPublishing
)

type ConnHandler func(message.Message, uint64, Stream) error

// Server Connection
// TODO: rename or add prefix (Server/Client)
type Conn struct {
	rwc       net.Conn
	bufr      *bufio.Reader
	bufw      *bufio.Writer
	transport *ChunkStreamLayer
	writer    *ChunkStreamWriter
	stateID   stateID

	streamer *ChunkStreamer

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
			err, ok := r.(error)
			if !ok {
				err = errors.Errorf("Panic message: %+v", r)
			}
			err = errors.WithStack(err)
			// TODO: fix
			log.Printf("Panic: %+v", err)
		}
	}()
	defer c.rwc.Close()

	if err := handshake.HandshakeWithClient(c.rwc, c.rwc); err != nil {
		return err
	}

	c.bufr = bufio.NewReaderSize(c.rwc, 4*1024) // TODO: fix buffer size
	c.bufw = bufio.NewWriterSize(c.rwc, 4*1024) // TODO: fix buffer size

	handler := &Handler{
		OnMessage: func(msg message.Message, timestamp uint64, s Stream) {
			c.handleMessage(msg, timestamp, s)
		},
	}
	c.transport = NewChunkStreamLayer(c.bufr, c.bufw, handler)
	c.streamer = NewChunkStreamer(c.bufr, c.bufw)

	c.stateID = stateIDConnecting // nextState: wait for "connect"

	for {
		var msg message.Message
		chunkStreamID, timestamp, err := c.read(&msg)
		if err != nil {
			return err
		}

		stream := &ChunkStreamIO{
			streamID: chunkStreamID,
			f: func(msg message.Message, streamID int) error {
				return c.write(msg, streamID)
			},
		}
		if err := c.handleMessage(msg, timestamp, stream); err != nil {
			return err
		}
	}
}

func (c *Conn) read(msg *message.Message) (int, uint64, error) {
	reader, err := c.streamer.NewChunkReader()
	if err != nil {
		return 0, 0, err
	}
	defer reader.Close()

	dec := message.NewDecoder(reader, reader.messageTypeID)
	if err := dec.Decode(msg); err != nil {
		return 0, 0, err
	}

	return reader.basicHeader.chunkStreamID, reader.timestamp, nil
}

func (c *Conn) write(msg message.Message, chunkStreamID int) error {
	writer, err := c.streamer.NewChunkWriter(chunkStreamID)
	if err != nil {
		return err
	}
	//defer writer.Close()

	enc := message.NewEncoder(writer)
	if err := enc.Encode(msg); err != nil {
		return err
	}
	writer.messageLength = uint32(writer.buf.Len())
	writer.messageTypeID = byte(msg.TypeID())

	return c.streamer.Sched(writer)
}

func (c *Conn) handleMessage(msg message.Message, timestamp uint64, s Stream) (err error) {
	switch c.stateID {
	case stateIDConnecting:
		err = c.handleConnectMessage(msg, timestamp, s)
	case stateIDCreateingStream:
		err = c.handleCreateStreamMessage(msg, timestamp, s)
	case stateIDControllingStream:
		err = c.handleControllingMessage(msg, timestamp, s)
	case stateIDPublishing:
		err = c.handlePublishStreamMessage(msg, timestamp, s)
	default:
		panic("unexpected state") // TODO: fix
	}
	if err != nil {
		return
	}

	c.bufw.Flush()
	return nil
}

func (c *Conn) handleConnectMessage(msg message.Message, timestamp uint64, s Stream) error {
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		switch msg.CommandName {
		case "connect":
			log.Printf("connect")

			// TODO: fix
			if err := s.Write(&message.CtrlWinAckSize{Size: 1 * 1024 * 1024}); err != nil {
				return err
			}

			// TODO: fix
			if err := s.Write(&message.SetPeerBandwidth{Size: 1 * 1024 * 1024, Limit: 1}); err != nil {
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
			if err := s.Write(m); err != nil {
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

func (c *Conn) handleCreateStreamMessage(msg message.Message, timestamp uint64, s Stream) error {
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

			if err := s.Write(m); err != nil {
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

func (c *Conn) handleControllingMessage(msg message.Message, timestamp uint64, s Stream) error {
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
			if err := s.Write(m); err != nil {
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

func (c *Conn) handlePublishStreamMessage(msg message.Message, timestamp uint64, s Stream) error {
	switch msg := msg.(type) {
	case *message.AudioMessage:
		return c.userHandler(msg, timestamp, s)
	case *message.VideoMessage:
		return c.userHandler(msg, timestamp, s)
	default:
		log.Printf("unexpected message: %+v", msg)
		return nil
	}
}
