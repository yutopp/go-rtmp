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
	r io.Reader

	cacheBuffer bytes.Buffer
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

func (dec *Decoder) Reset(r io.Reader) {
	dec.r = r
}

func (dec *Decoder) Decode(typeID TypeID, msg *Message) error {
	switch typeID {
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
		return fmt.Errorf("Unexpected message type(decode): ID = %d", typeID)
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
	*msg = &AudioMessage{
		Payload: dec.r, // Share an ownership of the reader
	}

	return nil
}

func (dec *Decoder) decodeVideoMessage(msg *Message) error {
	*msg = &VideoMessage{
		Payload: dec.r, // Share an ownership of the reader
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
	if err := dec.decodeDataMessage(dec.r, msg, func(r io.Reader) (AMFDecoder, EncodingType) {
		return amf0.NewDecoder(r), EncodingTypeAMF0
	}); err != nil {
		return err
	}

	return nil
}

func (dec *Decoder) decodeSharedObjectMessageAMF0(msg *Message) error {
	return fmt.Errorf("Not implemented: SharedObjectMessageAMF0")
}

func (dec *Decoder) decodeCommandMessageAMF0(msg *Message) error {
	if err := dec.decodeCommandMessage(dec.r, msg, func(r io.Reader) (AMFDecoder, EncodingType) {
		return amf0.NewDecoder(r), EncodingTypeAMF0
	}); err != nil {
		return err
	}

	return nil
}

func (dec *Decoder) decodeAggregateMessage(msg *Message) error {
	return fmt.Errorf("Not implemented: AggregateMessage")
}

func (dec *Decoder) decodeDataMessage(r io.Reader, msg *Message, f func(r io.Reader) (AMFDecoder, EncodingType)) error {
	d, encTy := f(r)

	var name string
	if err := d.Decode(&name); err != nil {
		return errors.Wrap(err, "Failed to decode name")
	}

	buf := &dec.cacheBuffer // TODO: Provide thread safety if needed
	buf.Reset()

	_, err := io.Copy(buf, dec.r)
	if err != nil {
		return errors.Wrap(err, "Failed to copy body payload")
	}

	// Copy ownership
	bin := make([]byte, len(buf.Bytes()))
	copy(bin, buf.Bytes())

	*msg = &DataMessage{
		Name:     name,
		Encoding: encTy,
		Body:     bin,
	}

	return nil
}

func (dec *Decoder) decodeCommandMessage(r io.Reader, msg *Message, f func(r io.Reader) (AMFDecoder, EncodingType)) error {
	d, encTy := f(r)

	var name string
	if err := d.Decode(&name); err != nil {
		return errors.Wrap(err, "Failed to decode name")
	}

	var transactionID int64
	if err := d.Decode(&transactionID); err != nil {
		return errors.Wrap(err, "Failed to decode transactionID")
	}

	buf := &dec.cacheBuffer // TODO: Provide thread safety if needed
	buf.Reset()

	_, err := io.Copy(buf, dec.r)
	if err != nil {
		return errors.Wrap(err, "Failed to copy body payload")
	}

	// Copy ownership
	bin := make([]byte, len(buf.Bytes()))
	copy(bin, buf.Bytes())

	*msg = &CommandMessage{
		CommandName:   name,
		TransactionID: transactionID,
		Encoding:      encTy,
		Body:          bin,
	}

	return nil
}
