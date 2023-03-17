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

	"github.com/livekit/go-rtmp/message"
)

func TestHandlerCallback(t *testing.T) {
	b := &rwcMock{}

	closer := make(chan struct{})
	handler := &testHandler{
		t:      t,
		closer: closer,
	}

	conn := newConn(b, &ConnConfig{
		Handler: handler,

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

	sconn := newServerConn(conn)
	go func() {
		<-closer
		sconn.Close()
	}()
	_ = sconn.Serve()
}

var _ Handler = (*testHandler)(nil)

type testHandler struct {
	DefaultHandler
	t      *testing.T
	closer chan struct{}
}

func (h *testHandler) OnServe(conn *Conn) {
	for _, s := range []*StreamControlState{conn.streamer.PeerState(), conn.streamer.SelfState()} {
		assert.Equal(h.t, uint32(1234), s.ChunkSize())
		assert.Equal(h.t, uint32(1234), s.AckWindowSize())
		assert.Equal(h.t, int32(1234), s.BandwidthWindowSize())
		assert.Equal(h.t, message.LimitTypeHard, s.BandwidthLimitType())
	}

	close(h.closer) // Finish testing
}
