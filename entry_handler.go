//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/sirupsen/logrus"

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

	msgHandler    messageHandler
	currentStream *Stream
	loggerEntry   *logrus.Entry
}

// newEntryHandler Create an incomplete new instance of entryHandler.
// msgHandler fields must be assigned by a caller of this function
func newEntryHandler(conn *Conn) *entryHandler {
	return &entryHandler{
		conn: conn,
	}
}

func (h *entryHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	h.currentStream = stream
	l := h.Logger()

	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		encTy := message.EncodingTypeAMF0
		err := h.msgHandler.HandleCommand(chunkStreamID, timestamp, encTy, &msg.CommandMessage, stream)
		if err == internal.ErrPassThroughMsg {
			return h.conn.handler.OnUnknownCommandMessage(timestamp, &msg.CommandMessage)
		}
		return err

	case *message.CommandMessageAMF3:
		encTy := message.EncodingTypeAMF3
		err := h.msgHandler.HandleCommand(chunkStreamID, timestamp, encTy, &msg.CommandMessage, stream)
		if err == internal.ErrPassThroughMsg {
			return h.conn.handler.OnUnknownCommandMessage(timestamp, &msg.CommandMessage)
		}
		return err

	case *message.DataMessageAMF0:
		encTy := message.EncodingTypeAMF0
		err := h.msgHandler.HandleData(chunkStreamID, timestamp, encTy, &msg.DataMessage, stream)
		if err == internal.ErrPassThroughMsg {
			return h.conn.handler.OnUnknownDataMessage(timestamp, &msg.DataMessage)
		}
		return err

	case *message.DataMessageAMF3:
		encTy := message.EncodingTypeAMF3
		err := h.msgHandler.HandleData(chunkStreamID, timestamp, encTy, &msg.DataMessage, stream)
		if err == internal.ErrPassThroughMsg {
			return h.conn.handler.OnUnknownDataMessage(timestamp, &msg.DataMessage)
		}
		return err

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
	if h.msgHandler != nil {
		l := h.Logger()
		l.Infof("Change state: From = %T, To = %T", h.msgHandler, msgHandler)
	}

	h.msgHandler = msgHandler
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
