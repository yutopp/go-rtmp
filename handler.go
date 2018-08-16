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

type Handler interface {
	OnServe()
	OnConnect(timestamp uint32, cmd *message.NetConnectionConnect) error
	OnCreateStream(timestamp uint32, cmd *message.NetConnectionCreateStream) error
	OnReleaseStream(timestamp uint32, cmd *message.NetConnectionReleaseStream) error
	OnDeleteStream(timestamp uint32, cmd *message.NetStreamDeleteStream) error
	OnPublish(timestamp uint32, cmd *message.NetStreamPublish) error
	OnPlay(timestamp uint32, cmd *message.NetStreamPlay) error
	OnFCPublish(timestamp uint32, cmd *message.NetStreamFCPublish) error
	OnFCUnpublish(timestamp uint32, cmd *message.NetStreamFCUnpublish) error
	OnSetDataFrame(timestamp uint32, data *message.NetStreamSetDataFrame) error
	OnAudio(timestamp uint32, payload []byte) error
	OnVideo(timestamp uint32, payload []byte) error
	OnUnknownMessage(timestamp uint32, msg message.Message) error
	OnUnknownCommandMessage(timestamp uint32, cmd *message.CommandMessage) error
	OnUnknownDataMessage(timestamp uint32, data *message.DataMessage) error
	OnClose()
}

type HandlerFactory func(conn *Conn) Handler
