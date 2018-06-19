//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/yutopp/go-rtmp/message"
)

type streamHandler interface {
	Handle(chunkStreamID int, timestamp uint32, msg message.Message, stream *Stream) error
}

// helpers

type amfWrapperFunc func(callback func(cmd *message.CommandMessage)) message.Message

func amf0Wrapper(callback func(cmd *message.CommandMessage)) message.Message {
	var m message.CommandMessageAMF0
	callback(&m.CommandMessage)
	return &m
}

func amf3Wrapper(callback func(cmd *message.CommandMessage)) message.Message {
	var m message.CommandMessageAMF3
	callback(&m.CommandMessage)
	return &m
}
