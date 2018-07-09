//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/pkg/errors"
	"io"
	"time"
)

type BitrateRejectorReader struct {
	reader         io.Reader
	maxBitrateKbps uint32

	readSize uint64
	now      func() time.Time // for mock
	last     time.Time
}

func NewBitrateRejectorReader(r io.Reader, maxBitrateKbps uint32) *BitrateRejectorReader {
	return &BitrateRejectorReader{
		reader:         r,
		maxBitrateKbps: maxBitrateKbps,

		now: time.Now,
	}
}

func (r *BitrateRejectorReader) Read(b []byte) (int, error) {
	n, err := r.reader.Read(b)
	if err != nil {
		return 0, err
	}

	cur := r.now()
	if r.last.IsZero() {
		r.last = cur
	}
	diff := cur.Sub(r.last)
	r.readSize += uint64(n)

	if diff >= 1*time.Second {
		bitrateKbps := (float64(r.readSize) / float64(diff/time.Second)) * 8 / 1024.0
		// reset
		r.readSize = 0
		r.last = cur

		if bitrateKbps > float64(r.maxBitrateKbps) {
			return 0, errors.Errorf(
				"Bitrate exceeded: Limit = %vkbps, Value = %vkbps",
				r.maxBitrateKbps,
				bitrateKbps,
			)
		}
	}

	return n, err
}
