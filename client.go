//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"context"
	"net"

	"github.com/pkg/errors"
)

type dialOptions struct {
	dialFunc func(ctx context.Context, network, addr string) (net.Conn, error)
}

func WithContextDialer(dialFunc func(context.Context, string, string) (net.Conn, error)) DialOption {
	return func(o *dialOptions) {
		o.dialFunc = dialFunc
	}
}

type DialOption func(*dialOptions)

func Dial(protocol, addr string, config *ConnConfig, opts ...DialOption) (*ClientConn, error) {
	opt := &dialOptions{
		dialFunc: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	}
	for _, o := range opts {
		o(opt)
	}

	if protocol != "rtmp" {
		return nil, errors.Errorf("Unknown protocol: %s", protocol)
	}

	// TODO: support ctx
	rwc, err := opt.dialFunc(context.Background(), "tcp", addr)
	if err != nil {
		return nil, err
	}

	return newClientConnWithSetup(rwc, config)
}

func DialWithDialer(dialer *net.Dialer, protocol, addr string, config *ConnConfig) (*ClientConn, error) {
	return Dial(protocol, addr, config, WithContextDialer(dialer.DialContext))
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
