//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientValidAddr(t *testing.T) {
	addr, err := makeValidAddr("host:123")
	require.Equal(t, nil, err)
	require.Equal(t, "host:123", addr)

	addr, err = makeValidAddr("host")
	require.Equal(t, nil, err)
	require.Equal(t, "host:1935", addr)

	addr, err = makeValidAddr("host:")
	require.Equal(t, nil, err)
	require.Equal(t, "host:", addr)

	addr, err = makeValidAddr(":1111")
	require.Equal(t, nil, err)
	require.Equal(t, ":1111", addr)
}
