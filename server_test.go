//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerCanClose(t *testing.T) {
	srv := NewServer(&ServerConfig{})

	go func(ch <-chan time.Time) {
		<-ch
		err := srv.Close()
		assert.Nil(t, err)
	}(time.After(1 * time.Second))

	l, err := net.Listen("tcp", "127.0.0.1:")
	assert.Nil(t, err)

	err = srv.Serve(l)
	assert.Equal(t, ErrClosed, err)
}
