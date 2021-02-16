//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"net"

	"github.com/pkg/errors"
)

func Dial(protocol, addr string, config *ConnConfig) (*ClientConn, error) {
	return DialWithDialer(&net.Dialer{}, protocol, addr, config)
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

func makeValidAddr(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if err, ok := err.(*net.AddrError); ok && err.Err == "missing port in address" {
			return makeValidAddr(addr + ":1935") // Default RTMP port
		}
		return "", err
	}
	return net.JoinHostPort(host, port), nil
}
