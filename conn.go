//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net"

	"github.com/yutopp/go-rtmp/handshake"
)

// Server Connection
// TODO: rename or add prefix (Server/Client)
type Conn struct {
	rwc        net.Conn
	bufr       *bufio.Reader
	bufw       *bufio.Writer
	streamer   *ChunkStreamer
	streams    map[uint32]*Stream
	maxStreams uint32
	handler    Handler
}

func NewConn(rwc net.Conn, handler Handler) *Conn {
	return &Conn{
		rwc:        rwc,
		handler:    handler,
		streams:    make(map[uint32]*Stream),
		maxStreams: 10, // TODO: fix
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

	// StreamID 0 is default NetConnection stream
	const DefaultNetConnectionStreamID = 0
	if err := c.createStream(DefaultNetConnectionStreamID, &netConnectionMessageHandler{
		conn: c,
		defaultHandler: &commonMessageHandler{
			conn: c,
		},
	}); err != nil {
		return err
	}

	c.streamer.netConnectionWriter = c.streams[DefaultNetConnectionStreamID].Write

	var streamFragment StreamFragment
	for {
		chunkStreamID, timestamp, err := c.streamer.Read(&streamFragment)
		if err != nil {
			return err
		}

		if err := c.handleStreamFragment(chunkStreamID, timestamp, &streamFragment); err != nil {
			return err
		}
	}
}

func (c *Conn) handleStreamFragment(chunkStreamID int, timestamp uint32, sf *StreamFragment) error {
	stream, ok := c.streams[sf.StreamID]
	if !ok {
		return fmt.Errorf("Specified stream is not created yet: StreamID = %d", sf.StreamID)
	}

	if err := stream.handler.Handle(chunkStreamID, timestamp, sf.Message, stream); err != nil {
		return err
	}

	return nil
}

func (c *Conn) createStream(streamID uint32, handler streamHandler) error {
	_, ok := c.streams[streamID]
	if ok {
		return fmt.Errorf("Stream already exists: StreamID = %d", streamID)
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
	for i := uint32(0); i < c.maxStreams; i++ {
		if err := c.createStream(i, handler); err != nil {
			continue
		}
		return i, nil
	}

	return 0, fmt.Errorf("Creating streams limit exceeded: Limit = %d", c.maxStreams)
}

// TODO: implement
func (c *Conn) deleteStream(streamID uint32) error {
	return nil
}
