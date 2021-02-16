//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeCommon(t *testing.T) {
	for _, tc := range testCases {
		tc := tc // capture

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)

			enc := NewEncoder(buf)
			err := enc.Encode(tc.Value)
			assert.Nil(t, err)
			assert.Equal(t, tc.Binary, buf.Bytes())
		})
	}
}
