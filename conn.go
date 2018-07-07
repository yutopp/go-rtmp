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
	"github.com/sirupsen/logrus"
	"io"
	"sync"

	"github.com/yutopp/go-rtmp/handshake"
)

// Server Connection
// TODO: rename or add prefix (Server/Client)
type Conn struct {
	rwc      io.ReadWriteCloser
	bufr     *bufio.Reader
	bufw     *bufio.Writer
	streamer *ChunkStreamer
	streams  map[uint32]*Stream
	streamsM sync.Mutex
	handler  Handler

	config *ConnConfig
	logger logrus.FieldLogger
}

type ConnConfig struct {
	SkipHandshakeVerification bool
	MaxStreams                uint32

	ReaderBufferSize int
	WriterBufferSize int
}

func (cb *ConnConfig) normalize() *ConnConfig {
	c := ConnConfig(*cb)

	if c.MaxStreams == 0 {
		c.MaxStreams = 10 // Default value
	}

	if c.ReaderBufferSize == 0 {
		c.ReaderBufferSize = 4 * 1024 // Default value
	}

	if c.WriterBufferSize == 0 {
		c.WriterBufferSize = 4 * 1024 // Default value
	}

	return &c
}

func NewConn(rwc io.ReadWriteCloser, config *ConnConfig) *Conn {
	if config == nil {
		config = &ConnConfig{}
	}
	config = config.normalize()

	return &Conn{
		rwc:     rwc,
		handler: &NopHandler{},
		streams: make(map[uint32]*Stream),

		config: config,
		logger: logrus.StandardLogger(),
	}
}

func (c *Conn) SetHandler(h Handler) {
	c.handler = h
}

func (c *Conn) SetLogger(l logrus.FieldLogger) {
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

	c.streamer = NewChunkStreamer(c.bufr, c.bufw)
	c.streamer.logger = c.logger

	// StreamID 0 is default control stream
	const DefaultControlStreamID = 0
	if err := c.createStream(DefaultControlStreamID, &controlStreamHandler{
		conn:   c,
		logger: c.logger,
	}); err != nil {
		return err
	}

	c.streamer.controlStreamWriter = c.streams[DefaultControlStreamID].Write

	var streamFragment StreamFragment
	for {
		select {
		case <-c.streamer.Done():
			return c.streamer.Err()

		default:
			chunkStreamID, timestamp, err := c.streamer.Read(&streamFragment)
			if err != nil {
				return err
			}

			if err := c.handleStreamFragment(chunkStreamID, timestamp, &streamFragment); err != nil {
				return err
			}
		}
	}
}

func (c *Conn) Close() error {
	if c.handler != nil {
		c.handler.OnClose()
	}

	if c.streamer != nil {
		_ = c.streamer.Close()
	}

	return c.rwc.Close()
}

func (c *Conn) handleStreamFragment(chunkStreamID int, timestamp uint32, sf *StreamFragment) error {
	stream, ok := c.streams[sf.StreamID]
	if !ok {
		return errors.Errorf("Specified stream is not created yet: StreamID = %d", sf.StreamID)
	}

	if err := stream.handler.Handle(chunkStreamID, timestamp, sf.Message, stream); err != nil {
		return err
	}

	return nil
}

func (c *Conn) createStream(streamID uint32, handler streamHandler) error {
	c.streamsM.Lock()
	defer c.streamsM.Unlock()

	_, ok := c.streams[streamID]
	if ok {
		return errors.Errorf("Stream already exists: StreamID = %d", streamID)
	}

	c.streams[streamID] = &Stream{
		streamID: streamID,
		handler:  handler,
		conn:     c,
		fragment: StreamFragment{
			StreamID: streamID,
		},
	}

	return nil
}

func (c *Conn) createStreamIfAvailable(handler streamHandler) (uint32, error) {
	for i := uint32(0); i < c.config.MaxStreams; i++ {
		if err := c.createStream(i, handler); err != nil {
			continue
		}
		return i, nil
	}

	return 0, errors.Errorf("Creating streams limit exceeded: Limit = %d", c.config.MaxStreams)
}

func (c *Conn) deleteStream(streamID uint32) error {
	c.streamsM.Lock()
	defer c.streamsM.Unlock()

	_, ok := c.streams[streamID]
	if !ok {
		return errors.Errorf("Stream not exists: StreamID = %d", streamID)
	}

	delete(c.streams, streamID)

	return nil
}
