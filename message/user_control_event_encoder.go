//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"encoding/binary"
	"io"
)

type UserControlEventEncoder struct {
	w io.Writer
}

func NewUserControlEventEncoder(w io.Writer) *UserControlEventEncoder {
	return &UserControlEventEncoder{
		w: w,
	}
}

func (enc *UserControlEventEncoder) Encode(msg interface{}) error {
	switch msg := msg.(type) {
	case *StreamBegin:
		return enc.encodeStreamBegin(msg)
	default:
		panic("unreachable")
	}
}

func (enc *UserControlEventEncoder) encodeStreamBegin(msg *StreamBegin) error {
	buf := make([]byte, 2+4)
	binary.BigEndian.PutUint16(buf[0:2], 0)           // [0:2]
	binary.BigEndian.PutUint32(buf[2:], msg.StreamID) // [2:6]

	_, err := enc.w.Write(buf)
	return err
}
