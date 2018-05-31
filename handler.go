//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

type Handler interface {
	OnConnect(timestamp uint32, args []interface{}) error
	OnPublish(timestamp uint32, args []interface{}) error
	OnPlay(timestamp uint32, args []interface{}) error
	OnAudio(timestamp uint32, payload []byte) error
	OnVideo(timestamp uint32, payload []byte) error
}

type HandlerFactory func() Handler
