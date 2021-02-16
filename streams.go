//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"sync"

	"github.com/pkg/errors"
)

// ControlStreamID StreamID 0 is a control stream
const ControlStreamID = 0

type streams struct {
	streams map[uint32]*Stream
	m       sync.Mutex

	conn *Conn
}

func newStreams(conn *Conn) *streams {
	return &streams{
		streams: make(map[uint32]*Stream),

		conn: conn,
	}
}

func (ss *streams) Create(streamID uint32) (*Stream, error) {
	ss.m.Lock()
	defer ss.m.Unlock()

	_, ok := ss.streams[streamID]
	if ok {
		return nil, errors.Errorf("Stream already exists: StreamID = %d", streamID)
	}
	if len(ss.streams) >= ss.conn.config.ControlState.MaxMessageStreams {
		return nil, errors.Errorf(
			"Creating message streams limit exceeded: Limit = %d",
			ss.conn.config.ControlState.MaxMessageStreams,
		)
	}

	ss.streams[streamID] = newStream(streamID, ss.conn)

	return ss.streams[streamID], nil
}

func (ss *streams) CreateIfAvailable() (*Stream, error) {
	for i := 0; i < ss.conn.config.ControlState.MaxMessageStreams; i++ {
		s, err := ss.Create(uint32(i))
		if err != nil {
			continue
		}
		return s, nil
	}

	return nil, errors.Errorf(
		"Creating streams limit exceeded: Limit = %d",
		ss.conn.config.ControlState.MaxMessageStreams,
	)
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
