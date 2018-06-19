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
