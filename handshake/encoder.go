//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package handshake

import (
	"encoding/binary"
	"io"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

func (e *Encoder) EncodeS0C0(h *S0C0) error {
	buf := [1]byte{byte(*h)}

	_, err := e.w.Write(buf[:])
	if err != nil {
		return err
	}

	return nil
}

func (e *Encoder) EncodeS1C1(h *S1C1) error {
	buf := [4]byte{}

	binary.BigEndian.PutUint32(buf[:], h.Time)
	if _, err := e.w.Write(buf[:]); err != nil {
		return err
	}

	if _, err := e.w.Write(h.Version[:]); err != nil {
		return err
	}

	if _, err := e.w.Write(h.Random[:]); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) EncodeS2C2(h *S2C2) error {
	buf := [4]byte{}

	binary.BigEndian.PutUint32(buf[:], h.Time)
	if _, err := e.w.Write(buf[:]); err != nil {
		return err
	}

	binary.BigEndian.PutUint32(buf[:], h.Time2)
	if _, err := e.w.Write(buf[:]); err != nil {
		return err
	}

	if _, err := e.w.Write(h.Random[:]); err != nil {
		return err
	}

	return nil
}
