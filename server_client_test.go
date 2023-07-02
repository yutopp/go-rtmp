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
	"github.com/stretchr/testify/require"
	"github.com/yutopp/go-amf0"

	"github.com/yutopp/go-rtmp/message"
)

const (
	chunkSize = 128
)

type serverCanAcceptConnectHandler struct {
	DefaultHandler
}

func TestServerCanAcceptConnect(t *testing.T) {
	config := &ConnConfig{
		Handler: &serverCanAcceptConnectHandler{},
		Logger:  logrus.StandardLogger(),
	}

	prepareConnection(t, config, func(c *ClientConn) {
		err := c.Connect(nil)
		require.Nil(t, err)
	})
}

type serverCanRejectConnectHandler struct {
	DefaultHandler
}

func (h *serverCanRejectConnectHandler) OnConnect(_ uint32, _ *message.NetConnectionConnect) error {
	return fmt.Errorf("Reject")
}

func TestServerCanRejectConnect(t *testing.T) {
	config := &ConnConfig{
		Handler: &serverCanRejectConnectHandler{},
		Logger:  logrus.StandardLogger(),
	}

	prepareConnection(t, config, func(c *ClientConn) {
		err := c.Connect(nil)
		require.Equal(t, &ConnectRejectedError{
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

type serverCanAcceptCreateStreamHandler struct {
	DefaultHandler
}

func TestServerCanAcceptCreateStream(t *testing.T) {
	config := &ConnConfig{
		Handler: &serverCanAcceptCreateStreamHandler{},
		Logger:  logrus.StandardLogger(),
		ControlState: StreamControlStateConfig{
			MaxMessageStreams: 2, // Control and another 1 stream
		},
	}

	prepareConnection(t, config, func(c *ClientConn) {
		err := c.Connect(nil)
		require.Nil(t, err)

		s0, err := c.CreateStream(nil, chunkSize)
		require.Nil(t, err)
		defer s0.Close()

		// Rejected because a number of message streams is exceeded the limits
		s1, err := c.CreateStream(nil, chunkSize)
		require.Equal(t, &CreateStreamRejectedError{
			TransactionID: 2,
			Result: &message.NetConnectionCreateStreamResult{
				StreamID: 0,
			},
		}, err)
		defer s1.Close()
	})
}

type serverCanAcceptDeleteStreamHandler struct {
	DefaultHandler
}

func TestServerCanAcceptDeleteStream(t *testing.T) {
	config := &ConnConfig{
		Handler: &serverCanAcceptDeleteStreamHandler{},
		Logger:  logrus.StandardLogger(),
		ControlState: StreamControlStateConfig{
			MaxMessageStreams: 2, // Control and another 1 stream
		},
	}

	prepareConnection(t, config, func(c *ClientConn) {
		err := c.Connect(nil)
		require.Nil(t, err)

		s0, err := c.CreateStream(nil, chunkSize)
		require.NoError(t, err)
		defer s0.Close()

		t.Run("Cannot delete a stream which does not exist", func(t *testing.T) {
			err = c.DeleteStream(&message.NetStreamDeleteStream{
				StreamID: 42,
			})
			require.Error(t, err)
		})

		t.Run("Can delete a stream", func(t *testing.T) {
			err = c.DeleteStream(&message.NetStreamDeleteStream{
				StreamID: s0.streamID,
			})
			require.NoError(t, err)
		})
	})
}

func prepareConnection(t *testing.T, config *ConnConfig, f func(c *ClientConn)) {
	// prepare server
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.Nil(t, err)

	srv := NewServer(&ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *ConnConfig) {
			return conn, config
		},
	})
	defer func() {
		err := srv.Close()
		require.Nil(t, err)
	}()

	go func() {
		err := srv.Serve(l)
		require.Equal(t, ErrClosed, err)
	}()

	// prepare client
	c, err := Dial("rtmp", l.Addr().String(), &ConnConfig{
		Logger: logrus.StandardLogger(),
	})
	require.Nil(t, err)
	defer func() {
		err := c.Close()
		require.Nil(t, err)
	}()

	f(c)
}
