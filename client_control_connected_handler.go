//
// Copyright (c) 2023- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/yutopp/go-rtmp/internal"
	"github.com/yutopp/go-rtmp/message"
)

// clientControlConnectedHandler Handle control messages from a server in flow of connected.
//
//	transitions:
//	  | _         -> self
type clientControlConnectedHandler struct {
	sh *streamHandler
}

var _ stateHandler = (*clientControlConnectedHandler)(nil)

func (h *clientControlConnectedHandler) onMessage(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
) error {
	return internal.ErrPassThroughMsg
}

func (h *clientControlConnectedHandler) onData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	body interface{},
) error {
	return internal.ErrPassThroughMsg
}

func (h *clientControlConnectedHandler) onCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
	body interface{},
) error {
	switch cmd := body.(type) {
	case *message.NetStreamOnStatus:
		if cmd.InfoObject.Code == message.NetStreamOnStatusCodePlayStart {
			h.sh.ChangeState(streamStateClientPlay)
		}

		return internal.ErrPassThroughMsg

	default:
		return internal.ErrPassThroughMsg
	}
}
