//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

//
type NetConnectionConnect struct {
	Command NetConnectionConnectCommand
}

type NetConnectionConnectCommand struct {
	App      string
	Type     string
	FlashVer string
	TCURL    string
}

func (t *NetConnectionConnect) FromArgs(args ...interface{}) error {
	command := args[0].(map[string]interface{})

	if v, ok := command["app"]; ok {
		if v, ok := v.(string); ok {
			t.Command.App = v
		}
	}

	if v, ok := command["type"]; ok {
		if v, ok := v.(string); ok {
			t.Command.Type = v
		}
	}

	if v, ok := command["flashVer"]; ok {
		if v, ok := v.(string); ok {
			t.Command.FlashVer = v
		}
	}

	if v, ok := command["tcUrl"]; ok {
		if v, ok := v.(string); ok {
			t.Command.TCURL = v
		}
	}

	return nil
}

func (t *NetConnectionConnect) ToArgs() ([]interface{}, error) {
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

func (t *NetConnectionConnectResult) ToArgs() ([]interface{}, error) {
	props := make(map[string]interface{})
	props["fmsVer"] = t.Properties.FMSVer
	props["capabilities"] = t.Properties.Capabilities
	props["mode"] = t.Properties.Mode

	info := make(map[string]interface{})
	info["level"] = t.Information.Level
	info["code"] = t.Information.Code
	type kv struct {
		Key   string `amf0:"ecma"`
		Value interface{}
	}
	data := make([]kv, 0)
	for k, v := range t.Information.Data {
		data = append(data, kv{Key: k, Value: v})
	}
	info["data"] = data
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

func (t *NetConnectionCreateStream) ToArgs() ([]interface{}, error) {
	panic("Not implemented")
}

//
type NetConnectionCreateStreamResult struct {
	StreamID uint32
}

func (t *NetConnectionCreateStreamResult) FromArgs(args ...interface{}) error {
	panic("Not implemented")
}

func (t *NetConnectionCreateStreamResult) ToArgs() ([]interface{}, error) {
	return []interface{}{
		nil, // no command object
		t.StreamID,
	}, nil
}
