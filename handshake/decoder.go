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

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

func (d *Decoder) DecodeS0C0(h *S0C0) error {
	buf := [1]byte{}

	if _, err := io.ReadAtLeast(d.r, buf[:], 1); err != nil {
		return err
	}
	*h = S0C0(buf[0])

	return nil
}

func (d *Decoder) DecodeS1C1(h *S1C1) error {
	var buf [4]byte

	if _, err := io.ReadAtLeast(d.r, buf[:], len(buf)); err != nil {
		return err
	}
	h.Time = binary.BigEndian.Uint32(buf[:])

	if _, err := io.ReadAtLeast(d.r, h.Version[:], len(h.Version)); err != nil {
		return err
	}

	if _, err := io.ReadAtLeast(d.r, h.Random[:], len(h.Random)); err != nil {
		return err
	}

	return nil
}

func (d *Decoder) DecodeS2C2(h *S2C2) error {
	var buf [4]byte

	if _, err := io.ReadAtLeast(d.r, buf[:], len(buf)); err != nil {
		return err
	}
	h.Time = binary.BigEndian.Uint32(buf[:])

	if _, err := io.ReadAtLeast(d.r, buf[:], len(buf)); err != nil {
		return err
	}
	h.Time2 = binary.BigEndian.Uint32(buf[:])

	if _, err := io.ReadAtLeast(d.r, h.Random[:], len(h.Random)); err != nil {
		return err
	}

	return nil
}
