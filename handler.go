//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"io"

	"github.com/yutopp/go-rtmp/message"
)

type Handler interface {
	OnServe(conn *Conn)
	OnConnect(timestamp uint32, cmd *message.NetConnectionConnect) error
	OnCreateStream(timestamp uint32, cmd *message.NetConnectionCreateStream) error
	OnReleaseStream(timestamp uint32, cmd *message.NetConnectionReleaseStream) error
	OnDeleteStream(timestamp uint32, cmd *message.NetStreamDeleteStream) error
	OnPublish(ctx *StreamContext, timestamp uint32, cmd *message.NetStreamPublish) error
	OnPlay(ctx *StreamContext, timestamp uint32, cmd *message.NetStreamPlay) error
	OnFCPublish(timestamp uint32, cmd *message.NetStreamFCPublish) error
	OnFCUnpublish(timestamp uint32, cmd *message.NetStreamFCUnpublish) error
	OnSetDataFrame(timestamp uint32, data *message.NetStreamSetDataFrame) error
	OnAudio(timestamp uint32, payload io.Reader) error
	OnVideo(timestamp uint32, payload io.Reader) error
	OnUnknownMessage(timestamp uint32, msg message.Message) error
	OnUnknownCommandMessage(timestamp uint32, cmd *message.CommandMessage) error
	OnUnknownDataMessage(timestamp uint32, data *message.DataMessage) error
	OnClose()
	OnError(args ...interface{})
}
