//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChunkBasicHeader(t *testing.T) {
	type testCase struct {
		name   string
		value  *chunkBasicHeader
		binary []byte
	}
	testCases := []testCase{
		testCase{
			name: "cs normal 1",
			value: &chunkBasicHeader{
				fmt:           1,
				chunkStreamID: 2,
			},
			binary: []byte{
				// 0b01       : fmt  = 1
				//     000010 : csID = 2
				0x42,
			},
		},
		testCase{
			name: "cs normal 2",
			value: &chunkBasicHeader{
				fmt:           2,
				chunkStreamID: 63,
			},
			binary: []byte{
				// 0b10       : fmt  = 2
				//     111111 : csID = 63
				0xbf,
			},
		},

		testCase{
			name: "cs medium 1",
			value: &chunkBasicHeader{
				fmt:           0,
				chunkStreamID: 64,
			},
			binary: []byte{
				// 0b00       : fmt  = 0
				//     000000 : csID(marker) = 0
				//
				0x00,
				// 0b00000000 : csID = 0 = 64 - 64
				0x00,
			},
		},
		testCase{
			name: "cs medium 2",
			value: &chunkBasicHeader{
				fmt:           1,
				chunkStreamID: 319,
			},
			binary: []byte{
				// 0b01       : fmt  = 1
				//     000000 : csID(marker) = 0
				0x40,
				// 0b11111111 : csID = 255 = 319 - 64
				0xff,
			},
		},

		testCase{
			name: "cs large 1",
			value: &chunkBasicHeader{
				fmt:           3,
				chunkStreamID: 320,
			},
			binary: []byte{
				// 0b11       : fmt  = 3
				//     000001 : csID(marker) = 0
				//
				0xc1,
				// 0b00000000,00000001 : csID = 256 = 320 - 64
				0x00, 0x01,
			},
		},
		testCase{
			name: "cs large 2",
			value: &chunkBasicHeader{
				fmt:           0,
				chunkStreamID: 65599,
			},
			binary: []byte{
				// 0b00       : fmt  = 0
				//     000001 : csID(marker) = 1
				0x01,
				// 0b11111111,11111111 : csID = 65535 = 65599 - 64
				0xff, 0xff,
			},
		},
	}

	t.Run("Encode", func(t *testing.T) {
		for _, tc := range testCases {
			tc := tc // capture

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				buf := new(bytes.Buffer)
				err := encodeChunkBasicHeader(buf, tc.value)
				assert.Nil(t, err)
				assert.Equal(t, tc.binary, buf.Bytes())
			})
		}
	})

	t.Run("Decode", func(t *testing.T) {
		for _, tc := range testCases {
			tc := tc // capture

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				bin := make([]byte, len(tc.binary))
				copy(bin, tc.binary) // copy ownership

				buf := bytes.NewBuffer(bin)
				var mh chunkBasicHeader
				err := decodeChunkBasicHeader(buf, &mh)
				assert.Nil(t, err)
				assert.Equal(t, tc.value, &mh)
			})
		}
	})
}

func TestChunkBasicHeaderError(t *testing.T) {
	buf := new(bytes.Buffer)
	err := encodeChunkBasicHeader(buf, &chunkBasicHeader{
		fmt:           3,
		chunkStreamID: 65600,
	})
	assert.NotNil(t, err)
}

func TestChunkMessageHeader(t *testing.T) {
	basic := &chunkMessageHeader{
		timestamp:       10,
		timestampDelta:  10,
		messageLength:   10,
		messageTypeID:   10,
		messageStreamID: 10,
	}

	extendedBoundary := &chunkMessageHeader{
		timestamp:       16777215,
		timestampDelta:  16777215,
		messageLength:   20,
		messageTypeID:   20,
		messageStreamID: 20,
	}

	extended := &chunkMessageHeader{
		timestamp:       16777216,
		timestampDelta:  16777216,
		messageLength:   30,
		messageTypeID:   30,
		messageStreamID: 30,
	}

	type testCase struct {
		name   string
		fmt    byte
		value  *chunkMessageHeader
		binary []byte
	}
	testCases := []testCase{
		testCase{
			name: "basic fmt 0",
			fmt:  0,
			value: &chunkMessageHeader{
				timestamp:       basic.timestamp,
				messageLength:   basic.messageLength,
				messageTypeID:   basic.messageTypeID,
				messageStreamID: basic.messageStreamID,
			},
			binary: []byte{
				// Timestamp 10(BigEndian, 24bits)
				0x00, 0x00, 0x0a,
				// MessageLength 10(BigEndian, 24bits)
				0x00, 0x00, 0x0a,
				// MessageTypeID 10(8bits)
				0x0a,
				// MessageStreamID 10(*LittleEndian*, 32bits)
				0x0a, 0x00, 0x00, 0x00,
			},
		},
		testCase{
			name: "basic fmt 1",
			fmt:  1,
			value: &chunkMessageHeader{
				timestampDelta: basic.timestampDelta,
				messageLength:  basic.messageLength,
				messageTypeID:  basic.messageTypeID,
			},
			binary: []byte{
				// Timestamp Delta 10(BigEndian, 24bits)
				0x00, 0x00, 0x0a,
				// MessageLength 10(BigEndian, 24bits)
				0x00, 0x00, 0x0a,
				// MessageTypeID 10(8bits)
				0x0a,
			},
		},
		testCase{
			name: "basic fmt 2",
			fmt:  2,
			value: &chunkMessageHeader{
				timestampDelta: basic.timestampDelta,
			},
			binary: []byte{
				// Timestamp Delta 10(BigEndian, 24bits)
				0x00, 0x00, 0x0a,
			},
		},

		testCase{
			name: "extended boundary fmt 0",
			fmt:  0,
			value: &chunkMessageHeader{
				timestamp:       extendedBoundary.timestamp,
				messageLength:   extendedBoundary.messageLength,
				messageTypeID:   extendedBoundary.messageTypeID,
				messageStreamID: extendedBoundary.messageStreamID,
			},
			binary: []byte{
				// Timestamp MARKER(BigEndian, 24bits)
				0xff, 0xff, 0xff,
				// MessageLength 20(BigEndian, 24bits)
				0x00, 0x00, 0x14,
				// MessageTypeID 20(8bits)
				0x14,
				// MessageStreamID 20(*LittleEndian*, 32bits)
				0x14, 0x00, 0x00, 0x00,
				// ExtendTimestamp 16777215(BigEndian, 32bits)
				0x00, 0xff, 0xff, 0xff,
			},
		},
		testCase{
			name: "extended boundary fmt 1",
			fmt:  1,
			value: &chunkMessageHeader{
				timestampDelta: extendedBoundary.timestampDelta,
				messageLength:  extendedBoundary.messageLength,
				messageTypeID:  extendedBoundary.messageTypeID,
			},
			binary: []byte{
				// Timestamp Delta MARKER(BigEndian, 24bits)
				0xff, 0xff, 0xff,
				// MessageLength 20(BigEndian, 24bits)
				0x00, 0x00, 0x14,
				// MessageTypeID 20(8bits)
				0x14,
				// ExtendTimestamp Delta 16777215(BigEndian, 32bits)
				0x00, 0xff, 0xff, 0xff,
			},
		},
		testCase{
			name: "extended boundary fmt 2",
			fmt:  2,
			value: &chunkMessageHeader{
				timestampDelta: extendedBoundary.timestampDelta,
			},
			binary: []byte{
				// Timestamp Delta MARKER(BigEndian, 24bits)
				0xff, 0xff, 0xff,
				// ExtendTimestamp Delta 0(BigEndian, 32bits)
				0x00, 0xff, 0xff, 0xff,
			},
		},

		testCase{
			name: "extended fmt 0",
			fmt:  0,
			value: &chunkMessageHeader{
				timestamp:       extended.timestamp,
				messageLength:   extended.messageLength,
				messageTypeID:   extended.messageTypeID,
				messageStreamID: extended.messageStreamID,
			},
			binary: []byte{
				// Timestamp MARKER(BigEndian, 24bits)
				0xff, 0xff, 0xff,
				// MessageLength 30(BigEndian, 24bits)
				0x00, 0x00, 0x1e,
				// MessageTypeID 30(8bits)
				0x1e,
				// MessageStreamID 30(*LittleEndian*, 32bits)
				0x1e, 0x00, 0x00, 0x00,
				// ExtendTimestamp 16777216(BigEndian, 32bits)
				0x01, 0x00, 0x00, 0x00,
			},
		},
		testCase{
			name: "extended fmt 1",
			fmt:  1,
			value: &chunkMessageHeader{
				timestampDelta: extended.timestampDelta,
				messageLength:  extended.messageLength,
				messageTypeID:  extended.messageTypeID,
			},
			binary: []byte{
				// Timestamp Delta MARKER(BigEndian, 24bits)
				0xff, 0xff, 0xff,
				// MessageLength 30(BigEndian, 24bits)
				0x00, 0x00, 0x1e,
				// MessageTypeID 30(8bits)
				0x1e,
				// ExtendTimestamp Delta 16777216(BigEndian, 32bits)
				0x01, 0x00, 0x00, 0x00,
			},
		},
		testCase{
			name: "extended fmt 2",
			fmt:  2,
			value: &chunkMessageHeader{
				timestampDelta: extended.timestampDelta,
			},
			binary: []byte{
				// Timestamp Delta MARKER(BigEndian, 24bits)
				0xff, 0xff, 0xff,
				// ExtendTimestamp Delta 16777216(BigEndian, 32bits)
				0x01, 0x00, 0x00, 0x00,
			},
		},

		testCase{
			name:   "fmt 3",
			fmt:    3,
			value:  &chunkMessageHeader{},
			binary: []byte(nil),
		},
	}

	t.Run("Encode", func(t *testing.T) {
		for _, tc := range testCases {
			tc := tc // capture

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				buf := new(bytes.Buffer)
				err := encodeChunkMessageHeader(buf, tc.fmt, tc.value)
				assert.Nil(t, err)
				assert.Equal(t, tc.binary, buf.Bytes())
			})
		}
	})

	t.Run("Decode", func(t *testing.T) {
		for _, tc := range testCases {
			tc := tc // capture

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				bin := make([]byte, len(tc.binary))
				copy(bin, tc.binary) // copy ownership

				buf := bytes.NewBuffer(bin)
				var mh chunkMessageHeader
				err := decodeChunkMessageHeader(buf, tc.fmt, &mh)
				assert.Nil(t, err)
				assert.Equal(t, tc.value, &mh)
			})
		}
	})
}
