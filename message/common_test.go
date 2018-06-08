//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

type testCase struct {
	Name string
	TypeID
	Value  Message
	Binary []byte
}

var testCases = []testCase{
	testCase{
		Name:   "AudioMessage",
		TypeID: TypeIDAudioMessage,
		Value: &AudioMessage{
			Payload: []byte("audio data"),
		},
		Binary: []byte("audio data"),
	},
	testCase{
		Name:   "VideoMessage",
		TypeID: TypeIDVideoMessage,
		Value: &VideoMessage{
			Payload: []byte("video data"),
		},
		Binary: []byte("video data"),
	},
}
