//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/pkg/errors"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	config *ServerConfig

	listener net.Listener
	mu       sync.Mutex
	doneCh   chan struct{}
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
	if err := srv.registerListener(l); err != nil {
		return errors.Wrap(err, "Already serverd")
	}

	defer l.Close()

	for {
		rwc, err := l.Accept()
		if err != nil {
			select {
			case <-srv.getDoneCh(): // closed
				return ErrClosed

			default: // do nothing
			}

			continue
		}

		go srv.handleConn(rwc)
	}
}

func (srv *Server) Close() error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	doneCh := srv.getDoneChLocked()
	select {
	case <-doneCh: // already closed
		return nil
	default:
		close(doneCh)
	}

	if srv.listener == nil {
		return nil
	}

	return srv.listener.Close()
}

func (srv *Server) registerListener(l net.Listener) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if srv.listener != nil {
		return errors.New("Listener is already registered")
	}

	srv.listener = l

	return nil
}

func (srv *Server) getDoneCh() chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	return srv.getDoneChLocked()
}

func (srv *Server) getDoneChLocked() chan struct{} {
	if srv.doneCh == nil {
		srv.doneCh = make(chan struct{})
	}

	return srv.doneCh
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
