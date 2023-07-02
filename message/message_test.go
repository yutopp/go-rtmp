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

	"github.com/stretchr/testify/require"
)

func assertEqualMessage(t *testing.T, expected, actual Message) {
	require.Equal(t, expected.TypeID(), actual.TypeID())

	switch expected := expected.(type) {
	case *AudioMessage:
		actual, ok := actual.(*AudioMessage)
		require.True(t, ok)

		assertEqualPayload(t, expected.Payload, actual.Payload)

	case *VideoMessage:
		actual, ok := actual.(*VideoMessage)
		require.True(t, ok)

		assertEqualPayload(t, expected.Payload, actual.Payload)

	case *DataMessage:
		actual, ok := actual.(*DataMessage)
		require.True(t, ok)

		require.Equal(t, expected.Name, actual.Name)
		require.Equal(t, expected.Encoding, actual.Encoding)
		assertEqualPayload(t, expected.Body, actual.Body)

	case *CommandMessage:
		actual, ok := actual.(*CommandMessage)
		require.True(t, ok)

		require.Equal(t, expected.CommandName, actual.CommandName)
		require.Equal(t, expected.TransactionID, actual.TransactionID)
		require.Equal(t, expected.Encoding, actual.Encoding)
		assertEqualPayload(t, expected.Body, actual.Body)

	default:
		require.Equal(t, expected, actual)
	}
}

func assertEqualPayload(t *testing.T, expected, actual io.Reader) {
	expectedBin, err := ioutil.ReadAll(expected)
	require.Nil(t, err)
	switch p := expected.(type) {
	case *bytes.Reader:
		defer func() {
			_, _ = p.Seek(0, io.SeekStart) // Restore test case states
		}()
	default:
		t.FailNow()
	}
	require.NotZero(t, len(expectedBin))

	actualBin, err := ioutil.ReadAll(actual)
	require.Nil(t, err)
	require.NotZero(t, len(actualBin))

	require.Equal(t, expectedBin, actualBin)
}
