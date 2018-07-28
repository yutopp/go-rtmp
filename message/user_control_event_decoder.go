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

type UserControlEventDecoder struct {
	r io.Reader
}

func NewUserControlEventDecoder(r io.Reader) *UserControlEventDecoder {
	return &UserControlEventDecoder{
		r: r,
	}
}

func (dec *UserControlEventDecoder) Decode(msg *UserCtrlEvent) error {
	buf := make([]byte, 2)
	if _, err := io.ReadAtLeast(dec.r, buf, 2); err != nil {
		return err
	}

	eventType := binary.BigEndian.Uint16(buf)
	switch eventType {
	case 0: // UserCtrlEventStreamBegin
		return dec.decodeStreamBegin(msg)
	case 1: // UserCtrlEventStreamEOF
		return dec.decodeStreamEOF(msg)
	case 2: // UserCtrlEventStreamDry
		return dec.decodeStreamDry(msg)
	case 3: // UserCtrlEventSetBufferLength
		return dec.decodeSetBufferLength(msg)
	case 4: // UserCtrlEventStreamIsRecorded
		return dec.decodeStreamIsRecorded(msg)
	case 6: // UserCtrlEventPingRequest
		return dec.decodePingRequest(msg)
	case 7: // UserCtrlEventPingResponse
		return dec.decodePingResponse(msg)
	default:
		return errors.Errorf("Unsupported type for UserCtrl: TypeID = %d", eventType)
	}
}

func (dec *UserControlEventDecoder) decodeStreamBegin(msg *UserCtrlEvent) error {
	buf := make([]byte, 4)
	if _, err := io.ReadAtLeast(dec.r, buf, 4); err != nil {
		return err
	}

	streamID := binary.BigEndian.Uint32(buf)

	*msg = &UserCtrlEventStreamBegin{
		StreamID: streamID,
	}

	return nil
}

func (dec *UserControlEventDecoder) decodeStreamEOF(msg *UserCtrlEvent) error {
	buf := make([]byte, 4)
	if _, err := io.ReadAtLeast(dec.r, buf, 4); err != nil {
		return err
	}

	streamID := binary.BigEndian.Uint32(buf)

	*msg = &UserCtrlEventStreamEOF{
		StreamID: streamID,
	}

	return nil
}

func (dec *UserControlEventDecoder) decodeStreamDry(msg *UserCtrlEvent) error {
	buf := make([]byte, 4)
	if _, err := io.ReadAtLeast(dec.r, buf, 4); err != nil {
		return err
	}

	streamID := binary.BigEndian.Uint32(buf)

	*msg = &UserCtrlEventStreamDry{
		StreamID: streamID,
	}

	return nil
}

func (dec *UserControlEventDecoder) decodeSetBufferLength(msg *UserCtrlEvent) error {
	buf := make([]byte, 8)
	if _, err := io.ReadAtLeast(dec.r, buf, 8); err != nil {
		return err
	}

	streamID := binary.BigEndian.Uint32(buf[0:4])
	lengthMs := binary.BigEndian.Uint32(buf[4:8])

	*msg = &UserCtrlEventSetBufferLength{
		StreamID: streamID,
		LengthMs: lengthMs,
	}

	return nil
}

func (dec *UserControlEventDecoder) decodeStreamIsRecorded(msg *UserCtrlEvent) error {
	buf := make([]byte, 4)
	if _, err := io.ReadAtLeast(dec.r, buf, 4); err != nil {
		return err
	}

	streamID := binary.BigEndian.Uint32(buf)

	*msg = &UserCtrlEventStreamIsRecorded{
		StreamID: streamID,
	}

	return nil
}

func (dec *UserControlEventDecoder) decodePingRequest(msg *UserCtrlEvent) error {
	buf := make([]byte, 4)
	if _, err := io.ReadAtLeast(dec.r, buf, 4); err != nil {
		return err
	}

	timestamp := binary.BigEndian.Uint32(buf)

	*msg = &UserCtrlEventPingRequest{
		Timestamp: timestamp,
	}

	return nil
}

func (dec *UserControlEventDecoder) decodePingResponse(msg *UserCtrlEvent) error {
	buf := make([]byte, 4)
	if _, err := io.ReadAtLeast(dec.r, buf, 4); err != nil {
		return err
	}

	timestamp := binary.BigEndian.Uint32(buf)

	*msg = &UserCtrlEventPingResponse{
		Timestamp: timestamp,
	}

	return nil
}
