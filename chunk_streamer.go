//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"context"
	"io"
	"math"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/livekit/go-rtmp/message"
)

const ctrlMsgChunkStreamID = 2

const maxWriterQueueSize = 64

type ChunkMessage struct {
	StreamID uint32
	Message  message.Message
}

type ChunkStreamer struct {
	r *ChunkStreamerReader
	w *ChunkStreamerWriter

	readers map[int]*ChunkStreamReader
	writers map[int]*ChunkStreamWriter
	mu      sync.Mutex

	writerSched *chunkStreamerWriterSched

	msgDec *message.Decoder
	msgEnc *message.Encoder

	selfState *StreamControlState
	peerState *StreamControlState

	err  error
	done chan struct{}

	controlStreamWriter func(chunkStreamID int, timestamp uint32, msg message.Message) error

	cacheBuffer []byte
	config      *StreamControlStateConfig
	logger      logrus.FieldLogger
}

func NewChunkStreamer(r io.Reader, w io.Writer, config *StreamControlStateConfig) *ChunkStreamer {
	if config == nil {
		config = defaultStreamControlStateConfig
	}

	cs := &ChunkStreamer{
		r: &ChunkStreamerReader{
			reader: r,
		},
		w: &ChunkStreamerWriter{
			writer: w,
		},

		readers: make(map[int]*ChunkStreamReader),
		writers: make(map[int]*ChunkStreamWriter),

		writerSched: &chunkStreamerWriterSched{
			writers: make(chan *ChunkStreamWriter, maxWriterQueueSize),
			stopCh:  make(chan struct{}),
		},

		msgDec: message.NewDecoder(nil),
		msgEnc: message.NewEncoder(nil),

		selfState: NewStreamControlState(config),
		peerState: NewStreamControlState(config),

		done: make(chan struct{}),

		cacheBuffer: make([]byte, 64*1024), // cache 64KB
		config:      config,
		logger:      logrus.StandardLogger(),
	}
	cs.writerSched.streamer = cs
	go cs.schedWriteLoop()

	return cs
}

func (cs *ChunkStreamer) Read(cmsg *ChunkMessage) (int, uint32, error) {
	reader, err := cs.NewChunkReader()
	if err != nil {
		return 0, 0, err
	}

	cs.msgDec.Reset(reader)
	if err := cs.msgDec.Decode(message.TypeID(reader.messageTypeID), &cmsg.Message); err != nil {
		return 0, 0, err
	}

	cmsg.StreamID = reader.messageStreamID

	return reader.basicHeader.chunkStreamID, uint32(reader.timestamp), nil
}

func (cs *ChunkStreamer) Write(
	ctx context.Context, // NOTE: Retire writing when a current chunk is busy
	chunkStreamID int,
	timestamp uint32,
	cmsg *ChunkMessage,
) error {
	writer, err := cs.NewChunkWriter(ctx, chunkStreamID)
	if err != nil {
		return err
	}
	//defer writer.Close()

	cs.msgEnc.Reset(writer)
	if err := cs.msgEnc.Encode(cmsg.Message); err != nil {
		return err
	}
	writer.timestamp = timestamp
	writer.messageLength = uint32(writer.buf.Len())
	writer.messageTypeID = byte(cmsg.Message.TypeID())
	writer.messageStreamID = cmsg.StreamID

	return cs.Sched(writer)
}

func (cs *ChunkStreamer) NewChunkReader() (*ChunkStreamReader, error) {
again:
	reader, err := cs.readChunk()
	if err != nil {
		return nil, err
	}
	if cs.r.FragmentReadBytes() >= uint32(cs.peerState.ackWindowSize/2) { // TODO: fix size
		if err := cs.sendAck(cs.r.TotalReadBytes()); err != nil {
			return nil, err
		}
		cs.r.ResetFragmentReadBytes()
	}

	if !reader.completed {
		goto again
	}
	return reader, nil
}

// NewChunkWriter Returns a writer for a chunkStreamID.
// Wait until writing have been finished if the writer is running.
func (cs *ChunkStreamer) NewChunkWriter(ctx context.Context, chunkStreamID int) (*ChunkStreamWriter, error) {
	writer, err := cs.prepareChunkWriter(chunkStreamID)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to prepare chunk writer")
	}
	if err := writer.Wait(ctx); err != nil {
		return nil, errors.Wrapf(err, "Failed to wait chunk writer")
	}

	return writer, nil
}

func (cs *ChunkStreamer) Sched(writer *ChunkStreamWriter) error {
	writer.newChunk = true
	return cs.writerSched.Sched(writer)
}

func (cs *ChunkStreamer) SelfState() *StreamControlState {
	return cs.selfState
}

func (cs *ChunkStreamer) PeerState() *StreamControlState {
	return cs.peerState
}

func (cs *ChunkStreamer) Done() <-chan struct{} {
	return cs.done
}

func (cs *ChunkStreamer) Err() error {
	return cs.err
}

func (cs *ChunkStreamer) Close() error {
	return cs.writerSched.Close()
}

// returns nil reader when chunk is fragmented.
func (cs *ChunkStreamer) readChunk() (*ChunkStreamReader, error) {
	var bh chunkBasicHeader
	if err := decodeChunkBasicHeader(cs.r, cs.cacheBuffer, &bh); err != nil {
		return nil, err
	}
	//cs.logger.Debugf("(READ) BasicHeader = %+v", bh)

	var mh chunkMessageHeader
	if err := decodeChunkMessageHeader(cs.r, bh.fmt, cs.cacheBuffer, &mh); err != nil {
		return nil, err
	}
	//cs.logger.Debugf("(READ) MessageHeader = %+v", mh)

	reader, err := cs.prepareChunkReader(bh.chunkStreamID)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to prepare chunk reader")
	}
	if reader.completed {
		reader.buf.Reset()
		reader.completed = false
	}

	reader.basicHeader = bh
	reader.messageHeader = mh

	switch bh.fmt {
	case 0:
		reader.timestamp = mh.timestamp
		reader.timestampDelta = 0 // reset
		reader.messageLength = mh.messageLength
		reader.messageTypeID = mh.messageTypeID
		reader.messageStreamID = mh.messageStreamID

	case 1:
		reader.timestampDelta = mh.timestampDelta
		reader.messageLength = mh.messageLength
		reader.messageTypeID = mh.messageTypeID

	case 2:
		reader.timestampDelta = mh.timestampDelta

	case 3:
		// DO NOTHING

	default:
		panic("unsupported chunk") // TODO: fix
	}

	//cs.logger.Debugf("(READ) MessageLength = %d, Current = %d", reader.messageLength, reader.buf.Len())

	expectLen := int(reader.messageLength) - reader.buf.Len()
	if expectLen <= 0 {
		panic("invalid state") // TODO fix
	}

	if uint32(expectLen) > cs.peerState.chunkSize {
		expectLen = int(cs.peerState.chunkSize)
	}
	//cs.logger.Debugf("(READ) Length = %d", expectLen)

	lr := io.LimitReader(cs.r, int64(expectLen))
	if _, err := io.CopyBuffer(&reader.buf, lr, cs.cacheBuffer); err != nil {
		return nil, err
	}
	//cs.logger.Debugf("(READ) Buffer: %+v", reader.buf.Bytes())

	if int(reader.messageLength)-reader.buf.Len() != 0 {
		// fragmented
		return reader, nil
	}

	// read completed, update timestamp
	reader.timestamp += reader.timestampDelta
	reader.completed = true

	return reader, nil
}

func (cs *ChunkStreamer) writeChunk(writer *ChunkStreamWriter) (bool, error) {
	cs.updateWriterHeader(writer)

	//cs.logger.Debugf("(WRITE) Headers: Basic = %+v / Message = %+v", writer.basicHeader, writer.messageHeader)
	//cs.logger.Debugf("(WRITE) Buffer: %+v", writer.buf.Bytes())

	expectLen := writer.buf.Len()
	if uint32(expectLen) > cs.selfState.chunkSize {
		expectLen = int(cs.selfState.chunkSize)
	}

	if err := encodeChunkBasicHeader(cs.w, &writer.basicHeader); err != nil {
		return false, err
	}
	if err := encodeChunkMessageHeader(cs.w, writer.basicHeader.fmt, &writer.messageHeader); err != nil {
		return false, err
	}

	if _, err := io.CopyN(cs.w, writer, int64(expectLen)); err != nil {
		return false, err
	}
	if err := cs.w.Flush(); err != nil {
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
	if writer.messageHeader.messageLength != writer.messageLength ||
		writer.messageTypeID != writer.messageHeader.messageTypeID {
		// header or type id is updated, change fmt to 1 to notify difference and update state
		writer.messageHeader.messageLength = writer.messageLength
		writer.messageHeader.messageTypeID = writer.messageTypeID
		fmt = 1
	}
	if writer.timestamp != writer.messageHeader.timestamp || writer.newChunk {
		if writer.timestamp >= writer.messageHeader.timestamp {
			writer.timestampDelta = writer.timestamp - writer.messageHeader.timestamp
		} else {
			// timestamp is reversed, clear timestamp data
			fmt = 0
			writer.timestampDelta = 0
		}
	}
	writer.newChunk = false
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

func (cs *ChunkStreamer) waitWriters() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Wait until that writers are finished for 3 seconds. (NOTE: 3s is addhoc value...)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for k, writer := range cs.writers {
		if err := writer.Wait(ctx); err != nil {
			cs.logger.Warnf("Failed to wait writer: ID = %d", k)
		}
	}
}

func (cs *ChunkStreamer) forceCloseWriters() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for _, writer := range cs.writers {
		//writer.lastErr = cs.err
		close(writer.closeCh)
	}
}

func (cs *ChunkStreamer) schedWriteLoop() {
	defer close(cs.done)
	cs.err = cs.writerSched.Run()

	if cs.err != nil {
		cs.forceCloseWriters()
	}
}

func (cs *ChunkStreamer) prepareChunkReader(chunkStreamID int) (*ChunkStreamReader, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	reader, ok := cs.readers[chunkStreamID]
	if !ok {
		if len(cs.readers) >= cs.config.MaxChunkStreams {
			return nil, errors.Errorf(
				"Creating chunk streams limit exceeded(Reader): Limit = %d",
				cs.config.MaxChunkStreams,
			)
		}

		reader = &ChunkStreamReader{}
		cs.readers[chunkStreamID] = reader
	}

	return reader, nil
}

func (cs *ChunkStreamer) prepareChunkWriter(chunkStreamID int) (*ChunkStreamWriter, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	writer, ok := cs.writers[chunkStreamID]
	if !ok {
		if len(cs.writers) >= cs.config.MaxChunkStreams {
			return nil, errors.Errorf(
				"Creating chunk streams limit exceeded(Writer): Limit = %d",
				cs.config.MaxChunkStreams,
			)
		}

		writer = &ChunkStreamWriter{
			ChunkStreamReader: ChunkStreamReader{
				basicHeader: chunkBasicHeader{
					chunkStreamID: chunkStreamID,
				},
				messageHeader: chunkMessageHeader{
					timestamp: math.MaxUint32, // initial state will be updated by writer.timestamp
				},
			},
			doneCh:   make(chan struct{}),
			closeCh:  make(chan struct{}),
			newChunk: true,
		}
		close(writer.doneCh)
		cs.writers[chunkStreamID] = writer
	}

	return writer, nil
}

func (cs *ChunkStreamer) sendAck(readBytes uint32) error {
	cs.logger.Debugf("Sending Ack...: Bytes = %d", readBytes)
	// TODO: fix timestamp
	return cs.controlStreamWriter(ctrlMsgChunkStreamID, 0, &message.Ack{
		SequenceNumber: readBytes,
	})
}

type chunkStreamerWriterSched struct {
	streamer *ChunkStreamer
	writers  chan *ChunkStreamWriter
	stopCh   chan struct{}
}

func (sched *chunkStreamerWriterSched) Sched(writer *ChunkStreamWriter) error {
	sched.writers <- writer

	return nil
}

func (sched *chunkStreamerWriterSched) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			errTmp, ok := r.(error)
			if !ok {
				errTmp = errors.Errorf("Panic: %+v", r)
			}
			err = errors.WithStack(errTmp)
		}
	}()

	for {
		select {
		case writer := <-sched.writers:
			isCompleted, err := sched.streamer.writeChunk(writer)
			if err != nil {
				writer.lastErr = err
				close(writer.doneCh)
				return err
			}
			if isCompleted {
				close(writer.doneCh)
				continue
			}

			// Enqueue writer
			sched.writers <- writer

		case <-sched.stopCh:
			return nil
		}
	}
}

func (sched *chunkStreamerWriterSched) Close() error {
	close(sched.stopCh)

	return nil
}
