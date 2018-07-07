//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestDecodeCommon(t *testing.T) {
	for _, tc := range testCases {
		tc := tc // capture

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			bin := make([]byte, len(tc.Binary))
			copy(bin, tc.Binary) // copy ownership

			buf := bytes.NewBuffer(bin)
			dec := NewDecoder(buf, tc.TypeID)
			dec.amfMessageParser = func(r io.Reader, d AMFDecoder, name string, v *AMFConvertible) error {
				return mockedParseAMFMessage(t, r, d, name, v)
			}

			var msg Message
			err := dec.Decode(&msg)
			assert.Nil(t, err)
			assert.Equal(t, tc.Value, msg)
		})
	}
}

func mockedParseAMFMessage(t *testing.T, r io.Reader, d AMFDecoder, name string, v *AMFConvertible) error {
	t.Logf("mockmock: %s", name)
	return nil
}
