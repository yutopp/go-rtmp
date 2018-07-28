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
	FMSVer       string
	Capabilities int // TODO: fix
	Mode         int // TODO: fix
}

type NetConnectionConnectResultInformation struct {
	Level       string // TODO: fix
	Code        string // TODO: fix
	Data        map[string]interface{}
	Application interface{} // TODO: fix
}

func (t *NetConnectionConnectResult) FromArgs(args ...interface{}) error {
	panic("Not implemented")
}

func (t *NetConnectionConnectResult) ToArgs(ty AMFType) ([]interface{}, error) {
	props := make(map[string]interface{})
	props["fmsVer"] = t.Properties.FMSVer
	props["capabilities"] = t.Properties.Capabilities
	props["mode"] = t.Properties.Mode

	info := make(map[string]interface{})
	info["level"] = t.Information.Level
	info["code"] = t.Information.Code

	switch ty {
	case AMFType0:
		data := make(amf0.ECMAArray)
		for k, v := range t.Information.Data {
			data[k] = v
		}
		info["data"] = data
	default:
		panic("Not implemented")
	}

	info["application"] = t.Information.Application

	return []interface{}{
		props,
		info,
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
