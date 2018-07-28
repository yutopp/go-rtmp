//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"encoding/binary"
	"github.com/pkg/errors"
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

func (enc *UserControlEventEncoder) Encode(msg UserCtrlEvent) error {
	switch msg := msg.(type) {
	case *UserCtrlEventStreamBegin:
		return enc.encodeStreamBegin(msg)
	case *UserCtrlEventStreamEOF:
		return enc.encodeStreamEOF(msg)
	case *UserCtrlEventStreamDry:
		return enc.encodeStreamDry(msg)
	case *UserCtrlEventSetBufferLength:
		return enc.encodeSetBufferLength(msg)
	case *UserCtrlEventStreamIsRecorded:
		return enc.encodeStreamIsRecorded(msg)
	case *UserCtrlEventPingRequest:
		return enc.encodePingRequest(msg)
	case *UserCtrlEventPingResponse:
		return enc.encodePingResponse(msg)
	default:
		return errors.Errorf("Unsupported type for UserCtrl: Type = %T", msg)
	}
}

func (enc *UserControlEventEncoder) encodeStreamBegin(msg *UserCtrlEventStreamBegin) error {
	buf := make([]byte, 2+4)
	binary.BigEndian.PutUint16(buf[0:2], 0)           // [0:2]: ID=0
	binary.BigEndian.PutUint32(buf[2:], msg.StreamID) // [2:6]

	_, err := enc.w.Write(buf)

	return err
}

func (enc *UserControlEventEncoder) encodeStreamEOF(msg *UserCtrlEventStreamEOF) error {
	buf := make([]byte, 2+4)
	binary.BigEndian.PutUint16(buf[0:2], 1)           // [0:2]: ID=1
	binary.BigEndian.PutUint32(buf[2:], msg.StreamID) // [2:6]

	_, err := enc.w.Write(buf)

	return err
}

func (enc *UserControlEventEncoder) encodeStreamDry(msg *UserCtrlEventStreamDry) error {
	buf := make([]byte, 2+4)
	binary.BigEndian.PutUint16(buf[0:2], 2)           // [0:2]: ID=2
	binary.BigEndian.PutUint32(buf[2:], msg.StreamID) // [2:6]

	_, err := enc.w.Write(buf)

	return err
}

func (enc *UserControlEventEncoder) encodeSetBufferLength(msg *UserCtrlEventSetBufferLength) error {
	buf := make([]byte, 2+4+4)
	binary.BigEndian.PutUint16(buf[0:2], 3)             // [0:2]: ID=e
	binary.BigEndian.PutUint32(buf[2:6], msg.StreamID)  // [2:6]
	binary.BigEndian.PutUint32(buf[6:10], msg.LengthMs) // [6:10]

	_, err := enc.w.Write(buf)

	return err
}

func (enc *UserControlEventEncoder) encodeStreamIsRecorded(msg *UserCtrlEventStreamIsRecorded) error {
	buf := make([]byte, 2+4)
	binary.BigEndian.PutUint16(buf[0:2], 4)           // [0:2]: ID=4
	binary.BigEndian.PutUint32(buf[2:], msg.StreamID) // [2:6]

	_, err := enc.w.Write(buf)

	return err
}

func (enc *UserControlEventEncoder) encodePingRequest(msg *UserCtrlEventPingRequest) error {
	buf := make([]byte, 2+4)
	binary.BigEndian.PutUint16(buf[0:2], 6)            // [0:2]: ID=6
	binary.BigEndian.PutUint32(buf[2:], msg.Timestamp) // [2:6]

	_, err := enc.w.Write(buf)

	return err
}

func (enc *UserControlEventEncoder) encodePingResponse(msg *UserCtrlEventPingResponse) error {
	buf := make([]byte, 2+4)
	binary.BigEndian.PutUint16(buf[0:2], 7)            // [0:2]: ID=7
	binary.BigEndian.PutUint32(buf[2:], msg.Timestamp) // [2:6]

	_, err := enc.w.Write(buf)

	return err
}
