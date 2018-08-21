//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStreams(t *testing.T) {
	b := &rwcMock{}

	streamer := NewChunkStreamer(b, b, nil)
	streams := newStreams(streamer, &StreamControlStateConfig{
		MaxMessageStreams: 1,
	})

	sid, err := streams.CreateIfAvailable(nil)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), sid)

	// Becomes error because number of max streams is 1
	_, err = streams.CreateIfAvailable(nil)
	assert.NotNil(t, err)

	err = streams.Delete(sid)
	assert.Nil(t, err)

	// Becomes error because the stream is already deleted
	err = streams.Delete(sid)
	assert.NotNil(t, err)
}
