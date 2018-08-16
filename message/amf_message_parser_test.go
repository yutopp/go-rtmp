//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/yutopp/go-amf0"
	"testing"
)

func TestParseAMFMessageConnect(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := parseAMFMessage(r, d, "connect", &v)
	assert.Nil(t, err)
	assert.Equal(t, &NetConnectionConnect{}, v)
}

func TestParseAMFMessageCreateStream(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := parseAMFMessage(r, d, "createStream", &v)
	assert.Nil(t, err)
	assert.Equal(t, &NetConnectionCreateStream{}, v)
}

func TestParseAMFMessageDeleteStream(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// number: 42
		0x00, 0x40, 0x45, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := parseAMFMessage(r, d, "deleteStream", &v)
	assert.Nil(t, err)
	assert.Equal(t, &NetStreamDeleteStream{
		StreamID: 42,
	}, v)
}

func TestParseAMFMessagePublish(t *testing.T) {
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
	err := parseAMFMessage(r, d, "publish", &v)
	assert.Nil(t, err)
	assert.Equal(t, &NetStreamPublish{
		PublishingName: "abc",
		PublishingType: "def",
	}, v)
}

func TestParseAMFMessagePlay(t *testing.T) {
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
	err := parseAMFMessage(r, d, "play", &v)
	assert.Nil(t, err)
	assert.Equal(t, &NetStreamPlay{
		StreamName: "abc",
		Start:      42,
	}, v)
}

func TestParseAMFMessageReleaseStream(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// string: abc
		0x02, 0x00, 0x03, 0x61, 0x62, 0x63,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := parseAMFMessage(r, d, "releaseStream", &v)
	assert.Nil(t, err)
	assert.Equal(t, &NetConnectionReleaseStream{
		StreamName: "abc",
	}, v)
}

func TestParseAMFMessageFCPublish(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// string: abc
		0x02, 0x00, 0x03, 0x61, 0x62, 0x63,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := parseAMFMessage(r, d, "FCPublish", &v)
	assert.Nil(t, err)
	assert.Equal(t, &NetStreamFCPublish{
		StreamName: "abc",
	}, v)
}

func TestParseAMFMessageFCUnpublish(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// string: abc
		0x02, 0x00, 0x03, 0x61, 0x62, 0x63,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := parseAMFMessage(r, d, "FCUnpublish", &v)
	assert.Nil(t, err)
	assert.Equal(t, &NetStreamFCUnpublish{
		StreamName: "abc",
	}, v)
}

func TestParseAMFMessageAtsetDataFrame(t *testing.T) {
	bin := []byte("payload")
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := parseAMFMessage(r, d, "@setDataFrame", &v)
	assert.Nil(t, err)
	assert.Equal(t, &NetStreamSetDataFrame{
		Payload: bin,
	}, v)
}

func TestParseAMFMessageGetStreamLength(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
		// string: abc
		0x02, 0x00, 0x03, 0x61, 0x62, 0x63,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := parseAMFMessage(r, d, "getStreamLength", &v)
	assert.Nil(t, err)
	assert.Equal(t, &NetStreamGetStreamLength{
		StreamName: "abc",
	}, v)
}

func TestParseAMFMessageNotExist(t *testing.T) {
	bin := []byte{
		// nil
		0x05,
	}
	r := bytes.NewReader(bin)
	d := amf0.NewDecoder(r)

	var v AMFConvertible
	err := parseAMFMessage(r, d, "hogehoge", &v)
	assert.Equal(t, &UnknownAMFParseError{
		Name: "hogehoge",
		Objs: []interface{}{nil},
	}, err)
	assert.Nil(t, v)
}
