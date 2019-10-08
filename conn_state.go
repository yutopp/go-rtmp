//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"math"

	"github.com/pkg/errors"
	"github.com/yutopp/go-rtmp/message"
)

const DefaultChunkSize = 128
const MaxChunkSize = 0xffffff // 5.4.1

type StreamControlState struct {
	chunkSize           uint32
	ackWindowSize       int32
	bandwidthWindowSize int32
	bandwidthLimitType  message.LimitType

	config *StreamControlStateConfig
}

type StreamControlStateConfig struct {
	DefaultChunkSize uint32
	MaxChunkSize     uint32
	MaxChunkStreams  int

	DefaultAckWindowSize int32
	MaxAckWindowSize     int32

	DefaultBandwidthWindowSize int32
	DefaultBandwidthLimitType  message.LimitType
	MaxBandwidthWindowSize     int32

	MaxMessageSize    uint32
	MaxMessageStreams int
}

func (cb *StreamControlStateConfig) normalize() *StreamControlStateConfig {
	c := StreamControlStateConfig(*cb)

	// chunks

	if c.DefaultChunkSize == 0 {
		c.DefaultChunkSize = DefaultChunkSize
	}

	if c.MaxChunkSize == 0 {
		c.MaxChunkSize = MaxChunkSize
	}

	if c.MaxChunkStreams == 0 {
		c.MaxChunkStreams = math.MaxUint32
	}

	// ack

	if c.DefaultAckWindowSize == 0 {
		c.DefaultAckWindowSize = math.MaxInt32
	}

	if c.MaxAckWindowSize == 0 {
		c.MaxAckWindowSize = math.MaxInt32
	}

	// bandwidth

	if c.DefaultBandwidthWindowSize == 0 {
		c.DefaultBandwidthWindowSize = math.MaxInt32
	}

	if c.MaxBandwidthWindowSize == 0 {
		c.MaxBandwidthWindowSize = math.MaxInt32
	}

	// message

	if c.MaxMessageStreams == 0 {
		c.MaxMessageStreams = math.MaxUint32
	}

	if c.MaxMessageSize == 0 {
		c.MaxMessageSize = MaxChunkSize // as same as chunk size
	}

	return &c
}

var defaultStreamControlStateConfig = (&StreamControlStateConfig{}).normalize()

func NewStreamControlState(config *StreamControlStateConfig) *StreamControlState {
	if config == nil {
		config = defaultStreamControlStateConfig
	}

	return &StreamControlState{
		chunkSize:           config.DefaultChunkSize,
		ackWindowSize:       config.DefaultAckWindowSize,
		bandwidthWindowSize: config.DefaultBandwidthWindowSize,
		bandwidthLimitType:  config.DefaultBandwidthLimitType,

		config: config,
	}
}

func (s *StreamControlState) ChunkSize() uint32 {
	return s.chunkSize
}

func (s *StreamControlState) SetChunkSize(chunkSize uint32) error {
	if chunkSize > MaxChunkSize {
		chunkSize = MaxChunkSize
	}

	if chunkSize > s.config.MaxChunkSize {
		return errors.Errorf("Exceeded configured max chunk size: Limit = %d, Value = %d", s.config.MaxChunkSize, chunkSize)
	}

	s.chunkSize = chunkSize

	return nil
}

func (s *StreamControlState) AckWindowSize() int32 {
	return s.ackWindowSize
}

func (s *StreamControlState) SetAckWindowSize(ackWindowSize int32) error {
	if ackWindowSize > s.config.MaxAckWindowSize {
		return errors.Errorf("Exceeded configured max ack window size: Limit = %d, Value = %d", s.config.MaxAckWindowSize, ackWindowSize)
	}

	s.ackWindowSize = ackWindowSize

	return nil
}

func (s *StreamControlState) BandwidthWindowSize() int32 {
	return s.bandwidthWindowSize
}

func (s *StreamControlState) BandwidthLimitType() message.LimitType {
	return s.bandwidthLimitType
}
