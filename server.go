//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"log"
	"net"
)

type Server struct {
	config *ServerConfig
}

type ServerConfig struct {
	HandlerFactory
	Conn *ConnConfig
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

		c := srv.newConn(rwc, srv.config.HandlerFactory(), srv.config.Conn)
		go func() {
			// TODO: fix
			if err := c.Serve(); err != nil {
				log.Printf("Serve error: Err = %+v", err)
			}
		}()
	}
}

func (srv *Server) newConn(rwc net.Conn, handler Handler, config *ConnConfig) *Conn {
	conn := NewConn(rwc, handler, config)

	return conn
}
