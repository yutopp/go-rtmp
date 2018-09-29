//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"

	"github.com/yutopp/go-rtmp/internal"
	"github.com/yutopp/go-rtmp/message"
)

type handlerState int

const (
	handlerStateServerNotConnected handlerState = iota
	handlerStateServerConnected
	handlerStateServerInactive
	handlerStateServerPublish
	handlerStateServerPlay
	handlerStateUnknown
)

func (s handlerState) String() string {
	switch s {
	case handlerStateServerNotConnected:
		return "NotConnected(Server)"
	case handlerStateServerConnected:
		return "Connected(Server)"
	case handlerStateServerInactive:
		return "Inactive(Server)"
	case handlerStateServerPublish:
		return "Publish(Server)"
	case handlerStateServerPlay:
		return "Play(Server)"
	default:
		return "<Unknown>"
	}
}

// entryHandler An entry message handler per streams.
type entryHandler struct {
	conn *Conn

	transactions *transactions
	encTy        message.EncodingType
	msgHandler   messageHandler
	m            sync.Mutex

	currentStream *Stream
	loggerEntry   *logrus.Entry
}

// newEntryHandler Create an incomplete new instance of entryHandler.
// msgHandler fields must be assigned by a caller of this function
func newEntryHandler(conn *Conn) *entryHandler {
	return &entryHandler{
		conn:         conn,
		transactions: newTransactions(),
		encTy:        message.EncodingTypeAMF0, // Defaule encoder type
	}
}

func (h *entryHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	h.currentStream = stream
	l := h.Logger()

	switch msg := msg.(type) {
	case *message.DataMessage:
		return h.handleData(chunkStreamID, timestamp, msg, stream)

	case *message.CommandMessage:
		return h.handleCommand(chunkStreamID, timestamp, msg, stream)

	case *message.SetChunkSize:
		l.Infof("Handle SetChunkSize: Msg = %#v", msg)
		return h.conn.streamer.PeerState().SetChunkSize(msg.ChunkSize)

	case *message.WinAckSize:
		l.Infof("Handle WinAckSize: Msg = %#v", msg)
		return h.conn.streamer.PeerState().SetAckWindowSize(msg.Size)

	default:
		err := h.msgHandler.Handle(chunkStreamID, timestamp, msg, stream)
		if err == internal.ErrPassThroughMsg {
			return h.conn.handler.OnUnknownMessage(timestamp, msg)
		}
		return err
	}
}

func (h *entryHandler) ChangeState(msgHandler messageHandler) {
	h.m.Lock()
	defer h.m.Unlock()

	if h.msgHandler != nil {
		l := h.Logger()
		l.Infof("Change state: From = %T, To = %T", h.msgHandler, msgHandler)
	}

	h.msgHandler = msgHandler
}

func (h *entryHandler) BorrowState(f func(messageHandler) error) error {
	h.m.Lock()
	defer h.m.Unlock()

	if h.msgHandler == nil {
		return errors.New("Nil state")
	}

	return f(h.msgHandler)
}

func (h *entryHandler) Clone() *entryHandler {
	return newEntryHandler(h.conn)
}

func (h *entryHandler) State() handlerState {
	switch h.msgHandler.(type) {
	case *serverControlNotConnectedHandler:
		return handlerStateServerNotConnected
	case *serverControlConnectedHandler:
		return handlerStateServerConnected
	case *serverDataInactiveHandler:
		return handlerStateServerInactive
	case *serverDataPublishHandler:
		return handlerStateServerPublish
	case *serverDataPlayHandler:
		return handlerStateServerPlay
	default:
		return handlerStateUnknown
	}
}

func (h *entryHandler) Logger() *logrus.Entry {
	if h.loggerEntry == nil {
		h.loggerEntry = h.conn.logger.WithFields(logrus.Fields{})
	}

	h.loggerEntry.Data["state"] = h.State()
	if h.currentStream != nil {
		h.loggerEntry.Data["stream_id"] = h.currentStream.streamID
	}

	return h.loggerEntry
}

func (h *entryHandler) handleData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	stream *Stream,
) error {
	bodyDecoder := message.DataBodyDecoderFor(dataMsg.Name)

	r := bytes.NewReader(dataMsg.Body)
	amfDec := message.NewAMFDecoder(r, dataMsg.Encoding)
	var value message.AMFConvertible
	if err := bodyDecoder(r, amfDec, &value); err != nil {
		return err
	}

	err := h.msgHandler.HandleData(chunkStreamID, timestamp, dataMsg, value, stream)
	if err == internal.ErrPassThroughMsg {
		return h.conn.handler.OnUnknownDataMessage(timestamp, dataMsg)
	}
	return err
}

func (h *entryHandler) handleCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
	stream *Stream,
) error {
	switch cmdMsg.CommandName {
	case "_result", "_error":
		t, err := h.transactions.At(cmdMsg.TransactionID)
		if err != nil {
			return errors.Wrap(err, "Got response to the unexpected transaction")
		}

		// Set result (NOTE: shoule use a mutex for t?)
		t.commandName = cmdMsg.CommandName
		t.encoding = cmdMsg.Encoding
		t.body = cmdMsg.Body
		close(t.doneCh)

		// Remove transacaction because this transaction is resolved
		if err := h.transactions.Delete(cmdMsg.TransactionID); err != nil {
			return errors.Wrap(err, "Unexpected behaviour: transaction is not found")
		}

		return nil

		// TODO: Support onStatus
	}

	r := bytes.NewReader(cmdMsg.Body)
	amfDec := message.NewAMFDecoder(r, cmdMsg.Encoding)
	bodyDecoder := message.CmdBodyDecoderFor(cmdMsg.CommandName, cmdMsg.TransactionID)

	var value message.AMFConvertible
	if err := bodyDecoder(r, amfDec, &value); err != nil {
		return err
	}

	err := h.msgHandler.HandleCommand(chunkStreamID, timestamp, cmdMsg, value, stream)
	if err == internal.ErrPassThroughMsg {
		return h.conn.handler.OnUnknownCommandMessage(timestamp, cmdMsg)
	}
	return err
}
