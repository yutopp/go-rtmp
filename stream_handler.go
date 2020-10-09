//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/edgeware/go-rtmp/internal"
	"github.com/edgeware/go-rtmp/message"
)

type streamState int

const (
	streamStateUnknown streamState = iota
	streamStateServerNotConnected
	streamStateServerConnected
	streamStateServerInactive
	streamStateServerPublish
	streamStateServerPlay
	streamStateClientNotConnected
	streamStateClientConnected
)

func (s streamState) String() string {
	switch s {
	case streamStateServerNotConnected:
		return "NotConnected(Server)"
	case streamStateServerConnected:
		return "Connected(Server)"
	case streamStateServerInactive:
		return "Inactive(Server)"
	case streamStateServerPublish:
		return "Publish(Server)"
	case streamStateServerPlay:
		return "Play(Server)"
	case streamStateClientNotConnected:
		return "NotConnected(Client)"
	case streamStateClientConnected:
		return "Connected(Client)"
	default:
		return "<Unknown>"
	}
}

// streamHandler A handler per streams.
// It holds a handler for each states and processes messages sent to the stream
type streamHandler struct {
	stream  *Stream
	handler stateHandler // A handler for each states
	state   streamState

	loggerEntry *logrus.Entry
	m           sync.Mutex
}

// newEntryHandler Create an incomplete new instance of entryHandler.
// msgHandler fields must be assigned by a caller of this function
func newStreamHandler(s *Stream) *streamHandler {
	return &streamHandler{
		stream: s,
	}
}

func (h *streamHandler) Handle(chunkStreamID int, timestamp uint32, msg message.Message) error {
	l := h.Logger()

	switch msg := msg.(type) {
	case *message.DataMessage:
		return h.handleData(chunkStreamID, timestamp, msg)

	case *message.CommandMessage:
		return h.handleCommand(chunkStreamID, timestamp, msg)

	case *message.SetChunkSize:
		l.Infof("Handle SetChunkSize: Msg = %#v", msg)
		return h.stream.streamer().PeerState().SetChunkSize(msg.ChunkSize)

	case *message.WinAckSize:
		l.Infof("Handle WinAckSize: Msg = %#v", msg)
		return h.stream.streamer().PeerState().SetAckWindowSize(msg.Size)

	default:
		err := h.handler.onMessage(chunkStreamID, timestamp, msg)
		if err == internal.ErrPassThroughMsg {
			return h.stream.userHandler().OnUnknownMessage(timestamp, msg)
		}
		return err
	}
}

func (h *streamHandler) ChangeState(state streamState) {
	h.m.Lock()
	defer h.m.Unlock()

	prevState := h.State()

	switch state {
	case streamStateUnknown:
		return // DO NOTHING
	case streamStateServerNotConnected:
		h.handler = &serverControlNotConnectedHandler{sh: h}
	case streamStateServerConnected:
		h.handler = &serverControlConnectedHandler{sh: h}
	case streamStateServerInactive:
		h.handler = &serverDataInactiveHandler{sh: h}
	case streamStateServerPublish:
		h.handler = &serverDataPublishHandler{sh: h}
	case streamStateServerPlay:
		h.handler = &serverDataPlayHandler{sh: h}
	case streamStateClientNotConnected:
		h.handler = &clientControlNotConnectedHandler{sh: h}
		// 	case streamStateClientConnected:
		// 		h.handler = &serverControlConnectedHandler{sh: h}
	default:
		panic("Unexpected")
	}
	h.state = state

	l := h.Logger()
	l.Infof("Change state: From = %s, To = %s", prevState, h.State())
}

func (h *streamHandler) State() streamState {
	return h.state
}

func (h *streamHandler) Logger() *logrus.Entry {
	if h.loggerEntry == nil {
		h.loggerEntry = h.stream.logger().WithFields(logrus.Fields{
			"stream_id": h.stream.streamID,
		})
	}

	h.loggerEntry.Data["state"] = h.State()

	return h.loggerEntry
}

func (h *streamHandler) handleData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
) error {
	bodyDecoder := message.DataBodyDecoderFor(dataMsg.Name)

	amfDec := message.NewAMFDecoder(dataMsg.Body, dataMsg.Encoding)
	var value message.AMFConvertible
	if err := bodyDecoder(dataMsg.Body, amfDec, &value); err != nil {
		return err
	}

	err := h.handler.onData(chunkStreamID, timestamp, dataMsg, value)
	if err == internal.ErrPassThroughMsg {
		return h.stream.userHandler().OnUnknownDataMessage(timestamp, dataMsg)
	}
	return err
}

func (h *streamHandler) handleCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
) error {
	switch cmdMsg.CommandName {
	case "_result", "_error":
		t, err := h.stream.transactions.At(cmdMsg.TransactionID)
		if err != nil {
			return errors.Wrap(err, "Got response to the unexpected transaction")
		}

		// Set result (NOTE: should use a mutex for it?)
		t.Reply(cmdMsg.CommandName, cmdMsg.Encoding, cmdMsg.Body)

		// Remove transacaction because this transaction is resolved
		if err := h.stream.transactions.Delete(cmdMsg.TransactionID); err != nil {
			return errors.Wrap(err, "Unexpected behavior: transaction is not found")
		}

		return nil

		// TODO: Support onStatus
	}

	amfDec := message.NewAMFDecoder(cmdMsg.Body, cmdMsg.Encoding)
	bodyDecoder := message.CmdBodyDecoderFor(cmdMsg.CommandName, cmdMsg.TransactionID)

	var value message.AMFConvertible
	if err := bodyDecoder(cmdMsg.Body, amfDec, &value); err != nil {
		return err
	}

	err := h.handler.onCommand(chunkStreamID, timestamp, cmdMsg, value)
	if err == internal.ErrPassThroughMsg {
		return h.stream.userHandler().OnUnknownCommandMessage(timestamp, cmdMsg)
	}

	return err
}
