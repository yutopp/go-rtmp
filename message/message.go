//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

type TypeID byte

const (
	TypeIDSetChunkSize            TypeID = 1
	TypeIDAbortMessage                   = 2
	TypeIDAck                            = 3
	TypeIDUserCtrl                       = 4
	TypeIDWinAckSize                     = 5
	TypeIDSetPeerBandwidth               = 6
	TypeIDAudioMessage                   = 8
	TypeIDVideoMessage                   = 9
	TypeIDDataMessageAMF3                = 15
	TypeIDSharedObjectMessageAMF3        = 16
	TypeIDCommandMessageAMF3             = 17
	TypeIDDataMessageAMF0                = 18
	TypeIDSharedObjectMessageAMF0        = 16
	TypeIDCommandMessageAMF0             = 20
	TypeIDAggregateMessage               = 22
)

// Message
type Message interface {
	TypeID() TypeID
}

// SetChunkSize (1)
type SetChunkSize struct {
	ChunkSize uint32
}

func (m *SetChunkSize) TypeID() TypeID {
	return TypeIDSetChunkSize
}

// AbortMessage (2)
type AbortMessage struct {
	ChunkStreamID uint32
}

func (m *AbortMessage) TypeID() TypeID {
	return TypeIDAbortMessage
}

// Ack (3)
type Ack struct {
	SequenceNumber uint32
}

func (m *Ack) TypeID() TypeID {
	return TypeIDAck
}

// UserCtrl (4)
type UserCtrl struct {
	Event UserCtrlEvent
}

func (m *UserCtrl) TypeID() TypeID {
	return TypeIDUserCtrl
}

// WinAckSize (5)
type WinAckSize struct {
	Size uint32
}

func (m *WinAckSize) TypeID() TypeID {
	return TypeIDWinAckSize
}

// SetPeerBandwidth (6)
type LimitType uint8

const (
	LimitTypeHard LimitType = 0
	LimitTypeSoft
	LimitTypeDynamic
)

type SetPeerBandwidth struct {
	Size  uint32
	Limit LimitType
}

func (m *SetPeerBandwidth) TypeID() TypeID {
	return TypeIDSetPeerBandwidth
}

// AudioMessage(8)
type AudioMessage struct {
	Payload []byte
}

func (m *AudioMessage) TypeID() TypeID {
	return TypeIDAudioMessage
}

// VideoMessage(9)
type VideoMessage struct {
	Payload []byte
}

func (m *VideoMessage) TypeID() TypeID {
	return TypeIDVideoMessage
}

// DataMessage (15, 18)
type DataMessage struct {
	Name string
	Data interface{}
}

type DataMessageAMF3 struct {
	DataMessage
}

func (m *DataMessageAMF3) TypeID() TypeID {
	return TypeIDDataMessageAMF3
}

type DataMessageAMF0 struct {
	DataMessage
}

func (m *DataMessageAMF0) TypeID() TypeID {
	return TypeIDDataMessageAMF0
}

// SharedObjectMessage (16, 19)
type SharedObjectMessage struct {
}

type SharedObjectMessageAMF3 struct {
	SharedObjectMessage
}

func (m *SharedObjectMessageAMF3) TypeID() TypeID {
	return TypeIDSharedObjectMessageAMF3
}

type SharedObjectMessageAMF0 struct {
	SharedObjectMessage
}

func (m *SharedObjectMessageAMF0) TypeID() TypeID {
	return TypeIDSharedObjectMessageAMF0
}

// CommandMessage (17, 20)
type CommandMessage struct {
	CommandName   string
	TransactionID int64
	Args          []interface{}
}

type CommandMessageAMF3 struct {
	CommandMessage
}

func (m *CommandMessageAMF3) TypeID() TypeID {
	return TypeIDCommandMessageAMF3
}

type CommandMessageAMF0 struct {
	CommandMessage
}

func (m *CommandMessageAMF0) TypeID() TypeID {
	return TypeIDCommandMessageAMF0
}

// AggregateMessage (22)
type AggregateMessage struct {
}

func (m *AggregateMessage) TypeID() TypeID {
	return TypeIDAggregateMessage
}
