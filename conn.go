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

type ConnHandler interface {
	OnConnect()
	OnPublish()
	OnPlay()
	OnAudio(timestamp uint32, payload []byte) error
	OnVideo(timestamp uint32, payload []byte) error
}

// Server Connection
// TODO: rename or add prefix (Server/Client)
type Conn struct {
	rwc      net.Conn
	bufr     *bufio.Reader
	bufw     *bufio.Writer
	stateID  stateID
	streamer *ChunkStreamer
	handler  ConnHandler
}

func NewConn(rwc net.Conn, handler ConnHandler) *Conn {
	return &Conn{
		rwc:     rwc,
		handler: handler,
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
	c.streamer = NewChunkStreamer(c.bufr, c.bufw)
	c.stateID = stateIDConnecting // nextState: wait for "connect"

	for {
		var msg message.Message
		chunkStreamID, timestamp, err := c.read(&msg)
		if err != nil {
			return err
		}

		if err := c.handleMessage(chunkStreamID, timestamp, msg); err != nil {
			return err
		}
	}
}

func (c *Conn) read(msg *message.Message) (int, uint32, error) {
	reader, err := c.streamer.NewChunkReader()
	if err != nil {
		return 0, 0, err
	}
	defer reader.Close()

	dec := message.NewDecoder(reader, reader.messageTypeID)
	if err := dec.Decode(msg); err != nil {
		return 0, 0, err
	}

	return reader.basicHeader.chunkStreamID, uint32(reader.timestamp), nil
}

func (c *Conn) write(chunkStreamID int, timestamp uint32, msg message.Message) error {
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
	writer.timestamp = timestamp

	return c.streamer.Sched(writer)
}

func (c *Conn) handleMessage(chunkStreamID int, timestamp uint32, msg message.Message) (err error) {
	switch c.stateID {
	case stateIDConnecting:
		err = c.handleConnectMessage(chunkStreamID, timestamp, msg)
	case stateIDCreateingStream:
		err = c.handleCreateStreamMessage(chunkStreamID, timestamp, msg)
	case stateIDControllingStream:
		err = c.handleControllingMessage(chunkStreamID, timestamp, msg)
	case stateIDPublishing:
		err = c.handlePublishStreamMessage(chunkStreamID, timestamp, msg)
	default:
		panic("unexpected state") // TODO: fix
	}
	if err != nil {
		return
	}

	return nil
}

func (c *Conn) handleConnectMessage(chunkStreamID int, timestamp uint32, msg message.Message) error {
	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		switch msg.CommandName {
		case "connect":
			log.Printf("connect: %+v", msg)

			// TODO: fix
			if err := c.write(chunkStreamID, timestamp, &message.CtrlWinAckSize{
				Size: 1 * 1024 * 1024,
			}); err != nil {
				return err
			}

			// TODO: fix
			if err := c.write(chunkStreamID, timestamp, &message.SetPeerBandwidth{
				Size:  1 * 1024 * 1024,
				Limit: 1,
			}); err != nil {
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
			if err := c.write(chunkStreamID, timestamp, m); err != nil {
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

func (c *Conn) handleCreateStreamMessage(chunkStreamID int, timestamp uint32, msg message.Message) error {
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

			if err := c.write(chunkStreamID, timestamp, m); err != nil {
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

func (c *Conn) handleControllingMessage(chunkStreamID int, timestamp uint32, msg message.Message) error {
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
			if err := c.write(chunkStreamID, timestamp, m); err != nil {
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

func (c *Conn) handlePublishStreamMessage(chunkStreamID int, timestamp uint32, msg message.Message) error {
	switch msg := msg.(type) {
	case *message.AudioMessage:
		return c.handler.OnAudio(timestamp, msg.Payload)
	case *message.VideoMessage:
		return c.handler.OnAudio(timestamp, msg.Payload)
	default:
		log.Printf("unexpected message: %+v", msg)
		return nil
	}
}
