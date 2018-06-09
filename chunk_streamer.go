//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"math"
	"sync"
)

type ChunkStreamer struct {
	bufr *bufio.Reader
	bufw *bufio.Writer

	chunkSize uint32
	readers   map[int]*ChunkStreamReader
	writers   map[int]*ChunkStreamWriter

	writerSched *chunkStreamerWriterSched
}

func NewChunkStreamer(bufr *bufio.Reader, bufw *bufio.Writer) *ChunkStreamer {
	cs := &ChunkStreamer{
		bufr: bufr,
		bufw: bufw,

		chunkSize: 128, // TODO fix
		readers:   make(map[int]*ChunkStreamReader),
		writers:   make(map[int]*ChunkStreamWriter),

		writerSched: &chunkStreamerWriterSched{
			isActive: make(chan bool, 1),
			writers:  make(map[int]*ChunkStreamWriter),
		},
	}
	cs.writerSched.streamer = cs
	go cs.schedWriteLoop()

	return cs
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

func (cs *ChunkStreamer) NewChunkWriter(chunkStreamID int) (*ChunkStreamWriter, error) {
	writer := cs.prepareChunkWriter(chunkStreamID)
	writer.m.Lock()
	defer writer.m.Unlock()

	return writer, nil
}

func (cs *ChunkStreamer) Sched(writer *ChunkStreamWriter) error {
	return cs.writerSched.sched(writer)
}

func (cs *ChunkStreamer) Close() error {
	return nil
}

// returns nil reader when chunk is fragmented.
func (cs *ChunkStreamer) readChunk() (*ChunkStreamReader, error) {
	var bh chunkBasicHeader
	if err := decodeChunkBasicHeader(cs.bufr, &bh); err != nil {
		return nil, err
	}
	log.Printf("(READ) BasicHeader = %+v", bh)

	var mh chunkMessageHeader
	if err := decodeChunkMessageHeader(cs.bufr, bh.fmt, &mh); err != nil {
		return nil, err
	}
	log.Printf("(READ) MessageHeader = %+v", mh)

	reader := cs.prepareChunkReader(bh.chunkStreamID)
	reader.basicHeader = bh
	reader.messageHeader = mh

	switch bh.fmt {
	case 0:
		reader.timestamp = uint64(mh.timestamp)
		reader.timestampDelta = 0 // reset
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
		reader.timestamp += uint64(reader.timestampDelta)

	default:
		panic("unsupported chunk") // TODO: fix
	}

	log.Printf("(READ) MessageLength = %d, Current = %d", reader.messageLength, reader.buf.Len())

	expectLen := int(reader.messageLength) - reader.buf.Len()
	if expectLen <= 0 {
		panic("invalid state") // TODO fix
	}

	if uint32(expectLen) > cs.chunkSize {
		expectLen = int(cs.chunkSize)
	}

	_, err := io.CopyN(reader.buf, cs.bufr, int64(expectLen))
	if err != nil {
		return nil, err
	}
	//log.Printf("(READ) Buffer: %+v", reader.buf.Bytes())

	if int(reader.messageLength)-reader.buf.Len() != 0 {
		// fragmented
		return nil, nil
	}

	return reader, nil
}

func (cs *ChunkStreamer) writeChunk(writer *ChunkStreamWriter) (bool, error) {
	fmt := byte(2) // default: only timestamp delta
	if writer.messageHeader.messageLength != writer.messageLength || writer.messageTypeID != writer.messageHeader.messageTypeID {
		// header or type id is updated, change fmt to 1 to notify difference and update state
		writer.messageHeader.messageLength = writer.messageLength
		writer.messageHeader.messageTypeID = writer.messageTypeID
		fmt = 1
	}
	if writer.timestamp != writer.messageHeader.timestamp {
		if writer.timestamp >= writer.messageHeader.timestamp {
			writer.messageHeader.timestampDelta = writer.timestamp - writer.messageHeader.timestamp
			writer.messageHeader.timestamp = writer.timestamp
		} else {
			// timestamp is reversed, clear timestamp data
			fmt = 0
			writer.messageHeader.timestampDelta = 0
			writer.messageHeader.timestamp = writer.timestamp
		}
	}
	writer.basicHeader.fmt = fmt

	log.Printf("(WRITE) Headers: Basic = %+v / Message = %+v", writer.basicHeader, writer.messageHeader)
	//log.Printf("(WRITE) Buffer: %+v", writer.buf.Bytes())

	expectLen := writer.buf.Len()
	if uint32(expectLen) > cs.chunkSize {
		expectLen = int(cs.chunkSize)
	}

	if err := encodeChunkBasicHeader(cs.bufw, &writer.basicHeader); err != nil {
		return false, err
	}
	if err := encodeChunkMessageHeader(cs.bufw, writer.basicHeader.fmt, &writer.messageHeader); err != nil {
		return false, err
	}

	if _, err := io.CopyN(cs.bufw, writer, int64(expectLen)); err != nil {
		return false, err
	}
	if err := cs.bufw.Flush(); err != nil {
		return false, err
	}

	if writer.buf.Len() != 0 {
		// fragmented
		return false, nil
	}

	return true, nil
}

func (cs *ChunkStreamer) schedWriteLoop() {
	// TODO: error handling
	_ = cs.writerSched.run()
}

func (cs *ChunkStreamer) prepareChunkReader(chunkStreamID int) *ChunkStreamReader {
	reader, ok := cs.readers[chunkStreamID]
	if !ok {
		reader = &ChunkStreamReader{
			buf: new(bytes.Buffer),
		}
		cs.readers[chunkStreamID] = reader
	}

	return reader
}

func (cs *ChunkStreamer) prepareChunkWriter(chunkStreamID int) *ChunkStreamWriter {
	writer, ok := cs.writers[chunkStreamID]
	if !ok {
		writer = &ChunkStreamWriter{
			basicHeader: chunkBasicHeader{
				chunkStreamID: chunkStreamID,
			},
			messageHeader: chunkMessageHeader{
				timestamp: math.MaxUint32, // initial state will be updated by writer.timestamp
			},
		}
		cs.writers[chunkStreamID] = writer
	}

	return writer
}

type chunkStreamerWriterSched struct {
	streamer *ChunkStreamer
	isActive chan bool
	m        sync.Mutex
	writers  map[int]*ChunkStreamWriter
}

func (sched *chunkStreamerWriterSched) sched(writer *ChunkStreamWriter) error {
	sched.m.Lock()
	defer sched.m.Unlock()

	_, ok := sched.writers[writer.basicHeader.chunkStreamID]
	if ok {
		return errors.New("Running writer")
	}

	writer.m.Lock()
	sched.writers[writer.basicHeader.chunkStreamID] = writer

	if len(sched.writers) > 0 {
		sched.isActive <- true
	}

	return nil
}

func (sched *chunkStreamerWriterSched) unSched(writer *ChunkStreamWriter) error {
	// Lock must be taken before calling this function.

	_, ok := sched.writers[writer.basicHeader.chunkStreamID]
	if !ok {
		return errors.New("Not running writer")
	}

	writer.m.Unlock()
	delete(sched.writers, writer.basicHeader.chunkStreamID)

	return nil
}

func (sched *chunkStreamerWriterSched) run() error {
	for {
		select {
		case <-sched.isActive:
			// TODO: error handling
			_ = sched.runActives()
		}
	}
}

func (sched *chunkStreamerWriterSched) runActives() error {
	sched.m.Lock()
	defer sched.m.Unlock()

	for _, writer := range sched.writers {
		isCompleted, err := sched.streamer.writeChunk(writer)
		// TODO: error handling
		if isCompleted || err != nil {
			_ = sched.unSched(writer)
		}
	}

	if len(sched.writers) > 0 {
		sched.isActive <- true
	}

	return nil
}
