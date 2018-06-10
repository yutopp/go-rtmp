//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bufio"
)

type ChunkStreamerReader struct {
	bufr           *bufio.Reader
	totalReadBytes uint64
}

func (r *ChunkStreamerReader) Read(b []byte) (int, error) {
	n, err := r.bufr.Read(b)
	r.totalReadBytes += uint64(n)
	return n, err
}
