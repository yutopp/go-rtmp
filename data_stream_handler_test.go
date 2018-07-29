//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/sirupsen/logrus"
	"testing"

	"github.com/yutopp/go-rtmp/message"
)

func BenchmarkHandlePublisherVideoMessage(b *testing.B) {
	h := &dataStreamHandler{
		state:   dataStreamStateHasPublisher,
		handler: &DefaultHandler{},
		logger:  logrus.StandardLogger(),
	}

	chunkStreamID := 0
	timestamp := uint32(0)
	msg := &message.VideoMessage{}
	stream := &Stream{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Handle(chunkStreamID, timestamp, msg, stream)
	}
}
