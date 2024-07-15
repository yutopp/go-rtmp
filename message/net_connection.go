//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/yutopp/go-amf0"
)

type NetConnectionConnectCode string

const (
	NetConnectionConnectCodeSuccess NetConnectionConnectCode = "NetConnection.Connect.Success"
	NetConnectionConnectCodeFailed  NetConnectionConnectCode = "NetConnection.Connect.Failed"
	NetConnectionConnectCodeClosed  NetConnectionConnectCode = "NetConnection.Connect.Closed"
)

type NetConnectionConnect struct {
	Command NetConnectionConnectCommand
}

type NetConnectionConnectCommand struct {
	App            string       `mapstructure:"app" amf0:"app"`
	Type           string       `mapstructure:"type" amf0:"type"`
	FlashVer       string       `mapstructure:"flashVer" amf0:"flashVer"`
	TCURL          string       `mapstructure:"tcUrl" amf0:"tcUrl"`
	Fpad           bool         `mapstructure:"fpad" amf0:"fpad"`
	Capabilities   int          `mapstructure:"capabilities" amf0:"capabilities"`
	AudioCodecs    int          `mapstructure:"audioCodecs" amf0:"audioCodecs"`
	VideoCodecs    int          `mapstructure:"videoCodecs" amf0:"videoCodecs"`
	VideoFunction  int          `mapstructure:"videoFunction" amf0:"videoFunction"`
	ObjectEncoding EncodingType `mapstructure:"objectEncoding" amf0:"objectEncoding"`
}

func (t *NetConnectionConnect) FromArgs(args ...interface{}) error {
	command, ok := args[0].(map[string]interface{})
	if !ok {
		return errors.Errorf("expect map[string]interface{} at arg[0], but got %T", args[0])
	}
	if err := mapstructure.Decode(command, &t.Command); err != nil {
		return errors.Wrapf(err, "failed to mapping arg[0] to NetConnectionConnectCommand")
	}

	return nil
}

func (t *NetConnectionConnect) ToArgs(ty EncodingType) ([]interface{}, error) {
	return []interface{}{
		t.Command,
	}, nil
}

type NetConnectionConnectResult struct {
	Properties  NetConnectionConnectResultProperties
	Information NetConnectionConnectResultInformation
}

type NetConnectionConnectResultProperties struct {
	FMSVer       string `mapstructure:"fmsVer" amf0:"fmsVer"`             // TODO: fix
	Capabilities int    `mapstructure:"capabilities" amf0:"capabilities"` // TODO: fix
	Mode         int    `mapstructure:"mode" amf0:"mode"`                 // TODO: fix
}

type NetConnectionConnectResultInformation struct {
	Level       string                   `mapstructure:"level" amf0:"level"` // TODO: fix
	Code        NetConnectionConnectCode `mapstructure:"code" amf0:"code"`
	Description string                   `mapstructure:"description" amf0:"description"`
	Data        amf0.ECMAArray           `mapstructure:"data" amf0:"data"`
}

func (t *NetConnectionConnectResult) FromArgs(args ...interface{}) error {
	properties, ok := args[0].(map[string]interface{})
	if !ok {
		return errors.Errorf("expect map[string]interface{} at arg[0], but got %T", args[0])
	}
	if err := mapstructure.Decode(properties, &t.Properties); err != nil {
		return errors.Wrapf(err, "failed to mapping arg[0] to NetConnectionConnectResultProperties")
	}

	information, ok := args[1].(map[string]interface{})
	if !ok {
		return errors.Errorf("expect map[string]interface{} at arg[1], but got %T", args[1])
	}
	if err := mapstructure.Decode(information, &t.Information); err != nil {
		return errors.Wrapf(err, "failed to mapping arg[1] to NetConnectionConnectResultInformation")
	}

	return nil
}

func (t *NetConnectionConnectResult) ToArgs(ty EncodingType) ([]interface{}, error) {
	return []interface{}{
		t.Properties,
		t.Information,
	}, nil
}

type NetConnectionCreateStream struct {
}

func (t *NetConnectionCreateStream) FromArgs(args ...interface{}) error {
	// args[0] // Will be nil...
	return nil
}

func (t *NetConnectionCreateStream) ToArgs(ty EncodingType) ([]interface{}, error) {
	return []interface{}{
		nil, // Just null
	}, nil
}

// TODO: fix for error messages
type NetConnectionCreateStreamResult struct {
	StreamID uint32
}

func (t *NetConnectionCreateStreamResult) FromArgs(args ...interface{}) error {
	// args[0] is unknown, ignore

	streamID, ok := args[1].(uint32)
	if !ok {
		return errors.Errorf("expect uint32 at arg[1], but got %T", args[1])
	}
	t.StreamID = streamID

	return nil
}

func (t *NetConnectionCreateStreamResult) ToArgs(ty EncodingType) ([]interface{}, error) {
	return []interface{}{
		nil, // no command object
		t.StreamID,
	}, nil
}

type NetConnectionReleaseStream struct {
	StreamName string
}

func (t *NetConnectionReleaseStream) FromArgs(args ...interface{}) error {
	// args[0] is unknown, ignore

	streamName, ok := args[1].(string)
	if !ok {
		return errors.Errorf("expect string at arg[1], but got %T", args[1])
	}
	t.StreamName = streamName

	return nil
}

func (t *NetConnectionReleaseStream) ToArgs(ty EncodingType) ([]interface{}, error) {
	return []interface{}{
		nil, // no command object
		t.StreamName,
	}, nil
}
