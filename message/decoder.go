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
	"github.com/pkg/errors"
	"io"

	"github.com/yutopp/go-amf0"
)

type Decoder struct {
	r      io.Reader
	typeID TypeID

	cacheBuffer bytes.Buffer
}

func NewDecoder(r io.Reader, typeID TypeID) *Decoder {
	return &Decoder{
		r:      r,
		typeID: typeID,
	}
}

func (dec *Decoder) Decode(msg *Message) error {
	switch dec.typeID {
	case TypeIDSetChunkSize:
		return dec.decodeSetChunkSize(msg)
	case TypeIDAbortMessage:
		return dec.decodeAbortMessage(msg)
	case TypeIDAck:
		return dec.decodeAck(msg)
	case TypeIDUserCtrl:
		return dec.decodeUserCtrl(msg)
	case TypeIDWinAckSize:
		return dec.decodeWinAckSize(msg)
	case TypeIDSetPeerBandwidth:
		return dec.decodeSetPeerBandwidth(msg)
	case TypeIDAudioMessage:
		return dec.decodeAudioMessage(msg)
	case TypeIDVideoMessage:
		return dec.decodeVideoMessage(msg)
	case TypeIDDataMessageAMF3:
		return dec.decodeDataMessageAMF3(msg)
	case TypeIDSharedObjectMessageAMF3:
		return dec.decodeSharedObjectMessageAMF3(msg)
	case TypeIDCommandMessageAMF3:
		return dec.decodeCommandMessageAMF3(msg)
	case TypeIDDataMessageAMF0:
		return dec.decodeDataMessageAMF0(msg)
	case TypeIDSharedObjectMessageAMF0:
		return dec.decodeSharedObjectMessageAMF0(msg)
	case TypeIDCommandMessageAMF0:
		return dec.decodeCommandMessageAMF0(msg)
	case TypeIDAggregateMessage:
		return dec.decodeAggregateMessage(msg)
	default:
		return fmt.Errorf("Unexpected message type(decode): ID = %d", dec.typeID)
	}
}

func (dec *Decoder) decodeSetChunkSize(msg *Message) error {
	buf := make([]byte, 4)
	if _, err := io.ReadAtLeast(dec.r, buf, 4); err != nil {
		return err
	}

	total := binary.BigEndian.Uint32(buf)

	bit := (total & 0x80000000) >> 31 // 0b1000,0000... >> 31
	chunkSize := total & 0x7fffffff   // 0b0111,1111...

	if bit != 0 {
		return fmt.Errorf("Invalid format: bit must be 0")
	}

	if chunkSize == 0 {
		return fmt.Errorf("Invalid format: chunk size is 0")
	}

	*msg = &SetChunkSize{
		ChunkSize: chunkSize,
	}

	return nil
}

func (dec *Decoder) decodeAbortMessage(msg *Message) error {
	buf := make([]byte, 4)
	if _, err := io.ReadAtLeast(dec.r, buf, 4); err != nil {
		return err
	}

	chunkStreamID := binary.BigEndian.Uint32(buf)

	*msg = &AbortMessage{
		ChunkStreamID: chunkStreamID,
	}

	return nil
}

func (dec *Decoder) decodeAck(msg *Message) error {
	buf := make([]byte, 4)
	if _, err := io.ReadAtLeast(dec.r, buf, 4); err != nil {
		return err
	}

	sequenceNumber := binary.BigEndian.Uint32(buf)

	*msg = &Ack{
		SequenceNumber: sequenceNumber,
	}

	return nil
}

func (dec *Decoder) decodeUserCtrl(msg *Message) error {
	ucmDec := NewUserControlEventDecoder(dec.r)

	var event UserCtrlEvent
	if err := ucmDec.Decode(&event); err != nil {
		return errors.Wrapf(err, "Failed to decode UserCtrl")
	}

	*msg = &UserCtrl{
		Event: event,
	}

	return nil
}

func (dec *Decoder) decodeWinAckSize(msg *Message) error {
	buf := make([]byte, 4)
	if _, err := io.ReadAtLeast(dec.r, buf, 4); err != nil {
		return err
	}

	size := int32(binary.BigEndian.Uint32(buf))

	*msg = &WinAckSize{
		Size: size,
	}

	return nil
}

func (dec *Decoder) decodeSetPeerBandwidth(msg *Message) error {
	buf := make([]byte, 5)
	if _, err := io.ReadAtLeast(dec.r, buf, 5); err != nil {
		return err
	}

	size := int32(binary.BigEndian.Uint32(buf[0:4]))
	limit := LimitType(buf[4])

	*msg = &SetPeerBandwidth{
		Size:  size,
		Limit: limit,
	}

	return nil
}

func (dec *Decoder) decodeAudioMessage(msg *Message) error {
	buf := &dec.cacheBuffer // TODO: Provide thread safety if needed
	buf.Reset()

	_, err := io.Copy(buf, dec.r)
	if err != nil {
		return err
	}

	// Copy ownership
	bin := make([]byte, len(buf.Bytes()))
	copy(bin, buf.Bytes())

	*msg = &AudioMessage{
		Payload: bin,
	}

	return nil
}

func (dec *Decoder) decodeVideoMessage(msg *Message) error {
	buf := &dec.cacheBuffer // TODO: Provide thread safety if needed
	buf.Reset()

	_, err := io.Copy(buf, dec.r)
	if err != nil {
		return err
	}

	// Copy ownership
	bin := make([]byte, len(buf.Bytes()))
	copy(bin, buf.Bytes())

	*msg = &VideoMessage{
		Payload: bin,
	}

	return nil
}

func (dec *Decoder) decodeDataMessageAMF3(msg *Message) error {
	return fmt.Errorf("Not implemented: DataMessageAMF3")
}

func (dec *Decoder) decodeSharedObjectMessageAMF3(msg *Message) error {
	return fmt.Errorf("Not implemented: SharedObjectMessageAMF3")
}

func (dec *Decoder) decodeCommandMessageAMF3(msg *Message) error {
	return fmt.Errorf("Not implemented: CommandMessageAMF3")
}

func (dec *Decoder) decodeDataMessageAMF0(msg *Message) error {
	// Copy ownership
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, dec.r); err != nil {
		return errors.Wrap(err, "Failed to copy buffer of DataMessageAMF0")
	}

	d := amf0.NewDecoder(&buf)

	var body DataMessageAMF0
	if err := dec.decodeDataMessage(&buf, d, &body.DataMessage); err != nil {
		return err
	}

	*msg = &body

	return nil
}

func (dec *Decoder) decodeSharedObjectMessageAMF0(msg *Message) error {
	return fmt.Errorf("Not implemented: SharedObjectMessageAMF0")
}

func (dec *Decoder) decodeCommandMessageAMF0(msg *Message) error {
	// Copy ownership
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, dec.r); err != nil {
		return errors.Wrap(err, "Failed to copy buffer of CommandMessageAMF0")
	}

	d := amf0.NewDecoder(&buf)

	var body CommandMessageAMF0
	if err := dec.decodeCommandMessage(&buf, d, &body.CommandMessage); err != nil {
		return err
	}

	*msg = &body

	return nil
}

func (dec *Decoder) decodeAggregateMessage(msg *Message) error {
	return fmt.Errorf("Not implemented: AggregateMessage")
}

func (dec *Decoder) decodeDataMessage(r io.Reader, d AMFDecoder, dataMsg *DataMessage) error {
	var name string
	if err := d.Decode(&name); err != nil {
		return err
	}

	// Not decoding body in this function to make an user select a decoder of the body.
	// The caller of this function should call `msg.Decoder.Decode()` with arbitrary `MsgDecoder`
	bodyDec := &BodyDecoder{
		reader:     r,
		amfDec:     d,
		Value:      nil, // Caller can get a decoded value from this
		MsgDecoder: nil, // Caller MUST set this function
	}

	*dataMsg = DataMessage{
		Name:    name,
		Decoder: bodyDec,
	}

	return nil
}

func (dec *Decoder) decodeCommandMessage(r io.Reader, d AMFDecoder, cmdMsg *CommandMessage) error {
	var name string
	if err := d.Decode(&name); err != nil {
		return err
	}

	var transactionID int64
	if err := d.Decode(&transactionID); err != nil {
		return err
	}

	// Not decoding body in this function to make an user select a decoder of the body.
	// The caller of this function should call `msg.Decoder.Decode()` with arbitrary `MsgDecoder`
	bodyDec := &BodyDecoder{
		reader:     r,
		amfDec:     d,
		Value:      nil, // Caller can get a decoded value from this
		MsgDecoder: nil, // Caller MUST set this function
	}

	*cmdMsg = CommandMessage{
		CommandName:   name,
		TransactionID: transactionID,
		Decoder:       bodyDec,
	}

	return nil
}
