//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"io"

	"github.com/yutopp/go-rtmp/message"
)

type ChunkStreamIO struct {
	streamID int
	f        func(msg message.Message, streamID int) error
}

func (w *ChunkStreamIO) Write(msg message.Message) error {
	return w.f(msg, w.streamID)
}

type ChunkStreamLayer struct {
	w       *ChunkStreamWriter
	state   *ChunkState
	handler *Handler
}

func NewChunkStreamLayer(r io.Reader, w io.Writer, h *Handler) *ChunkStreamLayer {
	return &ChunkStreamLayer{
		state:   NewChunkState(),
		handler: h,
	}
}

// TODO: implement
func (s *ChunkStreamLayer) Close() {
}
