//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/yutopp/go-amf0"

	"github.com/livekit/go-rtmp/message"
)

const (
	chunkSize = 128
)

func TestServerCanAcceptConnect(t *testing.T) {
	config := &ConnConfig{
		Handler: &ServerCanAcceptConnectHandler{},
		Logger:  logrus.StandardLogger(),
	}

	prepareConnection(t, config, func(c *ClientConn) {
		err := c.Connect(nil)
		assert.Nil(t, err)
	})
}

type ServerCanAcceptConnectHandler struct {
	DefaultHandler
}

func TestServerCanRejectConnect(t *testing.T) {
	config := &ConnConfig{
		Handler: &ServerCanRejectConnectHandler{},
		Logger:  logrus.StandardLogger(),
	}

	prepareConnection(t, config, func(c *ClientConn) {
		err := c.Connect(nil)
		assert.Equal(t, &ConnectRejectedError{
			TransactionID: 1,
			Result: &message.NetConnectionConnectResult{
				Properties: message.NetConnectionConnectResultProperties{
					FMSVer:       "GO-RTMP/0,0,0,0",
					Capabilities: 31,
					Mode:         1,
				},
				Information: message.NetConnectionConnectResultInformation{
					Level:       "error",
					Code:        "NetConnection.Connect.Failed",
					Description: "Connection failed.",
					Data:        amf0.ECMAArray{"type": "go-rtmp", "version": "master"},
				},
			},
		}, err)
	})
}

type ServerCanRejectConnectHandler struct {
	DefaultHandler
}

func (h *ServerCanRejectConnectHandler) OnConnect(_ uint32, _ *message.NetConnectionConnect) error {
	return fmt.Errorf("Reject")
}

func TestServerCanAcceptCreateStream(t *testing.T) {
	config := &ConnConfig{
		Handler: &ServerCanAcceptCreateStreamHandler{},
		Logger:  logrus.StandardLogger(),
		ControlState: StreamControlStateConfig{
			MaxMessageStreams: 2, // Control and another 1 stream
		},
	}

	prepareConnection(t, config, func(c *ClientConn) {
		err := c.Connect(nil)
		assert.Nil(t, err)

		s0, err := c.CreateStream(nil, chunkSize)
		assert.Nil(t, err)
		defer s0.Close()

		// Rejected because a number of message streams is exceeded the limits
		s1, err := c.CreateStream(nil, chunkSize)
		assert.Equal(t, &CreateStreamRejectedError{
			TransactionID: 2,
			Result: &message.NetConnectionCreateStreamResult{
				StreamID: 0,
			},
		}, err)
		defer s1.Close()
	})
}

type ServerCanAcceptCreateStreamHandler struct {
	DefaultHandler
}

func prepareConnection(t *testing.T, config *ConnConfig, f func(c *ClientConn)) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	assert.Nil(t, err)

	srv := NewServer(&ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *ConnConfig) {
			return conn, config
		},
	})
	defer func() {
		err := srv.Close()
		assert.Nil(t, err)
	}()

	go func() {
		err := srv.Serve(l)
		assert.Equal(t, ErrClosed, err)
	}()

	c, err := Dial("rtmp", l.Addr().String(), &ConnConfig{
		Logger: logrus.StandardLogger(),
	})
	assert.Nil(t, err)
	defer func() {
		err := c.Close()
		assert.Nil(t, err)
	}()

	f(c)
}
