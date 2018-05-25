//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"io"

	"github.com/yutopp/rtmp-go/internal"
	"github.com/yutopp/rtmp-go/message"
)

type ChunkStreamIO struct {
	streamID int
	f        func(streamID int, msg message.Message) error
}

func (w *ChunkStreamIO) Write(msg message.Message) error {
	return w.f(w.streamID, msg)
}

type ChunkStreamLayer struct {
	r       *ChunkStreamReader
	w       *ChunkStreamWriter
	state   *ChunkState
	handler *Handler
}

func NewChunkStreamLayer(r io.Reader, w io.Writer, h *Handler) *ChunkStreamLayer {
	return &ChunkStreamLayer{
		r:       NewChunkStreamReader(r),
		w:       NewChunkStreamWriter(w),
		state:   NewChunkState(),
		handler: h,
	}
}

func (s *ChunkStreamLayer) Serve() error {
	for {
		msg, streamID, err := s.readMessage()
		if err != nil {
			return nil
		}

		writer := &ChunkStreamIO{
			streamID: streamID,
			f:        s.writeMessage,
		}
		s.handler.OnMessage(msg, writer)
	}
}

func (s *ChunkStreamLayer) readMessageFragment() (int, message.Message, error) {
	return s.r.ReadChunk(s.state)
}

func (s *ChunkStreamLayer) readMessage() (message.Message, int, error) {
	for {
		streamID, msg, err := s.readMessageFragment()
		if err != nil {
			if err == internal.ErrChunkIsNotCompleted {
				continue
			}
			return nil, 0, err
		}
		return msg, streamID, err
	}
}

func (s *ChunkStreamLayer) writeMessageFragment(streamID int, msg message.Message) error {
	return s.w.WriteChunk(s.state, streamID, msg)
}

func (s *ChunkStreamLayer) writeMessage(streamID int, msg message.Message) error {
	for {
		err := s.writeMessageFragment(streamID, msg)
		if err != nil {
			if err == internal.ErrChunkIsNotCompleted {
				msg = nil
				continue
			}
			return err
		}
		return nil
	}
}

// TODO: implement
func (s *ChunkStreamLayer) Close() {
}
