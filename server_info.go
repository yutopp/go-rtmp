//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/yutopp/go-rtmp/message"
)

// ServerConnectResultProperties Sent to clients as result when Connect message is received
var serverConnectResultProperties = message.NetConnectionConnectResultProperties{
	FMSVer:       "GO-RTMP/0,0,0,0", // TODO: fix
	Capabilities: 31,                // TODO: fix
	Mode:         1,                 // TODO: fix
}

// ServerConnectResultData Sent to clients as result when Connect message is received
var serverConnectResultData = map[string]interface{}{
	"type":    "go-rtmp",
	"version": "master", // TODO: fix
}

// ServerInfo is an interface to provider server info.
// Users of go-rtmp can obfuscate this information by modifying SInfo field of ConnConfig.
type ServerInfo interface {
	GetServerConnectResultProperties() message.NetConnectionConnectResultProperties
	GetServerConnectResultData() map[string]interface{}
}

type DefaultServerInfo struct {
	ServerConnectResultProperties message.NetConnectionConnectResultProperties
	ServerConnectResultData       map[string]interface{}
}

func NewDefaultServerInfo() *DefaultServerInfo {
	return &DefaultServerInfo{
		// Sent to clients as result when Connect message is received
		ServerConnectResultProperties: message.NetConnectionConnectResultProperties{
			FMSVer:       "GO-RTMP/0,0,0,0", // TODO: fix
			Capabilities: 31,                // TODO: fix
			Mode:         1,                 // TODO: fix
		},
		// Sent to clients as result when Connect message is received
		ServerConnectResultData: map[string]interface{}{
			"type":    "go-rtmp",
			"version": "master", // TODO: fix
		},
	}
}
func (s *DefaultServerInfo) GetServerConnectResultProperties() message.NetConnectionConnectResultProperties {
	return s.ServerConnectResultProperties
}

func (s *DefaultServerInfo) GetServerConnectResultData() map[string]interface{} {
	return s.ServerConnectResultData
}

var defaultServerInfo = NewDefaultServerInfo()
