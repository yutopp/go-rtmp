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

	"github.com/yutopp/go-rtmp/message"
)

const MaxChunkSize = 0xffffff // 5.4.1

type ChunkStreamer struct {
	r    *ChunkStreamerReader
	bufw *bufio.Writer

	readers map[int]*ChunkStreamReader
	writers map[int]*ChunkStreamWriter

	writerSched *chunkStreamerWriterSched

	chunkSize uint32

	windowSize uint32
	limitType  message.LimitType
	lastAck    uint32

	peerWindowSize uint32
	peerLimitType  message.LimitType

	netConnectionWriter func(chunkStreamID int, timestamp uint32, msg message.Message) error
}

func NewChunkStreamer(bufr *bufio.Reader, bufw *bufio.Writer) *ChunkStreamer {
	cs := &ChunkStreamer{
		r: &ChunkStreamerReader{
			bufr: bufr,
		},
		bufw: bufw,

		chunkSize: 128, // TODO fix
		readers:   make(map[int]*ChunkStreamReader),
		writers:   make(map[int]*ChunkStreamWriter),

		writerSched: &chunkStreamerWriterSched{
			isActive: make(chan bool, 1),
			writers:  make(map[int]*ChunkStreamWriter),
		},

		windowSize: 100, // TODO: fix
		limitType:  message.LimitTypeHard,

		peerWindowSize: math.MaxUint32,
	}
	cs.writerSched.streamer = cs
	go cs.schedWriteLoop()

	return cs
}

func (cs *ChunkStreamer) Read(sf *StreamFragment) (int, uint32, error) {
	reader, err := cs.NewChunkReader()
	if err != nil {
		return 0, 0, err
	}
	defer reader.Close()

	dec := message.NewDecoder(reader, message.TypeID(reader.messageTypeID))
	if err := dec.Decode(&sf.Message); err != nil {
		return 0, 0, err
	}

	sf.StreamID = reader.messageStreamID

	return reader.basicHeader.chunkStreamID, uint32(reader.timestamp), nil
}

func (cs *ChunkStreamer) Write(chunkStreamID int, timestamp uint32, sf *StreamFragment) error {
	writer, err := cs.NewChunkWriter(chunkStreamID)
	if err != nil {
		return err
	}
	//defer writer.Close()

	enc := message.NewEncoder(writer)
	if err := enc.Encode(sf.Message); err != nil {
		return err
	}
	writer.timestamp = timestamp
	writer.messageLength = uint32(writer.buf.Len())
	writer.messageTypeID = byte(sf.Message.TypeID())
	writer.messageStreamID = sf.StreamID

	return cs.Sched(writer)
}

func (cs *ChunkStreamer) NewChunkReader() (*ChunkStreamReader, error) {
again:
	isCompleted, reader, err := cs.readChunk()
	if err != nil {
		return nil, err
	}
	if cs.r.totalReadBytes > uint64(cs.peerWindowSize/2) { // TODO: fix size
		if err := cs.sendAck(); err != nil {
			return nil, err
		}
	}

	if !isCompleted {
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
func (cs *ChunkStreamer) readChunk() (bool, *ChunkStreamReader, error) {
	var bh chunkBasicHeader
	if err := decodeChunkBasicHeader(cs.r, &bh); err != nil {
		return false, nil, err
	}
	log.Printf("(READ) BasicHeader = %+v", bh)

	var mh chunkMessageHeader
	if err := decodeChunkMessageHeader(cs.r, bh.fmt, &mh); err != nil {
		return false, nil, err
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
		reader.timestamp += uint64(mh.timestampDelta)
		reader.timestampDelta = 0 // reset
		reader.messageLength = mh.messageLength
		reader.messageTypeID = mh.messageTypeID

	case 2:
		reader.timestamp += uint64(mh.timestampDelta)
		reader.timestampDelta = mh.timestampDelta

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
	log.Printf("(READ) Length = %d", expectLen)

	_, err := io.CopyN(reader.buf, cs.r, int64(expectLen))
	if err != nil {
		return false, nil, err
	}
	//log.Printf("(READ) Buffer: %+v", reader.buf.Bytes())

	if int(reader.messageLength)-reader.buf.Len() != 0 {
		// fragmented
		return false, reader, nil
	}

	return true, reader, nil
}

func (cs *ChunkStreamer) writeChunk(writer *ChunkStreamWriter) (bool, error) {
	cs.updateWriterHeader(writer)

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

func (cs *ChunkStreamer) updateWriterHeader(writer *ChunkStreamWriter) {
	fmt := byte(2) // default: only timestamp delta
	if writer.messageHeader.messageLength != writer.messageLength || writer.messageTypeID != writer.messageHeader.messageTypeID {
		// header or type id is updated, change fmt to 1 to notify difference and update state
		writer.messageHeader.messageLength = writer.messageLength
		writer.messageHeader.messageTypeID = writer.messageTypeID
		fmt = 1
	}
	if writer.timestamp != writer.messageHeader.timestamp {
		if writer.timestamp >= writer.messageHeader.timestamp {
			writer.timestampDelta = writer.timestamp - writer.messageHeader.timestamp
		} else {
			// timestamp is reversed, clear timestamp data
			fmt = 0
			writer.timestampDelta = 0
		}
	}
	if writer.timestampDelta == writer.messageHeader.timestampDelta && fmt == 2 {
		fmt = 3
	}
	writer.messageHeader.timestampDelta = writer.timestampDelta
	writer.messageHeader.timestamp = writer.timestamp

	if writer.messageHeader.messageStreamID != writer.messageStreamID {
		fmt = 0
		writer.messageHeader.messageStreamID = writer.messageStreamID
	}
	writer.basicHeader.fmt = fmt
}

func (cs *ChunkStreamer) schedWriteLoop() {
	// TODO: error handling
	if err := cs.writerSched.run(); err != nil {
		log.Panicf("ERR: %+v", err)
	}
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

func (cs *ChunkStreamer) sendAck() error {
	log.Printf("Sending Ack...")
	// TODO: chunk stream id and fix timestamp
	return cs.netConnectionWriter(2, 0, &message.Ack{
		SequenceNumber: uint32(cs.r.totalReadBytes),
	})
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
			if err := sched.runActives(); err != nil {
				return err
			}
		}
	}
}

func (sched *chunkStreamerWriterSched) runActives() error {
	sched.m.Lock()
	defer sched.m.Unlock()

	for _, writer := range sched.writers {
		isCompleted, err := sched.streamer.writeChunk(writer)
		if isCompleted || err != nil {
			_ = sched.unSched(writer) // TODO: error check
			if err != nil {
				return err
			}
		}
	}

	if len(sched.writers) > 0 {
		sched.isActive <- true
	}

	return nil
}
