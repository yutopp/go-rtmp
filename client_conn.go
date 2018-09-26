//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/pkg/errors"
	"net"
	"sync"

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
		return nil, errors.Wrap(err, "Failed to handshake")
	}

	eh := newEntryHandler(conn)
	eh.ChangeState(&clientControlNotConnectedHandler{
		entry:       eh,
		connectedCh: make(chan struct{}),
	})
	if err := conn.streams.Create(ControlStreamID, eh); err != nil {
		return nil, errors.Wrap(err, "Failed to create control stream")
	}

	defaultStream, err := conn.streams.At(ControlStreamID)
	if err != nil {
		return nil, err
	}
	conn.streamer.controlStreamWriter = defaultStream.write

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

func (cc *ClientConn) Connect() error {
	if err := cc.controllable(); err != nil {
		return err
	}

	stream, err := cc.conn.streams.At(ControlStreamID)
	if err != nil {
		return err
	}

	connectedCh := make(chan *message.NetConnectionConnectResult)
	errCh := make(chan error)
	transactionID := int64(1) // Always 1 (7.2.1.1)
	err = stream.entryHandler.transactions.Create(transactionID, transaction{
		decoder: message.DecodeBodyConnectResult,
		callback: func(v interface{}, err error) {
			if err != nil {
				errCh <- err
				return
			}
			connectedCh <- v.(*message.NetConnectionConnectResult)
		},
	})
	if err != nil {
		return err
	}

	bodyEnc := &message.BodyEncoder{
		Value: &message.NetConnectionConnect{
			Command: message.NetConnectionConnectCommand{},
		},
		MsgEncoder: message.EncodeBodyAnyValues,
	}
	cmdMsg := &message.CommandMessage{
		CommandName:   "connect",
		TransactionID: transactionID,
		Encoder:       bodyEnc,
	}

	chunkStreamID := 3 // TODO: fix
	err = stream.WriteCommandMessage(
		chunkStreamID,
		0, // Timestamp is 0
		message.EncodingTypeAMF0,
		cmdMsg,
	)
	if err != nil {
		return err
	}

	// TODO: support timeout
	// TODO: check result
	select {
	case <-connectedCh:
	}

	return nil
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
