//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/pkg/errors"
	"sync"
)

// ControlStreamID StreamID 0 is a control stream
const ControlStreamID = 0

type streams struct {
	streamer *ChunkStreamer
	streams  map[uint32]*Stream
	m        sync.Mutex

	config *StreamControlStateConfig
}

func newStreams(streamer *ChunkStreamer, config *StreamControlStateConfig) *streams {
	return &streams{
		streamer: streamer,
		streams:  make(map[uint32]*Stream),
		config:   config,
	}
}

func (ss *streams) Create(streamID uint32, entryHandler *entryHandler) (*Stream, error) {
	ss.m.Lock()
	defer ss.m.Unlock()

	_, ok := ss.streams[streamID]
	if ok {
		return nil, errors.Errorf("Stream already exists: StreamID = %d", streamID)
	}
	if len(ss.streams) >= ss.config.MaxMessageStreams {
		return nil, errors.Errorf(
			"Creating message streams limit exceeded: Limit = %d",
			ss.config.MaxMessageStreams,
		)
	}

	ss.streams[streamID] = newStream(streamID, entryHandler, ss.streamer)

	return ss.streams[streamID], nil
}

func (ss *streams) CreateIfAvailable(entryHandler *entryHandler) (uint32, error) {
	for i := 0; i < ss.config.MaxMessageStreams; i++ {
		s, err := ss.Create(uint32(i), entryHandler)
		if err != nil {
			continue
		}
		return s.streamID, nil
	}

	return 0, errors.Errorf("Creating streams limit exceeded: Limit = %d", ss.config.MaxMessageStreams)
}

func (ss *streams) Delete(streamID uint32) error {
	ss.m.Lock()
	defer ss.m.Unlock()

	_, ok := ss.streams[streamID]
	if !ok {
		return errors.Errorf("Stream not exists: StreamID = %d", streamID)
	}

	delete(ss.streams, streamID)

	return nil
}

func (ss *streams) At(streamID uint32) (*Stream, error) {
	stream, ok := ss.streams[streamID]
	if !ok {
		return nil, errors.Errorf("Stream is not found: StreamID = %d", streamID)
	}

	return stream, nil
}
