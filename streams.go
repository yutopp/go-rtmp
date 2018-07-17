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

type streams struct {
	conn *Conn

	streams map[uint32]*Stream
	m       sync.Mutex

	config *StreamControlStateConfig
}

func newStreams(conn *Conn, config *StreamControlStateConfig) *streams {
	return &streams{
		conn:    conn,
		streams: make(map[uint32]*Stream),
		config:  config,
	}
}

func (ss *streams) At(streamID uint32) (*Stream, bool) {
	stream, ok := ss.streams[streamID]
	return stream, ok
}

func (ss *streams) Create(streamID uint32, handler streamHandler) error {
	ss.m.Lock()
	defer ss.m.Unlock()

	_, ok := ss.streams[streamID]
	if ok {
		return errors.Errorf("Stream already exists: StreamID = %d", streamID)
	}
	if len(ss.streams) >= ss.config.MaxMessageStreams {
		return errors.Errorf(
			"Creating message streams limit exceeded: Limit = %d",
			ss.config.MaxMessageStreams,
		)
	}

	ss.streams[streamID] = &Stream{
		streamID: streamID,
		handler:  handler,
		conn:     ss.conn,
		fragment: StreamFragment{
			StreamID: streamID,
		},
	}

	return nil
}

func (ss *streams) CreateIfAvailable(handler streamHandler) (uint32, error) {
	for i := 0; i < ss.config.MaxMessageStreams; i++ {
		if err := ss.Create(uint32(i), handler); err != nil {
			continue
		}
		return uint32(i), nil
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
