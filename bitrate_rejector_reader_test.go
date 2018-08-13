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
	"time"
)

func TestBitrateRejectorReaderRejected(t *testing.T) {
	br := bytes.NewReader(make([]byte, 4096))
	maxBitrateKbps := uint32(8) // 8Kbps

	r := NewBitrateRejectorReader(br, maxBitrateKbps)
	r.now = func() time.Time {
		return time.Unix(0, 0)
	}

	// Read 1KB=8Kbits
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, 1024, n)

	// simulate 1 sec
	r.now = func() time.Time {
		return time.Unix(1, 0)
	}

	// Read 8Kbits
	n, err = r.Read(buf)
	assert.EqualError(t, err, "Bitrate exceeded: Limit = 8kbps, Value = 16kbps")
}

func TestBitrateRejectorReaderAccepted(t *testing.T) {
	br := bytes.NewReader(make([]byte, 4096))
	maxBitrateKbps := uint32(8) // 8Kbps

	r := NewBitrateRejectorReader(br, maxBitrateKbps)
	r.now = func() time.Time {
		return time.Unix(0, 0)
	}

	buf := make([]byte, 512)
	for i := 0; i < 4096/512; i++ {
		// Read 512Bytes=4Kbits
		n, err := r.Read(buf)
		assert.Nil(t, err)
		assert.Equal(t, 512, n)

		// simulate 1 sec per loop
		r.now = func() time.Time {
			return time.Unix(int64(i), 0)
		}
	}
}
