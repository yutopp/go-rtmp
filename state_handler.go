//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/livekit/go-rtmp/message"
)

type stateHandler interface {
	onMessage(chunkStreamID int, timestamp uint32, msg message.Message) error
	onData(chunkStreamID int, timestamp uint32, dataMsg *message.DataMessage, body interface{}) error
	onCommand(chunkStreamID int, timestamp uint32, cmdMsg *message.CommandMessage, body interface{}) error
}
