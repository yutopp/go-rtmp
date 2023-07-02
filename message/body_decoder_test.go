//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yutopp/go-amf0"
)

func TestDecodeDataMessageAtsetDataFrame(t *testing.T) {
	bin := []byte("payload")
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := DataBodyDecoderFor("@setDataFrame")(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetStreamSetDataFrame{
		Payload: bin,
	}, v)
}

func TestDecodeDataMessageUnknown(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := DataBodyDecoderFor("hogehoge")(r, d, &v)
	require.Equal(t, &UnknownDataBodyDecodeError{
		Name: "hogehoge",
		Objs: []interface{}{nil},
	}, err)
	require.Nil(t, v)
}

func TestDecodeCmdMessageConnect(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("connect", 1)(r, d, &v) // Transaction is always 1 (7.2.1.1)
	require.Nil(t, err)
	require.Equal(t, &NetConnectionConnect{}, v)
}

func TestDecodeCmdMessageCreateStream(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("createStream", 42)(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetConnectionCreateStream{}, v)
}

func TestDecodeCmdMessageDeleteStream(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// number: 42
		0x00, 0x40, 0x45, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("deleteStream", 42)(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetStreamDeleteStream{
		StreamID: 42,
	}, v)
}

func TestDecodeCmdMessagePublish(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// string: abc
		0x02, 0x00, 0x03, 0x61, 0x62, 0x63,
		// string: def
		0x02, 0x00, 0x03, 0x64, 0x65, 0x66,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("publish", 42)(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetStreamPublish{
		PublishingName: "abc",
		PublishingType: "def",
	}, v)
}

func TestDecodeCmdMessagePlay(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// string: abc
		0x02, 0x00, 0x03, 0x61, 0x62, 0x63,
		// number: 42
		0x00, 0x40, 0x45, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("play", 42)(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetStreamPlay{
		StreamName: "abc",
		Start:      42,
	}, v)
}

func TestDecodeCmdMessageReleaseStream(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// string: abc
		0x02, 0x00, 0x03, 0x61, 0x62, 0x63,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("releaseStream", 42)(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetConnectionReleaseStream{
		StreamName: "abc",
	}, v)
}

func TestDecodeCmdMessageFCPublish(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// string: abc
		0x02, 0x00, 0x03, 0x61, 0x62, 0x63,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("FCPublish", 42)(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetStreamFCPublish{
		StreamName: "abc",
	}, v)
}

func TestDecodeCmdMessageFCUnpublish(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// string: abc
		0x02, 0x00, 0x03, 0x61, 0x62, 0x63,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("FCUnpublish", 42)(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetStreamFCUnpublish{
		StreamName: "abc",
	}, v)
}

func TestDecodeCmdMessageGetStreamLength(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// string: abc
		0x02, 0x00, 0x03, 0x61, 0x62, 0x63,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("getStreamLength", 42)(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetStreamGetStreamLength{
		StreamName: "abc",
	}, v)
}

func TestDecodeCmdMessagePing(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("ping", 42)(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetStreamPing{}, v)
}

func TestDecodeCmdMessageCloseStream(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("closeStream", 42)(r, d, &v)
	require.Nil(t, err)
	require.Equal(t, &NetStreamCloseStream{}, v)
}

func TestDecodeCmdMessageUnknown(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := CmdBodyDecoderFor("hogehoge", 42)(r, d, &v)
	require.Equal(t, &UnknownCommandBodyDecodeError{
		Name:          "hogehoge",
		TransactionID: 42,
		Objs:          []interface{}{nil},
	}, err)
	require.Nil(t, v)
}
