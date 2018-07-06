//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnStreams(t *testing.T) {
	b := &rwcMock{}

	conn := NewConn(b, &ConnConfig{
		MaxStreams: 1,
	})

	sid, err := conn.createStreamIfAvailable(nil)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), sid)

	// Becomes error because number of max streams is 1
	_, err = conn.createStreamIfAvailable(nil)
	assert.NotNil(t, err)

	err = conn.deleteStream(sid)
	assert.Nil(t, err)

	// Becomes error because the stream is already deleted
	err = conn.deleteStream(sid)
	assert.NotNil(t, err)
}

type rwcMock struct {
	bytes.Buffer
	Closed bool
}

func (m *rwcMock) Close() error {
	m.Closed = true
	return nil
}
