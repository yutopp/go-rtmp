//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yutopp/go-rtmp/message"
)

func TestConnConfig(t *testing.T) {
	b := &rwcMock{}

	conn := newConn(b, &ConnConfig{
		SkipHandshakeVerification: true,

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

	require.Equal(t, true, conn.config.SkipHandshakeVerification)

	require.Equal(t, 1234, conn.config.ReaderBufferSize)
	require.Equal(t, 1234, conn.config.WriterBufferSize)

	require.Equal(t, uint32(1234), conn.config.ControlState.DefaultChunkSize)
	require.Equal(t, uint32(1234), conn.config.ControlState.MaxChunkSize)
	require.Equal(t, 1234, conn.config.ControlState.MaxChunkStreams)

	require.Equal(t, int32(1234), conn.config.ControlState.DefaultAckWindowSize)
	require.Equal(t, int32(1234), conn.config.ControlState.MaxAckWindowSize)

	require.Equal(t, int32(1234), conn.config.ControlState.DefaultBandwidthWindowSize)
	require.Equal(t, message.LimitTypeHard, conn.config.ControlState.DefaultBandwidthLimitType)
	require.Equal(t, int32(1234), conn.config.ControlState.MaxBandwidthWindowSize)

	require.Equal(t, uint32(1234), conn.config.ControlState.MaxMessageSize)
	require.Equal(t, 1234, conn.config.ControlState.MaxMessageStreams)
}

type rwcMock struct {
	bytes.Buffer
	Closed bool
}

func (m *rwcMock) Close() error {
	m.Closed = true
	return nil
}
