//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamHandlerChangeState(t *testing.T) {
	rwc := &rwcMock{}
	c := newConn(rwc, nil)
	s := newStream(42, c)

	s.handler.ChangeState(streamStateUnknown)
	require.Equal(t, s.handler.state, streamStateUnknown)
	require.Equal(t, s.handler.handler, nil)

	s.handler.ChangeState(streamStateServerNotConnected)
	require.Equal(t, s.handler.state, streamStateServerNotConnected)
	require.Equal(t, s.handler.handler, &serverControlNotConnectedHandler{sh: s.handler})

	s.handler.ChangeState(streamStateServerConnected)
	require.Equal(t, s.handler.state, streamStateServerConnected)
	require.Equal(t, s.handler.handler, &serverControlConnectedHandler{sh: s.handler})

	s.handler.ChangeState(streamStateServerInactive)
	require.Equal(t, s.handler.state, streamStateServerInactive)
	require.Equal(t, s.handler.handler, &serverDataInactiveHandler{sh: s.handler})

	s.handler.ChangeState(streamStateServerPublish)
	require.Equal(t, s.handler.state, streamStateServerPublish)
	require.Equal(t, s.handler.handler, &serverDataPublishHandler{sh: s.handler})

	s.handler.ChangeState(streamStateServerPlay)
	require.Equal(t, s.handler.state, streamStateServerPlay)
	require.Equal(t, s.handler.handler, &serverDataPlayHandler{sh: s.handler})

	s.handler.ChangeState(streamStateClientNotConnected)
	require.Equal(t, s.handler.state, streamStateClientNotConnected)
	require.Equal(t, s.handler.handler, &clientControlNotConnectedHandler{sh: s.handler})
}

func TestStreamStateString(t *testing.T) {
	require.Equal(t, "<Unknown>", streamStateUnknown.String())
	require.Equal(t, "NotConnected(Server)", streamStateServerNotConnected.String())
	require.Equal(t, "Connected(Server)", streamStateServerConnected.String())
	require.Equal(t, "Inactive(Server)", streamStateServerInactive.String())
	require.Equal(t, "Publish(Server)", streamStateServerPublish.String())
	require.Equal(t, "Play(Server)", streamStateServerPlay.String())
	require.Equal(t, "NotConnected(Client)", streamStateClientNotConnected.String())
	require.Equal(t, "Connected(Client)", streamStateClientConnected.String())
}
