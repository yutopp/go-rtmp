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

func (t *NetStreamPublish) ToArgs(ty EncodingType) ([]interface{}, error) {
	panic("Not implemented")
}

//
type NetStreamPlay struct {
	CommandObject interface{}
	StreamName    string
	Start         int64
}

func (t *NetStreamPlay) FromArgs(args ...interface{}) error {
	//command := args[0] // will be nil
	t.StreamName = args[1].(string)
	t.Start = args[2].(int64)

	return nil
}

func (t *NetStreamPlay) ToArgs(ty EncodingType) ([]interface{}, error) {
	panic("Not implemented")
}

//
type NetStreamOnStatusLevel string

const (
	NetStreamOnStatusLevelStatus NetStreamOnStatusLevel = "status"
	NetStreamOnStatusLevelError  NetStreamOnStatusLevel = "error"
)

type NetStreamOnStatusCode string

const (
	NetStreamOnStatusCodeConnectSuccess      NetStreamOnStatusCode = "NetStream.Connect.Success"
	NetStreamOnStatusCodeConnectFailed       NetStreamOnStatusCode = "NetStream.Connect.Failed"
	NetStreamOnStatusCodeMuticastStreamReset NetStreamOnStatusCode = "NetStream.MulticastStream.Reset"
	NetStreamOnStatusCodePlayStart           NetStreamOnStatusCode = "NetStream.Play.Start"
	NetStreamOnStatusCodePlayFailed          NetStreamOnStatusCode = "NetStream.Play.Failed"
	NetStreamOnStatusCodePlayComplete        NetStreamOnStatusCode = "NetStream.Play.Complete"
	NetStreamOnStatusCodePublishBadName      NetStreamOnStatusCode = "NetStream.Publish.BadName"
	NetStreamOnStatusCodePublishFailed       NetStreamOnStatusCode = "NetStream.Publish.Failed"
	NetStreamOnStatusCodePublishStart        NetStreamOnStatusCode = "NetStream.Publish.Start"
	NetStreamOnStatusCodeUnpublishSuccess    NetStreamOnStatusCode = "NetStream.Unpublish.Success"
)

type NetStreamOnStatus struct {
	InfoObject NetStreamOnStatusInfoObject
}

type NetStreamOnStatusInfoObject struct {
	Level       NetStreamOnStatusLevel
	Code        NetStreamOnStatusCode
	Description string
}

func (t *NetStreamOnStatus) FromArgs(args ...interface{}) error {
	panic("Not implemented")
}

func (t *NetStreamOnStatus) ToArgs(ty EncodingType) ([]interface{}, error) {
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
type NetStreamDeleteStream struct {
	StreamID uint32
}

func (t *NetStreamDeleteStream) FromArgs(args ...interface{}) error {
	// args[0] is unknown, ignore
	t.StreamID = args[1].(uint32)

	return nil
}

func (t *NetStreamDeleteStream) ToArgs(ty EncodingType) ([]interface{}, error) {
	panic("Not implemented")
}

//
type NetStreamFCPublish struct {
	StreamName string
}

func (t *NetStreamFCPublish) FromArgs(args ...interface{}) error {
	// args[0] is unknown, ignore
	t.StreamName = args[1].(string)

	return nil
}

func (t *NetStreamFCPublish) ToArgs(ty EncodingType) ([]interface{}, error) {
	return []interface{}{
		nil, // no command object
		t.StreamName,
	}, nil
}

//
type NetStreamFCUnpublish struct {
	StreamName string
}

func (t *NetStreamFCUnpublish) FromArgs(args ...interface{}) error {
	// args[0] is unknown, ignore
	t.StreamName = args[1].(string)

	return nil
}

func (t *NetStreamFCUnpublish) ToArgs(ty EncodingType) ([]interface{}, error) {
	return []interface{}{
		nil, // no command object
		t.StreamName,
	}, nil
}

//
type NetStreamSetDataFrame struct {
	Payload []byte
}

func (t *NetStreamSetDataFrame) FromArgs(args ...interface{}) error {
	t.Payload = args[0].([]byte)

	return nil
}

func (t *NetStreamSetDataFrame) ToArgs(ty EncodingType) ([]interface{}, error) {
	panic("Not implemented")
}

//
type NetStreamGetStreamLength struct {
	StreamName string
}

func (t *NetStreamGetStreamLength) FromArgs(args ...interface{}) error {
	// args[0] is unknown, ignore
	t.StreamName = args[1].(string)

	return nil
}

func (t *NetStreamGetStreamLength) ToArgs(ty EncodingType) ([]interface{}, error) {
	return []interface{}{
		nil, // no command object
		t.StreamName,
	}, nil
}
