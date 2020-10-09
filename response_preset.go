//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/edgeware/go-rtmp/message"
)

// ResponsePreset is an interface to provider server info.
// Users of go-rtmp can obfuscate this information by modifying RPreset field of ConnConfig.
type ResponsePreset interface {
	GetServerConnectResultProperties() message.NetConnectionConnectResultProperties
	GetServerConnectResultData() map[string]interface{}
}

// DefaultResponsePreset gives a default ServerInfo.
type DefaultResponsePreset struct {
	ServerConnectResultProperties message.NetConnectionConnectResultProperties
	ServerConnectResultData       map[string]interface{}
}

// NewDefaultResponsePreset gives an instance of DefaultResponsePreset
func NewDefaultResponsePreset() *DefaultResponsePreset {
	return &DefaultResponsePreset{
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

// GetServerConnectResultProperties returns ServerConnectResultProperties.
func (r *DefaultResponsePreset) GetServerConnectResultProperties() message.NetConnectionConnectResultProperties {
	return r.ServerConnectResultProperties
}

// GetServerConnectResultData returns ServerConnectResultData.
func (r *DefaultResponsePreset) GetServerConnectResultData() map[string]interface{} {
	return r.ServerConnectResultData
}

var defaultResponsePreset ResponsePreset = NewDefaultResponsePreset()
