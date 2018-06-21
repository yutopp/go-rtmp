//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bufio"
	"io"
)

type ChunkStreamerWriter struct {
	writer io.Writer
}

func (w *ChunkStreamerWriter) Write(buf []byte) (int, error) {
	return w.writer.Write(buf)
}

func (w *ChunkStreamerWriter) Flush() error {
	bufw, ok := w.writer.(*bufio.Writer)
	if !ok {
		return nil
	}
	return bufw.Flush()
}
