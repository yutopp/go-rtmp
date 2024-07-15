//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"context"
	"crypto/tls"
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

// DialContext dials a connection to the specified address.
// The protocol must be "rtmp" or "rtmps".
func DialContext(ctx context.Context, protocol, addr string, config *ConnConfig, opts ...DialOption) (*ClientConn, error) {
	opt := &dialOptions{
		// default dialer
		dialFunc: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	}
	for _, o := range opts {
		o(opt)
	}

	if protocol != "rtmp" && protocol != "rtmps" {
		return nil, errors.Errorf("unknown protocol: %s", protocol)
	}

	rwc, err := opt.dialFunc(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	return newClientConnWithSetup(rwc, config)
}

// Dial dials a connection to the specified address.
func Dial(protocol, addr string, config *ConnConfig, opts ...DialOption) (*ClientConn, error) {
	return DialContext(context.Background(), protocol, addr, config, opts...)
}

// DialWithDialer dials a connection to the specified address with the specified dialer.
func DialWithDialer(dialer *net.Dialer, protocol, addr string, config *ConnConfig) (*ClientConn, error) {
	return Dial(protocol, addr, config, WithContextDialer(dialer.DialContext))
}

// TLSDialContext dials a connection to the specified address with TLS.
func TLSDialContext(ctx context.Context, protocol, addr string, config *ConnConfig, tlsConfig *tls.Config, opts ...DialOption) (*ClientConn, error) {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{},
		Config:    tlsConfig,
	}
	opts = append([]DialOption{WithContextDialer(dialer.DialContext)}, opts...)
	return DialContext(ctx, protocol, addr, config, opts...)
}

// TLSDial dials a connection to the specified address with TLS.
func TLSDial(protocol, addr string, config *ConnConfig, tlsConfig *tls.Config, opts ...DialOption) (*ClientConn, error) {
	return TLSDialContext(context.Background(), protocol, addr, config, tlsConfig, opts...)
}

// DialWithTLSDialer dials a connection to the specified address with the specified TLS dialer.
func DialWithTLSDialer(dialer *tls.Dialer, protocol, addr string, config *ConnConfig) (*ClientConn, error) {
	return Dial(protocol, addr, config, WithContextDialer(dialer.DialContext))
}
