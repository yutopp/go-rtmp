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

	"github.com/yutopp/go-rtmp/message"
)

func TestHandlerCallback(t *testing.T) {
	b := &rwcMock{}

	conn := NewConn(b, &ConnConfig{
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

	closer := make(chan struct{})
	handler := &testHandler{
		t:      t,
		conn:   conn,
		closer: closer,
	}
	conn.SetHandler(handler)

	go func() {
		select {
		case <-closer:
			conn.Close()
		}
	}()
	conn.Serve()
}

var _ Handler = (*testHandler)(nil)

type testHandler struct {
	t      *testing.T
	conn   *Conn
	closer chan struct{}
}

func (h *testHandler) OnInit(conn *Conn) {
	for _, s := range []*StreamControlState{h.conn.streamer.PeerState(), h.conn.streamer.SelfState()} {
		assert.Equal(h.t, uint32(1234), s.ChunkSize())
		assert.Equal(h.t, uint32(1234), s.AckWindowSize())
		assert.Equal(h.t, int32(1234), s.BandwidthWindowSize())
		assert.Equal(h.t, message.LimitTypeHard, s.BandwidthLimitType())
	}

	close(h.closer) // Finish testing
}

func (h *testHandler) OnCommand(timestamp uint32, cmd Command) error {
	return nil
}

func (h *testHandler) OnData(timestamp uint32, cmd Data) error {
	return nil
}

func (h *testHandler) OnAudio(timestamp uint32, payload []byte) error {
	return nil
}

func (h *testHandler) OnVideo(timestamp uint32, payload []byte) error {
	return nil
}

func (h *testHandler) OnClose() {
}
