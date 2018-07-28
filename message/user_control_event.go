//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

type UserCtrlEvent interface{}

// UserCtrlEventStreamBegin (0)
type UserCtrlEventStreamBegin struct {
	StreamID uint32
}

// UserCtrlEventStreamEOF (1)
type UserCtrlEventStreamEOF struct {
	StreamID uint32
}

// UserCtrlEventStreamDry (2)
type UserCtrlEventStreamDry struct {
	StreamID uint32
}

// UserCtrlEventSetBufferLength (3)
type UserCtrlEventSetBufferLength struct {
	StreamID uint32
	LengthMs uint32
}

// UserCtrlEventStreamIsRecorded (4)
type UserCtrlEventStreamIsRecorded struct {
	StreamID uint32
}

// UserCtrlEventPingRequest (6)
type UserCtrlEventPingRequest struct {
	Timestamp uint32
}

// UserCtrlEventPingResponse (7)
type UserCtrlEventPingResponse struct {
	Timestamp uint32
}
