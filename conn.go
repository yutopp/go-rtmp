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

// Server Connection
// TODO: rename or add prefix (Server/Client)
type Conn struct {
	rwc          net.Conn
	bufr         *bufio.Reader
	bufw         *bufio.Writer
	stateHandler stateHandler
	streamer     *ChunkStreamer
	handler      Handler
}

func NewConn(rwc net.Conn, handler Handler) *Conn {
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
	c.stateHandler = &connectMessageHandler{conn: c}

	for {
		var msg message.Message
		chunkStreamID, timestamp, err := c.streamer.Read(&msg)
		if err != nil {
			return err
		}

		if err := c.handleMessage(chunkStreamID, timestamp, msg); err != nil {
			return err
		}
	}
}

func (c *Conn) handleMessage(chunkStreamID int, timestamp uint32, msg message.Message) (err error) {
	if err := c.stateHandler.Handle(chunkStreamID, timestamp, msg); err != nil {
		return err
	}

	return nil
}
