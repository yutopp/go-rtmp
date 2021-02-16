//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertEqualMessage(t *testing.T, expected, actual Message) {
	assert.Equal(t, expected.TypeID(), actual.TypeID())

	switch expected := expected.(type) {
	case *AudioMessage:
		actual, ok := actual.(*AudioMessage)
		assert.True(t, ok)

		assertEqualPayload(t, expected.Payload, actual.Payload)

	case *VideoMessage:
		actual, ok := actual.(*VideoMessage)
		assert.True(t, ok)

		assertEqualPayload(t, expected.Payload, actual.Payload)

	case *DataMessage:
		actual, ok := actual.(*DataMessage)
		assert.True(t, ok)

		assert.Equal(t, expected.Name, actual.Name)
		assert.Equal(t, expected.Encoding, actual.Encoding)
		assertEqualPayload(t, expected.Body, actual.Body)

	case *CommandMessage:
		actual, ok := actual.(*CommandMessage)
		assert.True(t, ok)

		assert.Equal(t, expected.CommandName, actual.CommandName)
		assert.Equal(t, expected.TransactionID, actual.TransactionID)
		assert.Equal(t, expected.Encoding, actual.Encoding)
		assertEqualPayload(t, expected.Body, actual.Body)

	default:
		assert.Equal(t, expected, actual)
	}
}

func assertEqualPayload(t *testing.T, expected, actual io.Reader) {
	expectedBin, err := ioutil.ReadAll(expected)
	assert.Nil(t, err)
	switch p := expected.(type) {
	case *bytes.Reader:
		defer func() {
			_, _ = p.Seek(0, io.SeekStart) // Restore test case states
		}()
	default:
		t.FailNow()
	}
	assert.NotZero(t, len(expectedBin))

	actualBin, err := ioutil.ReadAll(actual)
	assert.Nil(t, err)
	assert.NotZero(t, len(actualBin))

	assert.Equal(t, expectedBin, actualBin)
}
