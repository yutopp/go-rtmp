//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

func (enc *Encoder) Reset(w io.Writer) {
	enc.w = w
}

// Encode
func (enc *Encoder) Encode(msg Message) error {
	switch msg := msg.(type) {
	case *SetChunkSize:
		return enc.encodeSetChunkSize(msg)
	case *AbortMessage:
		return enc.encodeAbortMessage(msg)
	case *Ack:
		return enc.encodeAck(msg)
	case *UserCtrl:
		return enc.encodeUserCtrl(msg)
	case *WinAckSize:
		return enc.encodeWinAckSize(msg)
	case *SetPeerBandwidth:
		return enc.encodeSetPeerBandwidth(msg)
	case *AudioMessage:
		return enc.encodeAudioMessage(msg)
	case *VideoMessage:
		return enc.encodeVideoMessage(msg)
	case *DataMessage:
		return enc.encodeDataMessage(msg)
	case *SharedObjectMessageAMF3:
		return enc.encodeSharedObjectMessageAMF3(msg)
	case *CommandMessage:
		return enc.encodeCommandMessage(msg)
	case *SharedObjectMessageAMF0:
		return enc.encodeSharedObjectMessageAMF0(msg)
	case *AggregateMessage:
		return enc.encodeAggregateMessage(msg)
	default:
		return fmt.Errorf("Unexpected message type(encode): ID = %d, Type = %T", msg.TypeID(), msg)
	}
}

func (enc *Encoder) encodeSetChunkSize(m *SetChunkSize) error {
	if m.ChunkSize < 1 || m.ChunkSize > 0x7fffffff {
		return fmt.Errorf("Invalid format: chunk size is out of range [1, 0x80000000)")
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, m.ChunkSize&0x7fffffff) // 0b0111,1111...

	if _, err := enc.w.Write(buf); err != nil { // TODO: length check
		return err
	}

	return nil
}

func (enc *Encoder) encodeAbortMessage(m *AbortMessage) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, m.ChunkStreamID) // [0:4]

	if _, err := enc.w.Write(buf); err != nil { // TODO: length check
		return err
	}

	return nil
}

func (enc *Encoder) encodeAck(m *Ack) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, m.SequenceNumber) // [0:4]

	if _, err := enc.w.Write(buf); err != nil { // TODO: length check
		return err
	}

	return nil
}

func (enc *Encoder) encodeUserCtrl(msg *UserCtrl) error {
	ucmEnc := NewUserControlEventEncoder(enc.w)
	return ucmEnc.Encode(msg.Event)
}

func (enc *Encoder) encodeWinAckSize(m *WinAckSize) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(m.Size)) // [0:4]

	if _, err := enc.w.Write(buf); err != nil { // TODO: length check
		return err
	}

	return nil
}

func (enc *Encoder) encodeSetPeerBandwidth(m *SetPeerBandwidth) error {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf, uint32(m.Size)) // [0:4]
	buf[4] = byte(m.Limit)

	if _, err := enc.w.Write(buf); err != nil { // TODO: length check
		return err
	}

	return nil
}

func (enc *Encoder) encodeAudioMessage(m *AudioMessage) error {
	if _, err := io.Copy(enc.w, m.Payload); err != nil {
		return err
	}

	return nil
}

func (enc *Encoder) encodeVideoMessage(m *VideoMessage) error {
	if _, err := io.Copy(enc.w, m.Payload); err != nil {
		return err
	}

	return nil
}

func (enc *Encoder) encodeSharedObjectMessageAMF3(m *SharedObjectMessageAMF3) error {
	return fmt.Errorf("Not implemented: SharedObjectMessageAMF3")
}

func (enc *Encoder) encodeDataMessage(m *DataMessage) error {
	e := NewAMFEncoder(enc.w, m.Encoding)

	if err := e.Encode(m.Name); err != nil {
		return err
	}

	if _, err := io.Copy(enc.w, bytes.NewReader(m.Body)); err != nil {
		return err
	}

	return nil
}

func (enc *Encoder) encodeSharedObjectMessageAMF0(m *SharedObjectMessageAMF0) error {
	return fmt.Errorf("Not implemented: SharedObjectMessageAMF0")
}

func (enc *Encoder) encodeCommandMessage(m *CommandMessage) error {
	e := NewAMFEncoder(enc.w, m.Encoding)

	if err := e.Encode(m.CommandName); err != nil {
		return err
	}
	if err := e.Encode(m.TransactionID); err != nil {
		return err
	}

	if _, err := io.Copy(enc.w, bytes.NewReader(m.Body)); err != nil {
		return err
	}

	return nil
}

func (enc *Encoder) encodeAggregateMessage(m *AggregateMessage) error {
	return fmt.Errorf("Not implemented: AggregateMessage")
}
