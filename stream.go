//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"bytes"
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/yutopp/go-rtmp/message"
)

// Stream represents a logical message stream
type Stream struct {
	streamID     uint32
	encTy        message.EncodingType
	transactions *transactions
	handler      *streamHandler
	cmsg         ChunkMessage

	conn *Conn
}

func newStream(streamID uint32, conn *Conn) *Stream {
	s := &Stream{
		streamID:     streamID,
		encTy:        message.EncodingTypeAMF0, // Default AMF encoding type
		transactions: newTransactions(),
		cmsg: ChunkMessage{
			StreamID: streamID,
		},

		conn: conn,
	}
	s.handler = newStreamHandler(s)

	return s
}

func (s *Stream) StreamID() uint32 {
	return s.streamID
}

func (s *Stream) WriteWinAckSize(chunkStreamID int, timestamp uint32, msg *message.WinAckSize) error {
	return s.Write(chunkStreamID, timestamp, msg)
}

func (s *Stream) WriteSetPeerBandwidth(chunkStreamID int, timestamp uint32, msg *message.SetPeerBandwidth) error {
	return s.Write(chunkStreamID, timestamp, msg)
}

func (s *Stream) WriteUserCtrl(chunkStreamID int, timestamp uint32, msg *message.UserCtrl) error {
	return s.Write(chunkStreamID, timestamp, msg)
}

func (s *Stream) Connect(
	body *message.NetConnectionConnect,
) (*message.NetConnectionConnectResult, error) {
	transactionID := int64(1) // Always 1 (7.2.1.1)
	t, err := s.transactions.Create(transactionID)
	if err != nil {
		return nil, err
	}

	if body == nil {
		body = &message.NetConnectionConnect{}
	}

	chunkStreamID := 3 // TODO: fix
	err = s.writeCommandMessage(
		chunkStreamID, 0, // Timestamp is 0
		"connect",
		transactionID,
		body,
	)
	if err != nil {
		return nil, err
	}

	// TODO: support timeout
	timeoutCtx := context.TODO()
	select {
	case <-timeoutCtx.Done():
		return nil, timeoutCtx.Err()
	case <-t.doneCh:
		amfDec := message.NewAMFDecoder(t.body, t.encoding)

		var value message.AMFConvertible
		if err := message.DecodeBodyConnectResult(t.body, amfDec, &value); err != nil {
			return nil, errors.Wrap(err, "Failed to decode result")
		}
		result := value.(*message.NetConnectionConnectResult)

		if t.commandName == "_error" {
			return nil, &ConnectRejectedError{
				TransactionID: transactionID,
				Result:        result,
			}
		}

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

func (s *Stream) CreateStream(body *message.NetConnectionCreateStream, chunkSize uint32) (*message.NetConnectionCreateStreamResult, error) {
	oldChunkSize := s.conn.streamer.selfState.chunkSize
	if chunkSize > 0 && chunkSize != oldChunkSize {
		logrus.Infof("Changing chunkSize %d->%d", oldChunkSize, chunkSize)
		s.conn.streamer.selfState.chunkSize = chunkSize
		err := s.WriteSetChunkSize(chunkSize)
		if err != nil {
			return nil, err
		}
	}

	transactionID := int64(2) // TODO: fix
	t, err := s.transactions.Create(transactionID)
	if err != nil {
		return nil, err
	}

	if body == nil {
		body = &message.NetConnectionCreateStream{}
	}

	chunkStreamID := 3 // TODO: fix
	err = s.writeCommandMessage(
		chunkStreamID, 0, // TODO: fix, Timestamp is 0
		"createStream",
		transactionID,
		body,
	)
	if err != nil {
		return nil, err
	}

	// TODO: support timeout
	// TODO: check result
	timeoutCtx := context.TODO()
	select {
	case <-timeoutCtx.Done():
		return nil, timeoutCtx.Err()
	case <-t.doneCh:
		amfDec := message.NewAMFDecoder(t.body, t.encoding)

		var value message.AMFConvertible
		if err := message.DecodeBodyCreateStreamResult(t.body, amfDec, &value); err != nil {
			return nil, errors.Wrap(err, "Failed to decode result")
		}
		result := value.(*message.NetConnectionCreateStreamResult)

		if t.commandName == "_error" {
			return nil, &CreateStreamRejectedError{
				TransactionID: transactionID,
				Result:        result,
			}
		}

		return result, nil
	}

	//return nil, errors.New("Failed to get result")
}

func (s *Stream) DeleteStream(body *message.NetStreamDeleteStream) error {
	chunkStreamID := 3 // TODO: fix

	return s.writeCommandMessage(
		chunkStreamID,
		0,
		"deleteStream",
		0,
		body,
	)
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

func (s *Stream) Play(
	body *message.NetStreamPlay,
) error {
	if body == nil {
		body = &message.NetStreamPlay{}
	}

	chunkStreamID := 3 // TODO: fix
	return s.writeCommandMessage(
		chunkStreamID,
		0, // TODO: fix, Timestamp is 0
		"play",
		int64(0), // Always 0, 7.2.2.1
		body,
	)
}

func (s *Stream) Publish(
	body *message.NetStreamPublish,
) error {
	if body == nil {
		body = &message.NetStreamPublish{}
	}

	chunkStreamID := 3 // TODO: fix
	return s.writeCommandMessage(
		chunkStreamID, 0, // TODO: fix, Timestamp is 0
		"publish",
		int64(0), // Always 0, 7.2.2.6
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

func (s *Stream) Close() error {
	s.assumeClosed()
	return nil // TODO: implement
}

func (s *Stream) assumeClosed() {
	// TODO: implement
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

	return s.Write(chunkStreamID, timestamp, &message.CommandMessage{
		CommandName:   commandName,
		TransactionID: transactionID,
		Encoding:      s.encTy,
		Body:          buf,
	})
}

func (s *Stream) WriteDataMessage(
	chunkStreamID int,
	timestamp uint32,
	name string,
	body message.AMFConvertible,
) error {
	buf := new(bytes.Buffer)
	amfEnc := message.NewAMFEncoder(buf, message.EncodingTypeAMF0)
	if err := message.EncodeBodyAnyValues(amfEnc, body); err != nil {
		return err
	}

	return s.Write(chunkStreamID, timestamp, &message.DataMessage{
		Name:     name,
		Encoding: message.EncodingTypeAMF0,
		Body:     buf,
	})
}

func (s *Stream) WriteSetChunkSize(chunkSize uint32) error {
	if chunkSize < 1 {
		return errors.New("chunksize < 1")
	}
	if chunkSize > 0x7fffffff {
		return errors.New("chunksize > 0x7fffffff")
	}
	msg := &message.SetChunkSize{ChunkSize: chunkSize}
	chunkStreamID := 2       // Correct according to 6.2
	var timeStamp uint32 = 0 // TODO. Send updated time
	return s.Write(chunkStreamID, timeStamp, msg)
}

func (s *Stream) Write(chunkStreamID int, timestamp uint32, msg message.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // TODO: Fix 5s
	defer cancel()

	s.cmsg.Message = msg
	return s.streamer().Write(ctx, chunkStreamID, timestamp, &s.cmsg)
}

func (s *Stream) handle(chunkStreamID int, timestamp uint32, msg message.Message) error {
	return s.handler.Handle(chunkStreamID, timestamp, msg)
}

func (s *Stream) streams() *streams {
	return s.conn.streams
}

func (s *Stream) streamer() *ChunkStreamer {
	return s.conn.streamer
}

func (s *Stream) userHandler() Handler {
	return s.conn.handler
}

func (s *Stream) logger() logrus.FieldLogger {
	return s.conn.logger
}
