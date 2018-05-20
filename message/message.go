//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

type TypeID byte

const (
	TypeID_                  TypeID = 0
	TypeIDUserCtrl                  = 4
	TypeIDCtrlWinAckSize            = 5
	TypeIDSetPeerBandwidth          = 6
	TypeIDAudioMessage              = 8
	TypeIDVideoMessage              = 9
	TypeIDDataMessageAMF0           = 18
	TypeIDCommandMessageAMF0        = 20
)

// Message
type Message interface {
	TypeID() TypeID
}

// UserCtrl (4)
type UserCtrl struct {
	Event UserCtrlEvent
}

func (m *UserCtrl) TypeID() TypeID {
	return TypeIDUserCtrl
}

// CtrlWinAckSize (5)
type CtrlWinAckSize struct {
	Size uint32
}

func (m *CtrlWinAckSize) TypeID() TypeID {
	return TypeIDCtrlWinAckSize
}

// SetPeerBandwidth (6)
type SetPeerBandwidth struct {
	Size  uint32
	Limit uint8
}

func (m *SetPeerBandwidth) TypeID() TypeID {
	return TypeIDSetPeerBandwidth
}

//
type AudioMessage struct {
	Payload []byte
}

func (m *AudioMessage) TypeID() TypeID {
	return TypeIDAudioMessage
}

//
type VideoMessage struct {
	Payload []byte
}

func (m *VideoMessage) TypeID() TypeID {
	return TypeIDVideoMessage
}

//
type DataMessageAMF0 struct {
	Name string
	Data interface{}
}

func (m *DataMessageAMF0) TypeID() TypeID {
	return TypeIDDataMessageAMF0
}

//
type CommandMessageAMF0 struct {
	CommandName   string
	TransactionID int64
	Command       interface{}
}

func (m *CommandMessageAMF0) TypeID() TypeID {
	return TypeIDCommandMessageAMF0
}
