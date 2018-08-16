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
	"sync"

	"github.com/yutopp/go-rtmp/handshake"
	"github.com/yutopp/go-rtmp/message"
)

// Server Connection
// TODO: rename or add prefix (Server/Client)
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
	SkipHandshakeVerification               bool
	IgnoreMessagesOnNotExistStream          bool
	IgnoreMessagesOnNotExistStreamThreshold uint32

	MaxBitrateKbps uint32

	ReaderBufferSize int
	WriterBufferSize int

	ControlState StreamControlStateConfig
	//DefaultReadTimeout time.Duration
	//DefaultWriteTimeout time.Duration
}

func (cb *ConnConfig) normalize() *ConnConfig {
	c := ConnConfig(*cb)

	if c.MaxBitrateKbps == 0 {
		c.MaxBitrateKbps = 8 * 1024 // 8MBps (Default)
	}

	if c.ReaderBufferSize == 0 {
		c.ReaderBufferSize = 4 * 1024 // 4KB (Default)
	}

	if c.WriterBufferSize == 0 {
		c.WriterBufferSize = 4 * 1024 // 4KB (Default)
	}

	c.ControlState = *c.ControlState.normalize()

	return &c
}

func NewConn(rwc io.ReadWriteCloser, config *ConnConfig) *Conn {
	if config == nil {
		config = &ConnConfig{}
	}
	config = config.normalize()

	conn := &Conn{
		rwc:     rwc,
		handler: &DefaultHandler{},

		config: config,
		logger: logrus.StandardLogger(),
	}
	conn.streams = newStreams(conn, &config.ControlState)

	return conn
}

func (c *Conn) SetHandler(h Handler) {
	// TODO: return error if conn is already served
	c.handler = h
}

func (c *Conn) SetLogger(l logrus.FieldLogger) {
	// TODO: return error if conn is already served
	c.logger = l
}

func (c *Conn) Serve() (err error) {
	defer func() {
		if r := recover(); r != nil {
			errTmp, ok := r.(error)
			if !ok {
				errTmp = errors.Errorf("Panic: %+v", r)
			}
			err = errors.WithStack(errTmp)
		}
	}()

	if err := handshake.HandshakeWithClient(c.rwc, c.rwc, &handshake.Config{
		SkipHandshakeVerification: c.config.SkipHandshakeVerification,
	}); err != nil {
		return err
	}

	c.bufr = bufio.NewReaderSize(c.rwc, c.config.ReaderBufferSize)
	c.bufw = bufio.NewWriterSize(c.rwc, c.config.WriterBufferSize)

	c.streamer = NewChunkStreamer(
		NewBitrateRejectorReader(c.bufr, c.config.MaxBitrateKbps),
		c.bufw,
		&c.config.ControlState,
	)
	c.streamer.logger = c.logger

	// StreamID 0 is default control stream
	const DefaultControlStreamID = 0
	if err := c.streams.Create(DefaultControlStreamID, &controlStreamHandler{
		streamer: c.streamer,
		streams:  c.streams,
		handler:  c.handler,
		logger:   c.logger,
	}); err != nil {
		return err
	}

	defaultStream, ok := c.streams.At(DefaultControlStreamID)
	if !ok {
		return errors.New("Unexpected: default stream is not found")
	}
	c.streamer.controlStreamWriter = defaultStream.write

	if c.handler != nil {
		c.handler.OnServe()
	}

	return c.serveLoop()
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
		if err := c.streamer.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	if err := c.rwc.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (c *Conn) serveLoop() error {
	var streamFragment StreamFragment
	for {
		select {
		case <-c.streamer.Done():
			return c.streamer.Err()

		default:
			chunkStreamID, timestamp, err := c.streamer.Read(&streamFragment)
			if err != nil {
				switch err := err.(type) {
				case *message.UnknownAMFParseError:
					// Ignore unknown amf object
					c.logger.Warnf("Ignored unknown amf packed message: Err = %+v", err)
					continue
				}
				return err
			}

			if err := c.handleStreamFragment(chunkStreamID, timestamp, &streamFragment); err != nil {
				return err // Shutdown the connection
			}
		}
	}
}

func (c *Conn) handleStreamFragment(chunkStreamID int, timestamp uint32, sf *StreamFragment) error {
	stream, ok := c.streams.At(sf.StreamID)
	if !ok {
		if c.config.IgnoreMessagesOnNotExistStream {
			c.logger.Warnf("Messages are received on not exist streams: StreamID = %d, Message = %#v",
				sf.StreamID,
				sf.Message,
			)

			if c.ignoredMessages < c.config.IgnoreMessagesOnNotExistStreamThreshold {
				c.ignoredMessages++
				return nil
			}
		}

		return errors.Errorf("Specified stream is not created yet: StreamID = %d", sf.StreamID)
	}

	if err := stream.handler.Handle(chunkStreamID, timestamp, sf.Message, stream); err != nil {
		return err
	}

	return nil
}

func (c *Conn) createStream(streamID uint32, handler streamHandler) error {
	return c.streams.Create(streamID, handler)
}

func (c *Conn) createStreamIfAvailable(handler streamHandler) (uint32, error) {
	return c.streams.CreateIfAvailable(handler)
}

func (c *Conn) deleteStream(streamID uint32) error {
	return c.streams.Delete(streamID)
}
