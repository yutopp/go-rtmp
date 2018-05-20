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

	"github.com/yutopp/amf0-go"
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
	case *UserCtrl:
		return enc.encodeUserCtrl(msg)
	case *CtrlWinAckSize:
		return enc.encodeCtrlWinAckSize(msg)
	case *SetPeerBandwidth:
		return enc.encodeSetPeerBandwidth(msg)
	case *CommandMessageAMF0:
		return enc.encodeCommandMessageAMF0(msg)
	}

	panic("unreachable!")
}

func (enc *Encoder) encodeUserCtrl(msg *UserCtrl) error {
	ucmEnc := NewUserControlEventEncoder(enc.w)
	return ucmEnc.Encode(msg.Event)
}

func (enc *Encoder) encodeCtrlWinAckSize(m *CtrlWinAckSize) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, m.Size) // [0:4]

	enc.w.Write(buf)

	return nil
}

func (enc *Encoder) encodeSetPeerBandwidth(m *SetPeerBandwidth) error {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf, m.Size) // [0:4]
	buf[4] = m.Limit

	enc.w.Write(buf)

	return nil
}

func (enc *Encoder) encodeCommandMessageAMF0(m *CommandMessageAMF0) error {
	amfEnc := amf0.NewEncoder(enc.w)
	if err := amfEnc.Encode(m.CommandName); err != nil {
		return err
	}
	if err := amfEnc.Encode(m.TransactionID); err != nil {
		return err
	}

	// if command is null, write null
	if m.Command == nil {
		return amfEnc.Encode(nil)
	}

	switch cmd := m.Command.(type) {
	case *NetConnectionResult:
		for _, object := range cmd.Objects {
			if err := amfEnc.Encode(object); err != nil {
				return err
			}
		}
		return nil

	case *NetStreamOnStatus:
		if err := amfEnc.Encode(cmd.CommandObject); err != nil {
			return err
		}
		if err := amfEnc.Encode(cmd.InfoObject); err != nil {
			return err
		}
		return nil

	default:
		panic("unsupported command(writer)")
	}
}
