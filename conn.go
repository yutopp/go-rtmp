//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bufio"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"sync"

	"github.com/yutopp/go-rtmp/message"
)

type Conn struct {
	rwc      io.ReadWriteCloser
	bufr     *bufio.Reader
	bufw     *bufio.Writer
	streamer *ChunkStreamer
	streams  *streams
	handler  Handler

	config *ConnConfig
	logger logrus.FieldLogger

	ignoredMessages uint32

	m        sync.Mutex
	isClosed bool
}

type ConnConfig struct {
	Handler                   Handler
	SkipHandshakeVerification bool

	IgnoreMessagesOnNotExistStream          bool
	IgnoreMessagesOnNotExistStreamThreshold uint32

	ReaderBufferSize int
	WriterBufferSize int

	ControlState StreamControlStateConfig

	Logger logrus.FieldLogger
}

func (cb *ConnConfig) normalize() *ConnConfig {
	c := ConnConfig(*cb)

	if c.Handler == nil {
		c.Handler = &DefaultHandler{}
	}

	if c.ReaderBufferSize == 0 {
		c.ReaderBufferSize = 4 * 1024 // 4KB (Default)
	}

	if c.WriterBufferSize == 0 {
		c.WriterBufferSize = 4 * 1024 // 4KB (Default)
	}

	c.ControlState = *c.ControlState.normalize()

	if c.Logger == nil {
		l := logrus.New()
		l.Out = ioutil.Discard

		c.Logger = l
	}

	return &c
}

func newConn(rwc io.ReadWriteCloser, config *ConnConfig) *Conn {
	if config == nil {
		config = &ConnConfig{}
	}
	config = config.normalize()

	conn := &Conn{
		rwc:     rwc,
		bufr:    bufio.NewReaderSize(rwc, config.ReaderBufferSize),
		bufw:    bufio.NewWriterSize(rwc, config.WriterBufferSize),
		handler: config.Handler,

		config: config,
		logger: config.Logger,
	}

	conn.streamer = NewChunkStreamer(conn.bufr, conn.bufw, &conn.config.ControlState)
	conn.streamer.logger = conn.logger

	conn.streams = newStreams(conn)

	return conn
}

func (c *Conn) Close() error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.isClosed {
		return nil
	}
	c.isClosed = true

	if c.handler != nil {
		c.handler.OnClose()
	}

	var result error
	if c.streamer != nil {
		c.streamer.waitWriters()
		if err := c.streamer.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	if err := c.rwc.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (c *Conn) handleMessageLoop() (err error) {
	defer func() {
		if r := recover(); r != nil {
			errTmp, ok := r.(error)
			if !ok {
				errTmp = errors.Errorf("Panic: %+v", r)
			}
			err = errors.WithStack(errTmp)
		}
	}()

	return c.runHandleMessageLoop()
}

func (c *Conn) runHandleMessageLoop() error {
	var cmsg ChunkMessage
	for {
		select {
		case <-c.streamer.Done():
			return c.streamer.Err()

		default:
			chunkStreamID, timestamp, closer, err := c.streamer.Read(&cmsg)
			if err != nil {
				return err
			}

			if err := c.handleMessage(chunkStreamID, timestamp, closer, &cmsg); err != nil {
				return err // Shutdown the connection
			}
		}
	}
}

func (c *Conn) handleMessage(chunkStreamID int, timestamp uint32, closer func(), cmsg *ChunkMessage) error {
	defer closer()

	stream, err := c.streams.At(cmsg.StreamID)
	if err != nil {
		if c.config.IgnoreMessagesOnNotExistStream {
			c.logger.Warnf("Messages are received on not exist streams: StreamID = %d, MessageType = %T",
				cmsg.StreamID,
				cmsg.Message,
			)

			if c.ignoredMessages < c.config.IgnoreMessagesOnNotExistStreamThreshold {
				c.ignoredMessages++
				return nil
			}
		}

		return errors.Errorf("Specified stream is not created yet: StreamID = %d", cmsg.StreamID)
	}

	if err := stream.handle(chunkStreamID, timestamp, cmsg.Message); err != nil {
		switch err := err.(type) {
		case *message.UnknownDataBodyDecodeError, *message.UnknownCommandBodyDecodeError:
			// Ignore unknown messsage body
			c.logger.Warnf("Ignored unknown message body: Err = %+v", err)
			return nil
		}
		return err
	}

	return nil
}
