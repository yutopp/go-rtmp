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
)

type ChunkStreamer struct {
	r io.Reader
	w io.Writer

	chunkSize uint32
	readers   map[int]*ChunkStreamReader
}

func NewChunkStreamer(r io.Reader, w io.Writer) *ChunkStreamer {
	return &ChunkStreamer{
		r: r,
		w: w,

		chunkSize: 128, // TODO fix
		readers:   make(map[int]*ChunkStreamReader),
	}
}

func (cs *ChunkStreamer) NewChunkReader() (*ChunkStreamReader, error) {
again:
	reader, err := cs.readChunk()
	if err != nil {
		return nil, err
	}
	if reader == nil {
		goto again
	}
	return reader, nil
}

// returns nil reader when chunk is fragmented.
func (cs *ChunkStreamer) readChunk() (*ChunkStreamReader, error) {
	var bh chunkBasicHeader
	if err := decodeChunkBasicHeader(cs.r, &bh); err != nil {
		return nil, err
	}
	log.Printf("basicHeader = %+v", bh)

	var mh chunkMessageHeader
	if err := decodeChunkMessageHeader(cs.r, bh.fmt, &mh); err != nil {
		return nil, err
	}
	log.Printf("messageHeader = %+v", mh)

	reader, ok := cs.readers[bh.chunkStreamID]
	if !ok {
		reader = &ChunkStreamReader{
			buf: new(bytes.Buffer),
		}
		cs.readers[bh.chunkStreamID] = reader
	}
	reader.basicHeader = bh
	reader.messageHeader = mh

	switch bh.fmt {
	case 0:
		reader.timestamp = uint64(mh.timestamp)
		reader.messageLength = mh.messageLength
		reader.messageTypeID = mh.messageTypeID
		reader.messageStreamID = mh.messageStreamID

	case 1:
		reader.timestampDelta = mh.timestampDelta
		reader.timestamp += uint64(reader.timestampDelta)
		reader.messageLength = mh.messageLength
		reader.messageTypeID = mh.messageTypeID

	case 2:
		reader.timestampDelta = mh.timestampDelta
		reader.timestamp += uint64(reader.timestampDelta)

	case 3:

	default:
		panic("unsupported chunk") // TODO: fix
	}

	log.Printf("MessageLength: %d / Rest=%d", reader.messageLength, reader.buf.Len())

	expectLen := int(reader.messageLength) - reader.buf.Len()
	if expectLen <= 0 {
		panic("invalid state") // TODO fix
	}

	if uint32(expectLen) > cs.chunkSize {
		expectLen = int(cs.chunkSize)
	}

	_, err := io.CopyN(reader.buf, cs.r, int64(expectLen))
	if err != nil {
		return nil, err
	}

	if int(reader.messageLength)-reader.buf.Len() != 0 {
		// fragmented
		return nil, nil
	}

	return reader, nil
}
