//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"crypto/tls"
	"net"

	"github.com/pkg/errors"
)

func Dial(protocol, addr string, config *ConnConfig) (*ClientConn, error) {
	return DialWithDialer(&net.Dialer{}, protocol, addr, config)
}

func TLSDial(protocol, addr string, config *ConnConfig, tlsConfig *tls.Config) (*ClientConn, error) {
	return DialWithTLSDialer(&tls.Dialer{
		NetDialer: &net.Dialer{},
		Config:    tlsConfig,
	}, protocol, addr, config)
}

func DialWithDialer(dialer *net.Dialer, protocol, addr string, config *ConnConfig) (*ClientConn, error) {
	if protocol != "rtmp" {
		return nil, errors.Errorf("Unknown protocol: %s", protocol)
	}

	rwc, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return newClientConnWithSetup(rwc, config)
}

func DialWithTLSDialer(dialer *tls.Dialer, protocol, addr string, config *ConnConfig) (*ClientConn, error) {
	if protocol != "rtmps" {
		return nil, errors.Errorf("Unknown protocol: %s", protocol)
	}

	rwc, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return newClientConnWithSetup(rwc, config)
}
