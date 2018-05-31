//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
	"sync"
)

type ChunkStreamWriter struct {
	basicHeader   chunkBasicHeader
	messageHeader chunkMessageHeader

	timestamp         uint32
	timestampForDelta uint32
	messageLength     uint32 // max, 24bits
	messageTypeID     byte
	messageStreamID   uint32

	m   sync.Mutex
	buf bytes.Buffer
}

func (w *ChunkStreamWriter) Read(b []byte) (int, error) {
	return w.buf.Read(b)
}

func (w *ChunkStreamWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}
