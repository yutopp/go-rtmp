//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"

	"github.com/yutopp/go-amf0"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
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
	case *DataMessageAMF3:
		return enc.encodeDataMessageAMF3(msg)
	case *SharedObjectMessageAMF3:
		return enc.encodeSharedObjectMessageAMF3(msg)
	case *CommandMessageAMF3:
		return enc.encodeCommandMessageAMF3(msg)
	case *DataMessageAMF0:
		return enc.encodeDataMessageAMF0(msg)
	case *SharedObjectMessageAMF0:
		return enc.encodeSharedObjectMessageAMF0(msg)
	case *CommandMessageAMF0:
		return enc.encodeCommandMessageAMF0(msg)
	case *AggregateMessage:
		return enc.encodeAggregateMessage(msg)
	default:
		return fmt.Errorf("Unexpected message type(encode): ID = %d, Type = %+v", msg.TypeID(), reflect.TypeOf(msg))
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
	return fmt.Errorf("Not implemented: AbortMessage")
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
	binary.BigEndian.PutUint32(buf, m.Size) // [0:4]

	enc.w.Write(buf) // TODO: error check

	return nil
}

func (enc *Encoder) encodeSetPeerBandwidth(m *SetPeerBandwidth) error {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf, m.Size) // [0:4]
	buf[4] = byte(m.Limit)

	enc.w.Write(buf) // TODO: error check

	return nil
}

func (enc *Encoder) encodeAudioMessage(m *AudioMessage) error {
	if _, err := enc.w.Write(m.Payload); err != nil {
		return err
	}

	return nil
}

func (enc *Encoder) encodeVideoMessage(m *VideoMessage) error {
	if _, err := enc.w.Write(m.Payload); err != nil {
		return err
	}

	return nil
}

func (enc *Encoder) encodeDataMessageAMF3(m *DataMessageAMF3) error {
	return fmt.Errorf("Not implemented: DataMessageAMF3")
}

func (enc *Encoder) encodeSharedObjectMessageAMF3(m *SharedObjectMessageAMF3) error {
	return fmt.Errorf("Not implemented: SharedObjectMessageAMF3")
}

func (enc *Encoder) encodeCommandMessageAMF3(m *CommandMessageAMF3) error {
	return fmt.Errorf("Not implemented: CommandMessageAMF3")
}

func (enc *Encoder) encodeDataMessageAMF0(m *DataMessageAMF0) error {
	return fmt.Errorf("Not implemented: DataMessageAMF0")
}

func (enc *Encoder) encodeSharedObjectMessageAMF0(m *SharedObjectMessageAMF0) error {
	return fmt.Errorf("Not implemented: SharedObjectMessageAMF0")
}

// TODO: support amf3
func (enc *Encoder) encodeCommandMessageAMF0(m *CommandMessageAMF0) error {
	amfEnc := amf0.NewEncoder(enc.w)
	if err := amfEnc.Encode(m.CommandName); err != nil {
		return err
	}
	if err := amfEnc.Encode(m.TransactionID); err != nil {
		return err
	}

	if m.Command == nil {
		return nil // Do nothing
	}

	args, err := m.Command.ToArgs()
	if err != nil {
		return err
	}

	for _, arg := range args {
		if err := amfEnc.Encode(arg); err != nil {
			return err
		}
	}

	return nil
}

func (enc *Encoder) encodeAggregateMessage(m *AggregateMessage) error {
	return fmt.Errorf("Not implemented: AggregateMessage")
}
