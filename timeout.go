//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"net"
	"time"
)

type rwcHasTimeout struct {
	conn         net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
	now          func() time.Time // for mock
}

func (rwc *rwcHasTimeout) Read(b []byte) (int, error) {
	if err := rwc.conn.SetReadDeadline(rwc.calcDeadline(rwc.readTimeout)); err != nil {
		return 0, err
	}

	return rwc.conn.Read(b)
}

func (rwc *rwcHasTimeout) Write(b []byte) (int, error) {
	if err := rwc.conn.SetWriteDeadline(rwc.calcDeadline(rwc.writeTimeout)); err != nil {
		return 0, err
	}

	return rwc.conn.Write(b)
}

func (rwc *rwcHasTimeout) Close() error {
	return rwc.conn.Close()
}

func (rwc *rwcHasTimeout) calcDeadline(d time.Duration) time.Time {
	if d == 0 {
		return time.Time{} // zero value means infinity
	}

	return rwc.now().Add(d)
}
