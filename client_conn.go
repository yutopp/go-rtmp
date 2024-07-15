//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"net"
	"sync"

	"github.com/pkg/errors"

	"github.com/yutopp/go-rtmp/handshake"
	"github.com/yutopp/go-rtmp/message"
)

// ClientConn A wrapper of a connection. It prorives client-side specific features.
type ClientConn struct {
	conn    *Conn
	lastErr error
	m       sync.RWMutex
}

func newClientConnWithSetup(c net.Conn, config *ConnConfig) (*ClientConn, error) {
	conn := newConn(c, config)

	if err := handshake.HandshakeWithServer(conn.rwc, conn.rwc, &handshake.Config{
		SkipHandshakeVerification: conn.config.SkipHandshakeVerification,
	}); err != nil {
		_ = conn.Close()
		return nil, errors.Wrap(err, "Failed to handshake")
	}

	ctrlStream, err := conn.streams.Create(ControlStreamID)
	if err != nil {
		_ = conn.Close()
		return nil, errors.Wrap(err, "Failed to create control stream")
	}
	ctrlStream.handler.ChangeState(streamStateClientNotConnected)

	conn.streamer.controlStreamWriter = ctrlStream.Write

	cc := &ClientConn{
		conn: conn,
	}
	go cc.startHandleMessageLoop()

	return cc, nil
}

func (cc *ClientConn) Close() error {
	return cc.conn.Close()
}

func (cc *ClientConn) LastError() error {
	cc.m.RLock()
	defer cc.m.RUnlock()

	return cc.lastErr
}

func (cc *ClientConn) Connect(body *message.NetConnectionConnect) error {
	if err := cc.controllable(); err != nil {
		return err
	}

	stream, err := cc.conn.streams.At(ControlStreamID)
	if err != nil {
		return err
	}

	result, err := stream.Connect(body)
	if err != nil {
		return err // TODO: wrap an error
	}

	// TODO: check result
	_ = result

	return nil
}

func (cc *ClientConn) CreateStream(body *message.NetConnectionCreateStream, chunkSize uint32) (*Stream, error) {
	if err := cc.controllable(); err != nil {
		return nil, err
	}

	stream, err := cc.conn.streams.At(ControlStreamID)
	if err != nil {
		return nil, err
	}

	result, err := stream.CreateStream(body, chunkSize)
	if err != nil {
		return nil, err // TODO: wrap an error
	}

	// TODO: check result

	newStream, err := cc.conn.streams.Create(result.StreamID)
	if err != nil {
		return nil, err
	}
	newStream.handler.ChangeState(streamStateClientConnected)

	return newStream, nil
}

func (cc *ClientConn) DeleteStream(body *message.NetStreamDeleteStream) error {
	if err := cc.controllable(); err != nil {
		return err
	}

	ctrlStream, err := cc.conn.streams.At(ControlStreamID)
	if err != nil {
		return err
	}

	// Check if stream id exists
	if _, err := cc.conn.streams.At(body.StreamID); err != nil {
		return err
	}

	if err := ctrlStream.DeleteStream(body); err != nil {
		return err
	}

	return cc.conn.streams.Delete(body.StreamID)
}

func (cc *ClientConn) startHandleMessageLoop() {
	if err := cc.conn.handleMessageLoop(); err != nil {
		cc.setLastError(err)
	}
}

func (cc *ClientConn) setLastError(err error) {
	cc.m.Lock()
	defer cc.m.Unlock()

	cc.lastErr = err
}

func (cc *ClientConn) controllable() error {
	err := cc.LastError()
	return errors.Wrap(err, "Client is in error state")
}
