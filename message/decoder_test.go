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
	"testing"
)

func TestDecodeCommon(t *testing.T) {
	for _, tc := range testCases {
		tc := tc // capture

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewReader(tc.Binary)
			dec := NewDecoder(buf)

			var msg Message
			err := dec.Decode(tc.TypeID, &msg)
			assert.Nil(t, err)
			assertEqualMessage(t, tc.Value, msg)
		})
	}
}

func BenchmarkDecode5KBVideoMessage(b *testing.B) {
	sizes := []struct {
		name string
		len  int
	}{
		{"5KB", 5 * 1024},
		{"2MB", 2 * 1024 * 1024},
	}
	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			buf := make([]byte, size.len)
			r := bytes.NewReader(buf)
			dec := NewDecoder(r)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Reset(buf)

				var msg Message
				dec.Decode(TypeIDVideoMessage, &msg)
			}
		})
	}
}
