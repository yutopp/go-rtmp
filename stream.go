//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/yutopp/go-rtmp/message"
)

// Stream represents a logical message stream
type Stream struct {
	streamID     uint32
	encTy        message.EncodingType
	entryHandler *entryHandler
	streamer     *ChunkStreamer
	cmsg         ChunkMessage
}

func newStream(streamID uint32, entryHandler *entryHandler, streamer *ChunkStreamer) *Stream {
	return &Stream{
		streamID:     streamID,
		encTy:        message.EncodingTypeAMF0, // Default AMF encoding type
		entryHandler: entryHandler,
		streamer:     streamer,
		cmsg: ChunkMessage{
			StreamID: streamID,
		},
	}
}

func (s *Stream) WriteWinAckSize(chunkStreamID int, timestamp uint32, msg *message.WinAckSize) error {
	return s.write(chunkStreamID, timestamp, msg)
}

func (s *Stream) WriteSetPeerBandwidth(chunkStreamID int, timestamp uint32, msg *message.SetPeerBandwidth) error {
	return s.write(chunkStreamID, timestamp, msg)
}

func (s *Stream) WriteUserCtrl(chunkStreamID int, timestamp uint32, msg *message.UserCtrl) error {
	return s.write(chunkStreamID, timestamp, msg)
}

// TODO: return server response
func (s *Stream) Connect() (*message.NetConnectionConnectResult, error) {
	transactionID := int64(1) // Always 1 (7.2.1.1)
	t, err := s.entryHandler.transactions.Create(transactionID)
	if err != nil {
		return nil, err
	}

	chunkStreamID := 3 // TODO: fix
	err = s.writeCommandMessage(
		chunkStreamID, 0, // Timestamp is 0
		"connect",
		transactionID,
		&message.NetConnectionConnect{},
	)
	if err != nil {
		return nil, err
	}

	// TODO: support timeout
	// TODO: check result
	select {
	case <-t.doneCh:
		r := bytes.NewReader(t.body)
		amfDec := message.NewAMFDecoder(r, t.encoding)

		var value message.AMFConvertible
		if err := message.DecodeBodyConnectResult(r, amfDec, &value); err != nil {
			return nil, errors.Wrap(err, "Failed to decode result")
		}
		result := value.(*message.NetConnectionConnectResult)

		return result, nil
	}

	//return nil, errors.New("Failed to get result")
}

func (s *Stream) ReplyConnect(
	chunkStreamID int,
	timestamp uint32,
	body *message.NetConnectionConnectResult,
) error {
	var commandName string
	switch body.Information.Code {
	case message.NetConnectionConnectCodeSuccess, message.NetConnectionConnectCodeClosed:
		commandName = "_result"
	case message.NetConnectionConnectCodeFailed:
		commandName = "_error"
	}

	return s.writeCommandMessage(
		chunkStreamID, timestamp,
		commandName,
		1, // 7.2.1.2, flow.6
		body,
	)
}

func (s *Stream) CreateStream() (*message.NetConnectionCreateStreamResult, error) {
	transactionID := int64(2) // TODO: fix
	t, err := s.entryHandler.transactions.Create(transactionID)
	if err != nil {
		return nil, err
	}

	chunkStreamID := 3 // TODO: fix
	err = s.writeCommandMessage(
		chunkStreamID, 0, // TODO: fix, Timestamp is 0
		"createStream",
		transactionID,
		&message.NetConnectionConnect{},
	)
	if err != nil {
		return nil, err
	}

	// TODO: support timeout
	// TODO: check result
	select {
	case <-t.doneCh:
		r := bytes.NewReader(t.body)
		amfDec := message.NewAMFDecoder(r, t.encoding)

		var value message.AMFConvertible
		if err := message.DecodeBodyCreateStreamResult(r, amfDec, &value); err != nil {
			return nil, errors.Wrap(err, "Failed to decode result")
		}
		result := value.(*message.NetConnectionCreateStreamResult)

		return result, nil
	}

	//return nil, errors.New("Failed to get result")
}

func (s *Stream) ReplyCreateStream(
	chunkStreamID int,
	timestamp uint32,
	transactionID int64,
	body *message.NetConnectionCreateStreamResult,
) error {
	commandName := "_result"
	if body == nil {
		commandName = "_error"
		body = &message.NetConnectionCreateStreamResult{
			StreamID: 0, // TODO: Change to error information object...
		}
	}

	return s.writeCommandMessage(
		chunkStreamID, timestamp,
		commandName,
		transactionID,
		body,
	)
}

func (s *Stream) NotifyStatus(
	chunkStreamID int,
	timestamp uint32,
	body *message.NetStreamOnStatus,
) error {
	return s.writeCommandMessage(
		chunkStreamID, timestamp,
		"onStatus",
		0, // 7.2.2
		body,
	)
}

func (s *Stream) writeCommandMessage(
	chunkStreamID int,
	timestamp uint32,
	commandName string,
	transactionID int64,
	body message.AMFConvertible,
) error {
	buf := new(bytes.Buffer)
	amfEnc := message.NewAMFEncoder(buf, s.encTy)
	if err := message.EncodeBodyAnyValues(amfEnc, body); err != nil {
		return err
	}

	return s.write(chunkStreamID, timestamp, &message.CommandMessage{
		CommandName:   commandName,
		TransactionID: transactionID,
		Encoding:      s.encTy,
		Body:          buf.Bytes(),
	})
}

func (s *Stream) write(chunkStreamID int, timestamp uint32, msg message.Message) error {
	s.cmsg.Message = msg
	return s.streamer.Write(chunkStreamID, timestamp, &s.cmsg)
}

func (s *Stream) handle(chunkStreamID int, timestamp uint32, msg message.Message) error {
	return s.entryHandler.Handle(chunkStreamID, timestamp, msg, s)
}
