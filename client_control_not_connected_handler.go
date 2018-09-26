//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/yutopp/go-rtmp/internal"
	"github.com/yutopp/go-rtmp/message"
)

var _ messageHandler = (*clientControlNotConnectedHandler)(nil)

// clientControlNotConnectedHandler Handle control messages from a server in flow of connecting.
//   transitions:
//     | "_result" -> controlStreamStateConnected
//     | _         -> self
//
type clientControlNotConnectedHandler struct {
	entry       *entryHandler
	connectedCh chan struct{}
}

func (h *clientControlNotConnectedHandler) Handle(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}

func (h *clientControlNotConnectedHandler) HandleCommand(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	cmdMsg *message.CommandMessage,
	body interface{},
	stream *Stream,
) error {
	l := h.entry.Logger()

	switch cmd := body.(type) {
	case *message.NetConnectionConnectResult:
		l.Info("ConnectResult")
		l.Infof("Result: Info = %+v, Props = %+v", cmd.Information, cmd.Properties)

		return nil

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *clientControlNotConnectedHandler) HandleData(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	dataMsg *message.DataMessage,
	body interface{},
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}
