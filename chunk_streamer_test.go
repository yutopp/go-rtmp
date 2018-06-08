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
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"github.com/yutopp/go-rtmp/message"
)

func TestStreamerSingleChunk(t *testing.T) {
	buf := new(bytes.Buffer)
	inbuf := bufio.NewReaderSize(buf, 2048)
	outbuf := bufio.NewWriterSize(buf, 2048)

	streamer := NewChunkStreamer(inbuf, outbuf)

	chunkStreamID := 2
	msg := &message.VideoMessage{
		Payload: []byte("testtesttest"),
	}
	timestamp := uint32(72)

	// write a message
	w, err := streamer.NewChunkWriter(chunkStreamID)
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

	_, err = streamer.NewChunkWriter(chunkStreamID) // wait for writing
	assert.Nil(t, err)

	// read a message
	r, err := streamer.NewChunkReader()
	assert.Nil(t, err)
	assert.NotNil(t, r)
	defer r.Close()

	dec := message.NewDecoder(r, message.TypeID(r.messageTypeID))
	var actualMsg message.Message
	err = dec.Decode(&actualMsg)
	assert.Nil(t, err)

	// check message
	assert.Equal(t, msg, actualMsg)
}

func TestStreamerMultipleChunk(t *testing.T) {
	buf := new(bytes.Buffer)
	buf.Grow(2048)
	inbuf := bufio.NewReaderSize(buf, 2048)
	outbuf := bufio.NewWriterSize(buf, 2048)

	streamer := NewChunkStreamer(inbuf, outbuf)

	chunkStreamID := 2
	msg := &message.VideoMessage{
		Payload: []byte(strings.Repeat("test", 128)),
	}
	timestamp := uint32(72)

	// write a message
	w, err := streamer.NewChunkWriter(chunkStreamID)
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

	_, err = streamer.NewChunkWriter(chunkStreamID) // wait for writing
	assert.Nil(t, err)

	// read a message
	r, err := streamer.NewChunkReader()
	assert.Nil(t, err)
	assert.NotNil(t, r)
	defer r.Close()

	dec := message.NewDecoder(r, message.TypeID(r.messageTypeID))
	var actualMsg message.Message
	err = dec.Decode(&actualMsg)
	assert.Nil(t, err)

	// check message
	assert.Equal(t, msg, actualMsg)
}
