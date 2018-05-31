//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"net"
)

type Server struct {
}

func (srv *Server) Serve(l net.Listener, handherFactory HandlerFactory) error {
	defer l.Close()

	for {
		rwc, err := l.Accept()
		if err != nil {
			continue
		}

		c := srv.newConn(rwc, handherFactory())
		go c.Serve()
	}
}

func (srv *Server) newConn(rwc net.Conn, handler Handler) *Conn {
	conn := NewConn(rwc, handler)

	return conn
}
