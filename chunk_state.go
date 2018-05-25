//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"time"
)

var timeNow = time.Now // For mock

type ChunkState struct {
	streams          map[int]*StreamState
	initialTimestamp time.Time
}

func NewChunkState() *ChunkState {
	return &ChunkState{
		streams:          make(map[int]*StreamState),
		initialTimestamp: timeNow(),
	}
}

func (c *ChunkState) StreamState(streamID int) *StreamState {
	state, ok := c.streams[streamID]
	if !ok {
		state = &StreamState{
			chunkSize: 128, // TODO: default
		}
		c.streams[streamID] = state
	}

	return state
}

type StreamState struct {
	chunkSize   uint32
	readerState *StreamReaderState
	writerState *StreamWriterState
}

func (s *StreamState) ReaderState() *StreamReaderState {
	if s.readerState == nil {
		s.readerState = &StreamReaderState{}
	}

	return s.readerState
}

func (s *StreamState) WriterState() *StreamWriterState {
	if s.writerState == nil {
		s.writerState = &StreamWriterState{
			timestamp: 0xffffffff,
		}
	}

	return s.writerState
}

type StreamReaderState struct {
	decoding bool

	restLength uint32

	timestamp       uint32
	timestampDelta  uint32
	messageLength   uint32
	messageTypeID   byte
	messageStreamID uint32

	messageBuffer []byte
}

type StreamWriterState struct {
	encoding   bool
	restLength uint32

	timestamp         uint32
	timestampForDelta uint32
	messageLength     uint32
	messageTypeID     byte
	messageStreamID   uint32

	cacheBuf []byte
}
