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
			dec := NewDecoder(buf, tc.TypeID)

			var msg Message
			err := dec.Decode(&msg)
			assert.Nil(t, err)
			assert.Equal(t, tc.Value, msg)
		})
	}
}

func BenchmarkDecodeVideoMessage(b *testing.B) {
	buf := new(bytes.Buffer)
	for i := 0; i < 1024; i++ {
		buf.WriteString("abcde")
	}
	if buf.Len() != 5*1024 {
		b.Fatalf("Buffer becomes unexpected state: Len = %d", buf.Len())
	}

	dec := NewDecoder(buf, TypeIDVideoMessage)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var msg Message
		dec.Decode(&msg)
	}
}
