//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
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
	}
}

func (h *entryHandler) Handle(csID int, timestamp uint32, msg message.Message, stream *Stream) error {
	h.currentStream = stream
	l := h.Logger()

	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		encTy := message.EncodingTypeAMF0
		cmdMsg := &msg.CommandMessage
		return h.handleCommand(csID, timestamp, encTy, cmdMsg, stream)

	case *message.CommandMessageAMF3:
		encTy := message.EncodingTypeAMF3
		cmdMsg := &msg.CommandMessage
		return h.handleCommand(csID, timestamp, encTy, cmdMsg, stream)

	case *message.DataMessageAMF0:
		encTy := message.EncodingTypeAMF0
		dataMsg := &msg.DataMessage
		return h.handleData(csID, timestamp, encTy, dataMsg, stream)

	case *message.DataMessageAMF3:
		encTy := message.EncodingTypeAMF3
		dataMsg := &msg.DataMessage
		return h.handleData(csID, timestamp, encTy, dataMsg, stream)

	case *message.SetChunkSize:
		l.Infof("Handle SetChunkSize: Msg = %#v", msg)
		return h.conn.streamer.PeerState().SetChunkSize(msg.ChunkSize)

	case *message.WinAckSize:
		l.Infof("Handle WinAckSize: Msg = %#v", msg)
		return h.conn.streamer.PeerState().SetAckWindowSize(msg.Size)

	default:
		err := h.msgHandler.Handle(csID, timestamp, msg, stream)
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

func (h *entryHandler) handleCommand(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	cmdMsg *message.CommandMessage,
	stream *Stream,
) error {
	var callback func(v interface{}, err error)
	switch cmdMsg.CommandName {
	case "_result":
		t, err := h.transactions.At(cmdMsg.TransactionID)
		if err != nil {
			return errors.Wrap(err, "Got response to the unexpected transaction")
		}
		// Remove transacaction because this transaction is resolved
		if err := h.transactions.Delete(cmdMsg.TransactionID); err != nil {
			return errors.Wrap(err, "Unexpected behaviour: transaction is not found")
		}

		cmdMsg.Decoder.MsgDecoder = t.decoder

	default:
		// TODO: support onStatus
		cmdMsg.Decoder.MsgDecoder = message.CmdBodyDecoderFor(cmdMsg.CommandName, cmdMsg.TransactionID)
	}
	if err := cmdMsg.Decoder.Decode(); err != nil {
		if callback != nil {
			callback(nil, err)
		}
		return err
	}

	if callback != nil {
		callback(cmdMsg.Decoder.Value, nil)
	}

	err := h.msgHandler.HandleCommand(chunkStreamID, timestamp, encTy, cmdMsg, cmdMsg.Decoder.Value, stream)
	if err == internal.ErrPassThroughMsg {
		return h.conn.handler.OnUnknownCommandMessage(timestamp, cmdMsg)
	}
	return err
}

func (h *entryHandler) handleData(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	dataMsg *message.DataMessage,
	stream *Stream,
) error {
	dataMsg.Decoder.MsgDecoder = message.DataBodyDecoderFor(dataMsg.Name)
	if err := dataMsg.Decoder.Decode(); err != nil {
		return err
	}

	err := h.msgHandler.HandleData(chunkStreamID, timestamp, encTy, dataMsg, dataMsg.Decoder.Value, stream)
	if err == internal.ErrPassThroughMsg {
		return h.conn.handler.OnUnknownDataMessage(timestamp, dataMsg)
	}
	return err
}
