//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
	"io"
	"log"

	"github.com/yutopp/rtmp-go/internal"
	"github.com/yutopp/rtmp-go/message"
)

type ChunkStreamReader struct {
	r io.Reader
}

func NewChunkStreamReader(r io.Reader) *ChunkStreamReader {
	return &ChunkStreamReader{
		r: r,
	}
}

// TODO: fix interface
func (cr *ChunkStreamReader) ReadChunk(chunkState *ChunkState) (int, message.Message, error) {
	var bh chunkBasicHeader
	if err := decodeChunkBasicHeader(cr.r, &bh); err != nil {
		return 0, nil, err
	}
	log.Printf("basicHeader = %+v", bh)

	var mh chunkMessageHeader
	if err := decodeChunkMessageHeader(cr.r, bh.fmt, &mh); err != nil {
		return 0, nil, err
	}
	log.Printf("messageHeader = %+v", mh)

	streamState := chunkState.StreamState(bh.chunkStreamID)
	state := streamState.ReaderState()
	switch bh.fmt {
	case 0:
		if state.decoding {
			panic("in decoding")
		}

		state.decoding = true

		state.timestamp = mh.timestamp
		state.messageLength = mh.messageLength
		state.messageTypeID = mh.messageTypeID
		state.messageStreamID = mh.messageStreamID

		state.restLength = state.messageLength

		// TODO: fix
		if uint32(len(state.messageBuffer)) < state.messageLength {
			n := make([]byte, 0, state.messageLength)
			copy(n, state.messageBuffer)
			state.messageBuffer = n
			log.Printf("Cache buffer updated")
		}

	case 1:
		if state.decoding {
			panic("in decoding")
		}

		state.decoding = true

		state.timestampDelta = mh.timestampDelta
		state.messageLength = mh.messageLength
		state.messageTypeID = mh.messageTypeID

		state.restLength = state.messageLength

		// TODO: fix
		if uint32(len(state.messageBuffer)) < state.messageLength {
			n := make([]byte, 0, state.messageLength)
			copy(n, state.messageBuffer)
			state.messageBuffer = n
			log.Printf("Cache buffer updated")
		}

	case 2:
		if state.decoding {
			panic("in decoding")
		}

		state.decoding = true

		state.timestampDelta = mh.timestampDelta

		state.restLength = state.messageLength

		// TODO: fix
		if uint32(len(state.messageBuffer)) < state.messageLength {
			n := make([]byte, 0, state.messageLength)
			copy(n, state.messageBuffer)
			state.messageBuffer = n
			log.Printf("Cache buffer updated")
		}

	case 3:
		if state.decoding {
			break
		}
		state.decoding = true
		state.restLength = state.messageLength

	default:
		panic("unsupported chunk")
	}

	log.Printf("MessageLength: %d / Rest=%d", state.messageLength, state.restLength)

	expectLen := state.restLength
	if expectLen > streamState.chunkSize {
		expectLen = streamState.chunkSize
	}

	offset := state.messageLength - state.restLength

	log.Printf("Offset: %d / Expect = %d", offset, expectLen)

	if state.restLength == 0 {
		panic("invalid state")
	}

	_, err := io.ReadAtLeast(cr.r, state.messageBuffer[offset:offset+expectLen], int(expectLen))
	if err != nil {
		panic(err)
	}

	log.Printf("BIN: %+x", state.messageBuffer[offset:offset+expectLen])

	state.restLength -= expectLen
	if state.restLength != 0 {
		return 0, nil, internal.ErrChunkIsNotCompleted
	}

	log.Printf("Start decode!")

	state.decoding = false

	// TODO: fix
	buf := bytes.NewBuffer(state.messageBuffer[:state.messageLength])
	dec := message.NewDecoder(buf, state.messageTypeID)
	var msg message.Message
	if err := dec.Decode(&msg); err != nil {
		return 0, nil, err
	}

	return bh.chunkStreamID, msg, nil
}
