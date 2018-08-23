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

	"github.com/yutopp/go-rtmp/message"
)

func TestConnConfig(t *testing.T) {
	b := &rwcMock{}

	conn := newConnFromIO(b, &ConnConfig{
		SkipHandshakeVerification: true,

		MaxBitrateKbps: 1234,

		ReaderBufferSize: 1234,
		WriterBufferSize: 1234,

		ControlState: StreamControlStateConfig{
			DefaultChunkSize: 1234,
			MaxChunkSize:     1234,
			MaxChunkStreams:  1234,

			DefaultAckWindowSize: 1234,
			MaxAckWindowSize:     1234,

			DefaultBandwidthWindowSize: 1234,
			DefaultBandwidthLimitType:  message.LimitTypeHard,
			MaxBandwidthWindowSize:     1234,

			MaxMessageStreams: 1234,
			MaxMessageSize:    1234,
		},
	})

	assert.Equal(t, true, conn.config.SkipHandshakeVerification)

	assert.Equal(t, uint32(1234), conn.config.MaxBitrateKbps)

	assert.Equal(t, 1234, conn.config.ReaderBufferSize)
	assert.Equal(t, 1234, conn.config.WriterBufferSize)

	assert.Equal(t, uint32(1234), conn.config.ControlState.DefaultChunkSize)
	assert.Equal(t, uint32(1234), conn.config.ControlState.MaxChunkSize)
	assert.Equal(t, 1234, conn.config.ControlState.MaxChunkStreams)

	assert.Equal(t, int32(1234), conn.config.ControlState.DefaultAckWindowSize)
	assert.Equal(t, int32(1234), conn.config.ControlState.MaxAckWindowSize)

	assert.Equal(t, int32(1234), conn.config.ControlState.DefaultBandwidthWindowSize)
	assert.Equal(t, message.LimitTypeHard, conn.config.ControlState.DefaultBandwidthLimitType)
	assert.Equal(t, int32(1234), conn.config.ControlState.MaxBandwidthWindowSize)

	assert.Equal(t, uint32(1234), conn.config.ControlState.MaxMessageSize)
	assert.Equal(t, 1234, conn.config.ControlState.MaxMessageStreams)
}

type rwcMock struct {
	bytes.Buffer
	Closed bool
}

func (m *rwcMock) Close() error {
	m.Closed = true
	return nil
}
