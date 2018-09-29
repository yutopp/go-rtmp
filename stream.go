//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/pkg/errors"
	"github.com/yutopp/go-rtmp/message"
)

// Stream represents a logical message stream
type Stream struct {
	streamID     uint32
	entryHandler *entryHandler
	streamer     *ChunkStreamer
	cmsg         ChunkMessage
}

func (s *Stream) WriteWinAckSize(chunkStreamID int, timestamp uint32, msg *message.WinAckSize) error {
	return s.write(chunkStreamID, timestamp, msg)
}

func (s *Stream) WriteSetPeerBandwidth(chunkStreamID int, timestamp uint32, msg *message.SetPeerBandwidth) error {
	return s.write(chunkStreamID, timestamp, msg)
}

func (s *Stream) WriteUserCtrl(chunkStreamID int, timestamp uint32, msg *message.UserCtrl) error {
	return s.write(chunkStreamID, timestamp, msg)
}

func (s *Stream) WriteCommandMessage(chunkStreamID int, timestamp uint32, amf message.EncodingType, m *message.CommandMessage) error {
	var msg message.Message
	switch amf {
	case message.EncodingTypeAMF0:
		msg = &message.CommandMessageAMF0{
			CommandMessage: *m,
		}
	default:
		return errors.Errorf("Unsupported amf type: %+v", amf)
	}

	return s.write(chunkStreamID, timestamp, msg)
}

func (s *Stream) write(chunkStreamID int, timestamp uint32, msg message.Message) error {
	s.cmsg.Message = msg
	return s.streamer.Write(chunkStreamID, timestamp, &s.cmsg)
}

func (s *Stream) handle(chunkStreamID int, timestamp uint32, msg message.Message) error {
	return s.entryHandler.Handle(chunkStreamID, timestamp, msg, s)
}
