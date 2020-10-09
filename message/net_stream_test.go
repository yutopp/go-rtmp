//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type netStreamTestCase struct {
	Name string
	Box  AMFConvertible

	Args        []interface{}
	ExpectedMsg AMFConvertible

	FromErr error
	ToErr   error
}

var netStreamTestCases = []netStreamTestCase{
	{
		Name: "NetStreamPublish OK",
		Box:  &NetStreamPublish{},
		Args: []interface{}{nil, "aaa", "bbb"},
		ExpectedMsg: &NetStreamPublish{
			CommandObject:  nil,
			PublishingName: "aaa",
			PublishingType: "bbb",
		},
	},
	netStreamTestCase{
		Name: "NetStreamReleaseStream OK",
		Box:  &NetStreamReleaseStream{},
		Args: []interface{}{nil, "theStream"}, // First argument is unknown
		ExpectedMsg: &NetStreamReleaseStream{
			StreamName: "theStream",
		},
	},
	netStreamTestCase{
		Name: "NetStreamFCPublish OK",
		Box:  &NetStreamFCPublish{},
		Args: []interface{}{nil, "theStream"}, // First argument is unknown
		ExpectedMsg: &NetStreamFCPublish{
			StreamName: "theStream",
		},
	},
}

func TestConvertNetStreamMessages(t *testing.T) {
	for _, tc := range netStreamTestCases {
		tc := tc // capture

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			// Make a message from args
			err := tc.Box.FromArgs(tc.Args...)
			assert.Equal(t, tc.FromErr, err)

			if err != nil {
				return
			}
			assert.Equal(t, tc.ExpectedMsg, tc.Box) // Message <- Args0

			// Make args from message
			args, err := tc.Box.ToArgs(EncodingTypeAMF0) // TODO: fix interface...
			assert.Equal(t, tc.ToErr, err)

			if err != nil {
				return
			}
			assert.Equal(t, tc.Args, args) // Args0 <- Message
		})
	}
}
