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

func TestDecodeCommonDataMsg(t *testing.T) {
	for _, tc := range dataMsgTestCases {
		tc := tc // capture

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewReader(tc.Binary)
			dec := NewDecoder(buf, tc.TypeID)

			var msg Message
			err := dec.Decode(&msg)
			assert.Nil(t, err)

			var dataMsg *DataMessage
			switch m := msg.(type) {
			case *DataMessageAMF3:
				dataMsg = &m.DataMessage
			case *DataMessageAMF0:
				dataMsg = &m.DataMessage
			default:
				assert.Fail(t, "Unexpected msg", m)
			}

			var tcDataMsg *DataMessage
			switch m := tc.Value.(type) {
			case *DataMessageAMF3:
				tcDataMsg = &m.DataMessage
			case *DataMessageAMF0:
				tcDataMsg = &m.DataMessage
			default:
				assert.Fail(t, "Unexpected tc msg", m)
			}

			assert.Equal(t, tcDataMsg.Name, dataMsg.Name)
		})
	}
}

func TestDecodeCommonCmdMsg(t *testing.T) {
	for _, tc := range cmdMsgTestCases {
		tc := tc // capture

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewReader(tc.Binary)
			dec := NewDecoder(buf, tc.TypeID)

			var msg Message
			err := dec.Decode(&msg)
			assert.Nil(t, err)

			var cmdMsg *CommandMessage
			switch m := msg.(type) {
			case *CommandMessageAMF3:
				cmdMsg = &m.CommandMessage
			case *CommandMessageAMF0:
				cmdMsg = &m.CommandMessage
			default:
				assert.Fail(t, "Unexpected msg", m)
			}

			var tcCmdMsg *CommandMessage
			switch m := tc.Value.(type) {
			case *CommandMessageAMF3:
				tcCmdMsg = &m.CommandMessage
			case *CommandMessageAMF0:
				tcCmdMsg = &m.CommandMessage
			default:
				assert.Fail(t, "Unexpected tc msg", m)
			}

			assert.Equal(t, tcCmdMsg.CommandName, cmdMsg.CommandName)
			assert.Equal(t, tcCmdMsg.TransactionID, cmdMsg.TransactionID)
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
