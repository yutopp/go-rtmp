//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
)

// Must call Close after reading.
type ChunkStreamReader struct {
	basicHeader   chunkBasicHeader
	messageHeader chunkMessageHeader

	timestamp       uint32
	timestampDelta  uint32
	messageLength   uint32 // max, 24bits
	messageTypeID   byte
	messageStreamID uint32

	buf       bytes.Buffer
	completed bool
}

func (r *ChunkStreamReader) Read(b []byte) (int, error) {
	return r.buf.Read(b)
}
