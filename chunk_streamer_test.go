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
	"context"
	"fmt"
	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/yutopp/go-rtmp/message"
)

func TestStreamerSingleChunk(t *testing.T) {
	buf := new(bytes.Buffer)
	inbuf := bufio.NewReaderSize(buf, 2048)
	outbuf := bufio.NewWriterSize(buf, 2048)

	streamer := NewChunkStreamer(inbuf, outbuf, nil)

	chunkStreamID := 2
	videoContent := []byte("testtesttest")
	msg := &message.VideoMessage{
		Payload: bytes.NewReader(videoContent),
	}
	timestamp := uint32(72)

	// write a message
	w, err := streamer.NewChunkWriter(context.Background(), chunkStreamID)
	assert.Nil(t, err)
	assert.NotNil(t, w)

	enc := message.NewEncoder(w)
	err = enc.Encode(msg)
	assert.Nil(t, err)
	w.messageLength = uint32(w.buf.Len())
	w.messageTypeID = byte(msg.TypeID())
	w.timestamp = timestamp
	err = streamer.Sched(w)
	assert.Nil(t, err)

	_, err = streamer.NewChunkWriter(context.Background(), chunkStreamID) // wait for writing
	assert.Nil(t, err)

	// read a message
	isCompleted, r, err := streamer.readChunk()
	assert.Nil(t, err)
	assert.True(t, isCompleted)
	assert.NotNil(t, r)
	defer r.Close()

	dec := message.NewDecoder(r)
	var actualMsg message.Message
	err = dec.Decode(message.TypeID(r.messageTypeID), &actualMsg)
	assert.Nil(t, err)
	assert.Equal(t, uint64(timestamp), r.timestamp)

	// check message
	assert.Equal(t, actualMsg.TypeID(), msg.TypeID())
	actualMsgT := actualMsg.(*message.VideoMessage)
	actualContent, _ := ioutil.ReadAll(actualMsgT.Payload)
	assert.Equal(t, actualContent, videoContent)
}

func TestStreamerMultipleChunk(t *testing.T) {
	const chunkSize = 128
	const payloadUnit = "test"

	buf := bytes.NewBuffer(make([]byte, 0, 2048))
	inbuf := bufio.NewReaderSize(buf, 2048)
	outbuf := bufio.NewWriterSize(buf, 2048)

	streamer := NewChunkStreamer(inbuf, outbuf, nil)

	chunkStreamID := 2
	videoContent := []byte(strings.Repeat(payloadUnit, chunkSize))
	msg := &message.VideoMessage{
		// will be chunked (chunkSize * len(payloadUnit))
		Payload: bytes.NewReader(videoContent),
	}
	timestamp := uint32(72)

	// write a message
	w, err := streamer.NewChunkWriter(context.Background(), chunkStreamID)
	assert.Nil(t, err)
	assert.NotNil(t, w)

	enc := message.NewEncoder(w)
	err = enc.Encode(msg)
	assert.Nil(t, err)
	w.messageLength = uint32(w.buf.Len())
	w.messageTypeID = byte(msg.TypeID())
	w.timestamp = timestamp
	err = streamer.Sched(w)
	assert.Nil(t, err)

	_, err = streamer.NewChunkWriter(context.Background(), chunkStreamID) // wait for writing
	assert.Nil(t, err)

	// read a message
	var r *ChunkStreamReader
	for i := 0; i < len(payloadUnit); i++ {
		_, r, err = streamer.readChunk()
		assert.Nil(t, err)
	}
	assert.NotNil(t, r)
	defer r.Close()

	dec := message.NewDecoder(r)
	var actualMsg message.Message
	err = dec.Decode(message.TypeID(r.messageTypeID), &actualMsg)
	assert.Nil(t, err)
	assert.Equal(t, uint64(timestamp), r.timestamp)

	// check message
	assert.Equal(t, actualMsg.TypeID(), msg.TypeID())
	actualMsgT := actualMsg.(*message.VideoMessage)
	actualContent, _ := ioutil.ReadAll(actualMsgT.Payload)
	assert.Equal(t, actualContent, videoContent)
}

func TestStreamerChunkExample1(t *testing.T) {
	type write struct {
		timestamp uint32
		length    int
	}

	type read struct {
		timestamp  uint32
		fmt        byte
		isComplete bool
	}

	type testCase struct {
		name            string
		chunkStreamID   int
		typeID          byte
		messageStreamID uint32
		writeCases      []write
		readCases       []read
	}

	tcs := []testCase{
		// Example #1
		testCase{
			name:            "Example #1",
			chunkStreamID:   3,
			typeID:          8,
			messageStreamID: 12345,
			writeCases: []write{
				write{timestamp: 1000, length: 32},
				write{timestamp: 1020, length: 32},
				write{timestamp: 1040, length: 32},
				write{timestamp: 1060, length: 32},
			},
			readCases: []read{
				read{timestamp: 1000, fmt: 0, isComplete: true},
				read{timestamp: 1020, fmt: 2, isComplete: true},
				read{timestamp: 1040, fmt: 3, isComplete: true},
				read{timestamp: 1060, fmt: 3, isComplete: true},
			},
		},
		// Example #2
		testCase{
			name:            "Example #2",
			chunkStreamID:   4,
			typeID:          9,
			messageStreamID: 12346,
			writeCases: []write{
				write{timestamp: 1000, length: 307},
			},
			readCases: []read{
				read{timestamp: 1000, fmt: 0},
				read{timestamp: 1000, fmt: 3},
				read{timestamp: 1000, fmt: 3, isComplete: true},
			},
		},
		// Original #1 fmt0 -> fmt3, fmt2 -> fmt3
		testCase{
			name:            "Original #1",
			chunkStreamID:   5,
			typeID:          10,
			messageStreamID: 22346,
			writeCases: []write{
				write{timestamp: 1000, length: 200},
				write{timestamp: 2000, length: 200},
			},
			readCases: []read{
				read{timestamp: 1000, fmt: 0},
				read{timestamp: 1000, fmt: 3, isComplete: true},
				read{timestamp: 1000, fmt: 2}, // timestamp delta is not updated in this time
				read{timestamp: 2000, fmt: 3, isComplete: true},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(make([]byte, 0, 2048))
			inbuf := bufio.NewReaderSize(buf, 2048)
			outbuf := bufio.NewWriterSize(buf, 2048)

			streamer := NewChunkStreamer(inbuf, outbuf, nil)

			for i, wc := range tc.writeCases {
				t.Run(fmt.Sprintf("Write: %d", i), func(t *testing.T) {
					w, err := streamer.NewChunkWriter(context.Background(), tc.chunkStreamID)
					assert.Nil(t, err)
					assert.NotNil(t, w)

					bin := make([]byte, wc.length)

					w.messageLength = uint32(len(bin))
					w.messageTypeID = tc.typeID
					w.messageStreamID = tc.messageStreamID
					w.timestamp = wc.timestamp
					w.buf.Write(bin)

					err = streamer.Sched(w)
					assert.Nil(t, err)
				})
			}

			_, err := streamer.NewChunkWriter(context.Background(), tc.chunkStreamID) // wait for writing
			assert.Nil(t, err)

			for i, rc := range tc.readCases {
				t.Run(fmt.Sprintf("Read: %d", i), func(t *testing.T) {
					isCompleted, r, err := streamer.readChunk()
					assert.Nil(t, err)
					assert.NotNil(t, r)

					assert.Equal(t, rc.fmt, r.basicHeader.fmt)
					assert.Equal(t, uint64(rc.timestamp), r.timestamp)
					assert.Equal(t, rc.isComplete, isCompleted)

					if isCompleted {
						r.Close()
					}
				})
			}
		})
	}
}

func TestWriteToInvalidWriter(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 2048))
	inbuf := bufio.NewReaderSize(buf, 2048)

	streamer := NewChunkStreamer(inbuf, &AlwaysErrorWriter{}, nil)

	// Write some data
	chunkStreamID := 10
	timestamp := uint32(0)
	err := streamer.Write(context.Background(), chunkStreamID, timestamp, &ChunkMessage{
		StreamID: 0,
		Message:  &message.Ack{},
	})
	assert.Nil(t, err)

	<-streamer.Done()
	assert.EqualErrorf(t, streamer.Err(), "Always error!", "")
}

type AlwaysErrorWriter struct{}

func (w *AlwaysErrorWriter) Write(buf []byte) (int, error) {
	return 0, fmt.Errorf("Always error!")
}

func TestChunkStreamerHasNoLeaksOfGoroutines(t *testing.T) {
	defer leaktest.Check(t)()

	buf := new(bytes.Buffer)
	inbuf := bufio.NewReaderSize(buf, 2048)
	outbuf := bufio.NewWriterSize(buf, 2048)

	streamer := NewChunkStreamer(inbuf, outbuf, nil)

	err := streamer.Close()
	assert.Nil(t, err)

	<-streamer.Done()
}

func TestChunkStreamerStreamsLimitation(t *testing.T) {
	buf := new(bytes.Buffer)
	inbuf := bufio.NewReaderSize(buf, 2048)
	outbuf := bufio.NewWriterSize(buf, 2048)

	streamer := NewChunkStreamer(inbuf, outbuf, &StreamControlStateConfig{
		MaxChunkStreams: 1,
	})
	defer streamer.Close()

	{
		_, err := streamer.prepareChunkReader(0)
		assert.Nil(t, err)

		_, err = streamer.prepareChunkReader(1)
		assert.EqualError(t, err, "Creating chunk streams limit exceeded(Reader): Limit = 1")
	}

	{
		_, err := streamer.prepareChunkWriter(0)
		assert.Nil(t, err)

		_, err = streamer.prepareChunkWriter(1)
		assert.EqualError(t, err, "Creating chunk streams limit exceeded(Writer): Limit = 1")
	}
}

func TestChunkStreamerDualWriter(t *testing.T) {
	buf := new(bytes.Buffer)
	inbuf := bufio.NewReaderSize(buf, 2048)
	outbuf := bufio.NewWriterSize(ioutil.Discard, 2048)

	streamer := NewChunkStreamer(inbuf, outbuf, nil)

	largePayload := []byte(strings.Repeat("abcdabcd12341234", 512))

	// Writes messages alternately to two streams 20 times without clogging
	const N = 20
	for i := 0; i < N; i++ {
		chunkStreamID := 10 + i%2

		// Write some data
		timestamp := uint32(0)
		err := streamer.Write(context.Background(), chunkStreamID, timestamp, &ChunkMessage{
			StreamID: 0,
			Message: &message.VideoMessage{
				Payload: bytes.NewReader(largePayload),
			},
		})
		assert.Nil(t, err)
	}

	streamer.waitWriters()

	err := streamer.Close()
	assert.Nil(t, err)

	<-streamer.Done()
	assert.Equal(t, nil, streamer.Err())
}

func TestChunkStreamerDualWriterWithoutWaiting(t *testing.T) {
	buf := new(bytes.Buffer)
	inbuf := bufio.NewReaderSize(buf, 2048)
	outbuf := bufio.NewWriterSize(ioutil.Discard, 2048)

	streamer := NewChunkStreamer(inbuf, outbuf, nil)

	largePayload := []byte(strings.Repeat("abcdabcd12341234", 512))

	// Writes messages alternately to two streams 20 times without clogging
	const N = 20
	for i := 0; i < N; i++ {
		chunkStreamID := 10 + i%2

		// Write some data
		timestamp := uint32(0)
		err := streamer.Write(context.Background(), chunkStreamID, timestamp, &ChunkMessage{
			StreamID: 0,
			Message: &message.VideoMessage{
				Payload: bytes.NewReader(largePayload),
			},
		})
		assert.Nil(t, err)
	}

	err := streamer.Close()
	assert.Nil(t, err)

	<-streamer.Done()
	assert.Equal(t, nil, streamer.Err())
}

func TestChunkStreamerNewChunkWriterTwice(t *testing.T) {
	buf := new(bytes.Buffer)
	inbuf := bufio.NewReaderSize(buf, 2048)
	outbuf := bufio.NewWriterSize(ioutil.Discard, 2048)

	streamer := NewChunkStreamer(inbuf, outbuf, nil)

	chunkStreamID := 10

	_, err := streamer.NewChunkWriter(context.Background(), chunkStreamID)
	assert.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = streamer.NewChunkWriter(ctx, chunkStreamID) // Try to acquire same chunk writer
	assert.EqualError(t, err, "Failed to wait chunk writer: context deadline exceeded")

	err = streamer.Close()
	assert.Nil(t, err)

	<-streamer.Done()
	assert.Equal(t, nil, streamer.Err())
}
