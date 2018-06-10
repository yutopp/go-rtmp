//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/yutopp/go-amf0"
)

type Decoder struct {
	r      io.Reader
	typeID TypeID
}

func NewDecoder(r io.Reader, typeID TypeID) *Decoder {
	return &Decoder{
		r:      r,
		typeID: typeID,
	}
}

func (dec *Decoder) Decode(msg *Message) error {
	switch dec.typeID {
	case TypeIDAudioMessage:
		return dec.decodeAudioMessage(msg)
	case TypeIDVideoMessage:
		return dec.decodeVideoMessage(msg)
	case TypeIDDataMessageAMF0:
		return dec.decodeDataMessageAMF0(msg)
	case TypeIDCommandMessageAMF0:
		return dec.decodeCommandMessageAMF0(msg)
	default:
		return fmt.Errorf("Unexpected message type: %d", dec.typeID)
	}
}

func (dec *Decoder) decodeAudioMessage(msg *Message) error {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, dec.r)
	if err != nil {
		return err
	}

	*msg = &AudioMessage{
		Payload: buf.Bytes(),
	}

	return nil
}

func (dec *Decoder) decodeVideoMessage(msg *Message) error {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, dec.r)
	if err != nil {
		return err
	}

	*msg = &VideoMessage{
		Payload: buf.Bytes(),
	}

	return nil
}

func (dec *Decoder) decodeDataMessageAMF3(msg *Message) error {
	return fmt.Errorf("Not implemented: DataMessageAMF3")
}

func (dec *Decoder) decodeDataMessageAMF0(msg *Message) error {
	d := amf0.NewDecoder(dec.r)
	var body DataMessageAMF0
	if err := dec.decodeDataMessage(d, &body.DataMessage); err != nil {
		return err
	}

	*msg = &body

	return nil
}

// TODO: support amf3
func (dec *Decoder) decodeDataMessage(d *amf0.Decoder, dataMsg *DataMessage) error {
	var name string
	if err := d.Decode(&name); err != nil {
		return err
	}
	log.Printf("name = %+v", name)

	var data interface{}
	switch name {
	case "onMetaData":
		var metadata map[string]interface{}
		if err := d.Decode(&metadata); err != nil {
			return err
		}
		log.Printf("onMetaData: metadata = %+v", metadata)
		data = &NetStreamOnMetaData{
			MetaData: metadata,
		}

	case "@setDataFrame":
		// TODO: implement
		log.Println("Ignored data message: @setDataFrame")

	default:
		return errors.New("Not supported data message: " + name)
	}

	*dataMsg = DataMessage{
		Name: name,
		Data: data,
	}

	return nil
}

func (dec *Decoder) decodeCommandMessageAMF3(msg *Message) error {
	return fmt.Errorf("Not implemented: CommandMessageAMF3")
}

func (dec *Decoder) decodeCommandMessageAMF0(msg *Message) error {
	d := amf0.NewDecoder(dec.r)
	var body CommandMessageAMF0
	if err := dec.decodeCommandMessage(d, &body.CommandMessage); err != nil {
		return err
	}

	*msg = &body

	return nil
}

// TODO: support amf3
func (dec *Decoder) decodeCommandMessage(d *amf0.Decoder, cmdMsg *CommandMessage) error {
	var name string
	if err := d.Decode(&name); err != nil {
		return err
	}
	log.Printf("name = %+v", name)

	var transactionID int64
	if err := d.Decode(&transactionID); err != nil {
		return err
	}

	log.Printf("transactionID = %+v", transactionID)

	var args []interface{}
	switch name {
	case "connect":
		var object map[string]interface{}
		if err := d.Decode(&object); err != nil {
			return err
		}
		log.Printf("command: object = %+v", object)
		args = []interface{}{
			object,
		}

	case "releaseStream":
		log.Printf("ignored releaseStream")

	case "createStream":
		var object interface{}
		if err := d.Decode(&object); err != nil {
			return err
		}
		args = []interface{}{
			object,
		}

	case "publish":
		var commandObject interface{}
		if err := d.Decode(&commandObject); err != nil {
			return err
		}
		var publishingName string
		if err := d.Decode(&publishingName); err != nil {
			return err
		}
		var publishingType string
		if err := d.Decode(&publishingType); err != nil {
			return err
		}
		args = []interface{}{
			commandObject,
			publishingName,
			publishingType,
		}

	case "FCPublish":
		log.Printf("ignored FCPublish")

	default:
		return errors.New("Not supported command: " + name)
	}

	*cmdMsg = CommandMessage{
		CommandName:   name,
		TransactionID: transactionID,
		Args:          args,
	}

	return nil
}
