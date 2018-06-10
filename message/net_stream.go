//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

//
type NetStreamPublish struct {
	CommandObject  interface{}
	PublishingName string
	PublishingType string
}

func (t *NetStreamPublish) FromArgs(args ...interface{}) error {
	//command := args[0] // will be nil
	t.PublishingName = args[1].(string)
	t.PublishingType = args[2].(string)

	return nil
}

func (t *NetStreamPublish) ToArgs() ([]interface{}, error) {
	panic("Not implemented")
}

//
type NetStreamOnStatus struct {
	InfoObject NetStreamOnStatusInfoObject
}

type NetStreamOnStatusInfoObject struct {
	Level       string // TODO: fix
	Code        string // TODO: fix
	Description string
}

func (t *NetStreamOnStatus) FromArgs(args ...interface{}) error {
	panic("Not implemented")
}

func (t *NetStreamOnStatus) ToArgs() ([]interface{}, error) {
	info := make(map[string]interface{})
	info["level"] = t.InfoObject.Level
	info["code"] = t.InfoObject.Code
	info["description"] = t.InfoObject.Description

	return []interface{}{
		nil, // Always nil
		info,
	}, nil
}

//
type NetStreamOnMetaData struct {
	MetaData map[string]interface{} // TODO: to more detailed data
}
