//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
	"errors"
	"io"
	"log"
	"time"

	"github.com/yutopp/rtmp-go/internal"
	"github.com/yutopp/rtmp-go/message"
)

type ChunkStreamWriter struct {
	w           io.Writer
	chunkStates map[int]*chunkStreamWriteState

	errCh chan error
}

type chunkStreamWriteState struct {
}

func NewChunkStreamWriter(w io.Writer) *ChunkStreamWriter {
	return &ChunkStreamWriter{
		w:           w,
		chunkStates: make(map[int]*chunkStreamWriteState),
		errCh:       make(chan error),
	}
}

// TODO: fix interface
func (cw *ChunkStreamWriter) WriteChunk(chunkState *ChunkState, streamID int, msg message.Message) error {
	streamState := chunkState.StreamState(streamID)
	state := streamState.WriterState()

	if state.encoding && msg != nil {
		return errors.New("Value of msg must be empty when encoding")
	}

	fmt := byte(2) // default: only timestamp delta
	if !state.encoding {
		if msg == nil {
			return errors.New("Value of msg is empty")
		}

		// initial state
		buf := bytes.NewBuffer(state.cacheBuf[:0])
		enc := message.NewEncoder(buf)
		if err := enc.Encode(msg); err != nil {
			return err
		}
		state.cacheBuf = buf.Bytes()

		typeID := byte(msg.TypeID())
		messageLength := uint32(len(state.cacheBuf)) // TODO: check overflow
		if state.messageTypeID != typeID || state.messageLength != messageLength {
			fmt = 1 // update message type id and message length
		}
		state.messageTypeID = typeID
		state.messageLength = messageLength
		state.restLength = messageLength

		relTimestampNs := timeNow().Sub(chunkState.initialTimestamp).Nanoseconds()
		timestamp := uint32(relTimestampNs / int64(time.Millisecond))
		if timestamp < state.timestamp {
			fmt = 0 // timestamp is reversed, update all
			state.timestampForDelta = timestamp
		}
		state.timestamp = timestamp
	}
	state.encoding = true

	bh := &chunkBasicHeader{
		fmt:           fmt,
		chunkStreamID: streamID,
	}
	mh := &chunkMessageHeader{
		timestamp:       state.timestamp,
		timestampDelta:  state.timestamp - state.timestampForDelta,
		messageLength:   state.messageLength,
		messageTypeID:   state.messageTypeID,
		messageStreamID: uint32(0), // always 0(other value is not supported)
	}

	log.Printf("Basic: %+v / Body: %+v", bh, mh)

	expectLen := state.restLength
	if expectLen > streamState.chunkSize {
		expectLen = streamState.chunkSize
	}

	offset := state.messageLength - state.restLength

	if state.restLength == 0 {
		panic("invalid state")
	}

	if err := encodeChunkBasicHeader(cw.w, bh); err != nil {
		return err
	}
	if err := encodeChunkMessageHeader(cw.w, bh.fmt, mh); err != nil {
		return err
	}

	if _, err := cw.w.Write(state.cacheBuf[offset : offset+expectLen]); err != nil {
		return err
	}

	state.restLength -= expectLen
	if state.restLength != 0 {
		return internal.ErrChunkIsNotCompleted
	}

	state.encoding = false

	return nil
}
