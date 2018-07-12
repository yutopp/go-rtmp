//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"io"
	"net"
	"time"
)

type Server struct {
	config *ServerConfig
}

type ServerConfig struct {
	HandlerFactory
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Conn         *ConnConfig
}

func NewServer(config *ServerConfig) *Server {
	return &Server{
		config: config,
	}
}

func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()

	for {
		rwc, err := l.Accept()
		if err != nil {
			continue
		}

		go srv.handleConn(rwc)
	}
}

func (srv *Server) handleConn(conn net.Conn) {
	c := NewConn(&rwcHasTimeout{
		conn:         conn,
		readTimeout:  srv.config.ReadTimeout,
		writeTimeout: srv.config.WriteTimeout,
		now:          time.Now,
	}, srv.config.Conn)
	defer c.Close()

	handler := srv.config.HandlerFactory(c)
	c.SetHandler(handler)

	if err := c.Serve(); err != nil {
		if err == io.EOF {
			c.logger.Infof("Server closed")
			return
		}
		c.logger.Errorf("Server closed by error: Err = %+v", err)
	}
}
