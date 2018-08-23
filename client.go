//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"net"
	"net/url"
)

func Dial(urlBase string, config *ConnConfig) (*Conn, error) {
	return DialWithDialer(&net.Dialer{}, urlBase, config)
}

func DialWithDialer(dialer *net.Dialer, urlBase string, config *ConnConfig) (*Conn, error) {
	u, err := url.Parse(urlBase)
	if err != nil {
		return nil, err
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "1935" // RTMP default port
	}

	rwc, err := dialer.Dial("tcp", host+":"+port)
	if err != nil {
		return nil, err
	}

	return newConn(rwc, config), nil
}
