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

// Stream represents logical stream
type Stream struct {
	streamID uint32
	handler  streamHandler
	conn     *Conn
	fragment StreamFragment
}

func (s *Stream) Write(chunkStreamID int, timestamp uint32, msg message.Message) error {
	s.fragment.Message = msg
	return s.conn.streamer.Write(chunkStreamID, timestamp, &s.fragment)
}

type StreamFragment struct {
	StreamID uint32
	Message  message.Message
}
