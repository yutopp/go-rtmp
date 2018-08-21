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

// entryHandler An entry message handler per streams.
type entryHandler struct {
	streamer *ChunkStreamer
	streams  *streams
	handler  Handler
	logger   logrus.FieldLogger

	msgHandler    messageHandler
	currentStream *Stream
	loggerEntry   *logrus.Entry
}

// newEntryHandler Create an incomplete new instance of entryHandler.
// msgHandler fields must be assigned by a caller of this function
func newEntryHandler(streamer *ChunkStreamer, streams *streams, handler Handler, logger logrus.FieldLogger) *entryHandler {
	return &entryHandler{
		streamer: streamer,
		streams:  streams,
		handler:  handler,
		logger:   logger,
	}
}

func (h *entryHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error {
	h.currentStream = stream
	l := h.Logger()

	switch msg := msg.(type) {
	case *message.CommandMessageAMF0:
		err := h.msgHandler.HandleCommand(chunkStreamID, timestamp, message.EncodingTypeAMF0, &msg.CommandMessage, stream)
		if err == internal.ErrPassThroughMsg {
			return h.handler.OnUnknownCommandMessage(timestamp, &msg.CommandMessage)
		}
		return err

	case *message.CommandMessageAMF3:
		err := h.msgHandler.HandleCommand(chunkStreamID, timestamp, message.EncodingTypeAMF3, &msg.CommandMessage, stream)
		if err == internal.ErrPassThroughMsg {
			return h.handler.OnUnknownCommandMessage(timestamp, &msg.CommandMessage)
		}
		return err

	case *message.DataMessageAMF0:
		err := h.msgHandler.HandleData(chunkStreamID, timestamp, message.EncodingTypeAMF0, &msg.DataMessage, stream)
		if err == internal.ErrPassThroughMsg {
			return h.handler.OnUnknownDataMessage(timestamp, &msg.DataMessage)
		}
		return err

	case *message.DataMessageAMF3:
		err := h.msgHandler.HandleData(chunkStreamID, timestamp, message.EncodingTypeAMF3, &msg.DataMessage, stream)
		if err == internal.ErrPassThroughMsg {
			return h.handler.OnUnknownDataMessage(timestamp, &msg.DataMessage)
		}
		return err

	case *message.SetChunkSize:
		l.Infof("Handle SetChunkSize: Msg = %#v", msg)
		return h.streamer.PeerState().SetChunkSize(msg.ChunkSize)

	case *message.WinAckSize:
		l.Infof("Handle WinAckSize: Msg = %#v", msg)
		return h.streamer.PeerState().SetAckWindowSize(msg.Size)

	default:
		err := h.msgHandler.Handle(chunkStreamID, timestamp, msg, stream)
		if err == internal.ErrPassThroughMsg {
			return h.handler.OnUnknownMessage(timestamp, msg)
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
	return newEntryHandler(
		h.streamer,
		h.streams,
		h.handler,
		h.logger,
	)
}

func (h *entryHandler) State() string {
	switch h.msgHandler.(type) {
	case *serverControlNotConnectedHandler:
		return "NotConnected(Server)"
	case *serverControlConnectedHandler:
		return "Connected(Server)"
	case *serverDataInactiveHandler:
		return "Inactive(Server)"
	case *serverDataPublishHandler:
		return "Publish(Server)"
	case *serverDataPlayHandler:
		return "Play(Server)"
	default:
		return "<Unknown>"
	}
}

func (h *entryHandler) Logger() *logrus.Entry {
	if h.loggerEntry == nil {
		h.loggerEntry = h.logger.WithFields(logrus.Fields{})
	}

	h.loggerEntry.Data["state"] = h.State()
	if h.currentStream != nil {
		h.loggerEntry.Data["stream_id"] = h.currentStream.streamID
	}

	return h.loggerEntry
}
