//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStreamHandlerChangeState(t *testing.T) {
	rwc := &rwcMock{}
	c := newConn(rwc, nil)
	s := newStream(42, c)

	s.handler.ChangeState(streamStateUnknown)
	assert.Equal(t, s.handler.state, streamStateUnknown)
	assert.Equal(t, s.handler.handler, nil)

	s.handler.ChangeState(streamStateServerNotConnected)
	assert.Equal(t, s.handler.state, streamStateServerNotConnected)
	assert.Equal(t, s.handler.handler, &serverControlNotConnectedHandler{sh: s.handler})

	s.handler.ChangeState(streamStateServerConnected)
	assert.Equal(t, s.handler.state, streamStateServerConnected)
	assert.Equal(t, s.handler.handler, &serverControlConnectedHandler{sh: s.handler})

	s.handler.ChangeState(streamStateServerInactive)
	assert.Equal(t, s.handler.state, streamStateServerInactive)
	assert.Equal(t, s.handler.handler, &serverDataInactiveHandler{sh: s.handler})

	s.handler.ChangeState(streamStateServerPublish)
	assert.Equal(t, s.handler.state, streamStateServerPublish)
	assert.Equal(t, s.handler.handler, &serverDataPublishHandler{sh: s.handler})

	s.handler.ChangeState(streamStateServerPlay)
	assert.Equal(t, s.handler.state, streamStateServerPlay)
	assert.Equal(t, s.handler.handler, &serverDataPlayHandler{sh: s.handler})

	s.handler.ChangeState(streamStateClientNotConnected)
	assert.Equal(t, s.handler.state, streamStateClientNotConnected)
	assert.Equal(t, s.handler.handler, &clientControlNotConnectedHandler{sh: s.handler})
}

func TestStreamStateString(t *testing.T) {
	assert.Equal(t, "<Unknown>", streamStateUnknown.String())
	assert.Equal(t, "NotConnected(Server)", streamStateServerNotConnected.String())
	assert.Equal(t, "Connected(Server)", streamStateServerConnected.String())
	assert.Equal(t, "Inactive(Server)", streamStateServerInactive.String())
	assert.Equal(t, "Publish(Server)", streamStateServerPublish.String())
	assert.Equal(t, "Play(Server)", streamStateServerPlay.String())
	assert.Equal(t, "NotConnected(Client)", streamStateClientNotConnected.String())
	assert.Equal(t, "Connected(Client)", streamStateClientConnected.String())
}
