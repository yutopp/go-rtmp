//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreams(t *testing.T) {
	b := &rwcMock{}
	conn := newConn(b, &ConnConfig{
		ControlState: StreamControlStateConfig{
			MaxMessageStreams: 1,
		},
	})

	streams := newStreams(conn)

	s, err := streams.CreateIfAvailable()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), s.streamID)

	// Becomes error because number of max streams is 1
	_, err = streams.CreateIfAvailable()
	assert.NotNil(t, err)

	err = streams.Delete(s.streamID)
	assert.Nil(t, err)

	// Becomes error because the stream is already deleted
	err = streams.Delete(s.streamID)
	assert.NotNil(t, err)
}
