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

//
type NetConnectionConnect struct {
	Command NetConnectionConnectCommand
}

type NetConnectionConnectCommand struct {
	App           string `mapstructure:"app"`
	Type          string `mapstructure:"type"`
	FlashVer      string `mapstructure:"flashVer"`
	TCURL         string `mapstructure:"tcUrl"`
	Fpad          bool   `mapstructure:"fpad"`
	Capabilities  int    `mapstructure:"capabilities"`
	AudioCodecs   int    `mapstructure:"audioCodecs"`
	VideoCodecs   int    `mapstructure:"videoCodecs"`
	VideoFunction int    `mapstructure:"videoFunction"`
}

func (t *NetConnectionConnect) FromArgs(args ...interface{}) error {
	command := args[0].(map[string]interface{})
	if err := mapstructure.Decode(command, &t.Command); err != nil {
		return errors.Wrapf(err, "Failed to mapping NetConnectionConnect")
	}

	return nil
}

func (t *NetConnectionConnect) ToArgs(ty AMFType) ([]interface{}, error) {
	panic("Not implemented")
}

//
type NetConnectionConnectResult struct {
	Properties  NetConnectionConnectResultProperties
	Information NetConnectionConnectResultInformation
}

type NetConnectionConnectResultProperties struct {
	FMSVer       string `amf0:"fmsVer"`       // TODO: fix
	Capabilities int    `amf0:"capabilities"` // TODO: fix
	Mode         int    `amf0:"mode"`         // TODO: fix
}

type NetConnectionConnectResultInformation struct {
	Level       string         `amf0:"level"` // TODO: fix
	Code        string         `amf0:"code"`  // TODO: fix
	Description string         `amf0:"description"`
	Data        amf0.ECMAArray `amf0:"data"`
}

func (t *NetConnectionConnectResult) FromArgs(args ...interface{}) error {
	panic("Not implemented")
}

func (t *NetConnectionConnectResult) ToArgs(ty AMFType) ([]interface{}, error) {
	return []interface{}{
		t.Properties,
		t.Information,
	}, nil
}

//
type NetConnectionCreateStream struct {
}

func (t *NetConnectionCreateStream) FromArgs(args ...interface{}) error {
	// args[0] // Will be nil...
	return nil
}

func (t *NetConnectionCreateStream) ToArgs(ty AMFType) ([]interface{}, error) {
	panic("Not implemented")
}

//
type NetConnectionCreateStreamResult struct {
	StreamID uint32
}

func (t *NetConnectionCreateStreamResult) FromArgs(args ...interface{}) error {
	panic("Not implemented")
}

func (t *NetConnectionCreateStreamResult) ToArgs(ty AMFType) ([]interface{}, error) {
	return []interface{}{
		nil, // no command object
		t.StreamID,
	}, nil
}

//
type NetConnectionReleaseStream struct {
	StreamName string
}

func (t *NetConnectionReleaseStream) FromArgs(args ...interface{}) error {
	// args[0] is unknown, ignore
	t.StreamName = args[1].(string)

	return nil
}

func (t *NetConnectionReleaseStream) ToArgs(ty AMFType) ([]interface{}, error) {
	return []interface{}{
		nil, // no command object
		t.StreamName,
	}, nil
}
