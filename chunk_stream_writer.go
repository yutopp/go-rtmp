//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"context"
	"sync"
)

type ChunkStreamWriter struct {
	ChunkStreamReader

	doneCh   chan struct{}
	closeCh  chan struct{}
	lastErr  error
	aqM      sync.Mutex
	newChunk bool
}

func (w *ChunkStreamWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *ChunkStreamWriter) Wait(ctx context.Context) error {
	w.aqM.Lock()
	defer w.aqM.Unlock()

	select {
	case <-w.doneCh:
		if w.lastErr != nil {
			return w.lastErr
		}

		w.doneCh = make(chan struct{})
		return nil

	case <-w.closeCh:
		return w.lastErr

	case <-ctx.Done():
		return ctx.Err()
	}
}
