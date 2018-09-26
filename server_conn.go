//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/yutopp/go-rtmp/handshake"
)

// serverConn A wrapper of a connection. It prorives server-side specific features.
type serverConn struct {
	conn *Conn
}

func newServerConn(conn *Conn) *serverConn {
	return &serverConn{
		conn: conn,
	}
}

func (sc *serverConn) Serve() error {
	if err := handshake.HandshakeWithClient(sc.conn.rwc, sc.conn.rwc, &handshake.Config{
		SkipHandshakeVerification: sc.conn.config.SkipHandshakeVerification,
	}); err != nil {
		return err
	}

	eh := newEntryHandler(sc.conn)
	eh.ChangeState(&serverControlNotConnectedHandler{entry: eh})
	if err := sc.conn.streams.Create(ControlStreamID, eh); err != nil {
		return err
	}

	defaultStream, err := sc.conn.streams.At(ControlStreamID)
	if err != nil {
		return err
	}
	sc.conn.streamer.controlStreamWriter = defaultStream.write

	if sc.conn.handler != nil {
		sc.conn.handler.OnServe()
	}

	return sc.conn.handleMessageLoop()
}

func (sc *serverConn) Close() error {
	return sc.conn.Close()
}
