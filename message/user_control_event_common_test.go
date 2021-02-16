//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

type uceTestCase struct {
	Name   string
	Value  UserCtrlEvent
	Binary []byte
}

var uceTestCases = []uceTestCase{
	{
		Name: "StreamBegin",
		Value: &UserCtrlEventStreamBegin{
			StreamID: 1234,
		},
		Binary: []byte{
			// ID=0
			0x00, 0x00,
			// StreamID=1234
			0x00, 0x00, 0x04, 0xd2,
		},
	},
	{
		Name: "StreamEOF",
		Value: &UserCtrlEventStreamEOF{
			StreamID: 1234,
		},
		Binary: []byte{
			// ID=1
			0x00, 0x01,
			// StreamID=1234
			0x00, 0x00, 0x04, 0xd2,
		},
	},
	{
		Name: "StreamDry",
		Value: &UserCtrlEventStreamDry{
			StreamID: 1234,
		},
		Binary: []byte{
			// ID=2
			0x00, 0x02,
			// StreamID=1234
			0x00, 0x00, 0x04, 0xd2,
		},
	},
	{
		Name: "SetBufferLength",
		Value: &UserCtrlEventSetBufferLength{
			StreamID: 1234,
			LengthMs: 5678,
		},
		Binary: []byte{
			// ID=3
			0x00, 0x03,
			// StreamID=1234
			0x00, 0x00, 0x04, 0xd2,
			// LengthMs=5678
			0x00, 0x00, 0x16, 0x2e,
		},
	},
	{
		Name: "StreamIsRecorded",
		Value: &UserCtrlEventStreamIsRecorded{
			StreamID: 1234,
		},
		Binary: []byte{
			// ID=4
			0x00, 0x04,
			// StreamID=1234
			0x00, 0x00, 0x04, 0xd2,
		},
	},
	{
		Name: "PingRequest",
		Value: &UserCtrlEventPingRequest{
			Timestamp: 1234,
		},
		Binary: []byte{
			// ID=6
			0x00, 0x06,
			// Timestamp=1234
			0x00, 0x00, 0x04, 0xd2,
		},
	},
	{
		Name: "PingResponse",
		Value: &UserCtrlEventPingResponse{
			Timestamp: 1234,
		},
		Binary: []byte{
			// ID=7
			0x00, 0x07,
			// Timestamp=1234
			0x00, 0x00, 0x04, 0xd2,
		},
	},
}
