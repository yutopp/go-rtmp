//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/pkg/errors"

	"github.com/yutopp/go-rtmp/internal"
	"github.com/yutopp/go-rtmp/message"
)

var _ messageHandler = (*serverControlNotConnectedHandler)(nil)

// serverControlNotConnectedHandler Handle control messages from a client which has not send connect at server side.
//   transitions:
//     | "connect" -> controlStreamStateConnected
//     | _         -> self
//
type serverControlNotConnectedHandler struct {
	entry *entryHandler
}

func (h *serverControlNotConnectedHandler) Handle(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlNotConnectedHandler) HandleData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	body interface{},
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlNotConnectedHandler) HandleCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
	body interface{},
	stream *Stream,
) (err error) {
	l := h.entry.Logger()

	switch cmd := body.(type) {
	case *message.NetConnectionConnect:
		l.Info("Connect")
		defer func() {
			if err != nil {
				result := h.newConnectErrorResult()

				l.Infof("Connect(Error): ResponseBody = %#v", result)
				if err1 := stream.ReplyConnect(chunkStreamID, timestamp, result); err1 != nil {
					err = errors.Wrapf(err, "Failed to reply response: Err = %+v", err1)
				}
			}
		}()

		if err := h.entry.conn.handler.OnConnect(timestamp, cmd); err != nil {
			return err
		}

		l.Infof("Set win ack size: Size = %+v", h.entry.conn.streamer.SelfState().AckWindowSize())
		if err := stream.WriteWinAckSize(ctrlMsgChunkStreamID, timestamp, &message.WinAckSize{
			Size: h.entry.conn.streamer.SelfState().AckWindowSize(),
		}); err != nil {
			return err
		}

		l.Infof("Set peer bandwidth: Size = %+v, Limit = %+v",
			h.entry.conn.streamer.SelfState().BandwidthWindowSize(),
			h.entry.conn.streamer.SelfState().BandwidthLimitType(),
		)
		if err := stream.WriteSetPeerBandwidth(ctrlMsgChunkStreamID, timestamp, &message.SetPeerBandwidth{
			Size:  h.entry.conn.streamer.SelfState().BandwidthWindowSize(),
			Limit: h.entry.conn.streamer.SelfState().BandwidthLimitType(),
		}); err != nil {
			return err
		}

		l.Infof("Stream Begin: ID = %d", 0)
		if err := stream.WriteUserCtrl(ctrlMsgChunkStreamID, timestamp, &message.UserCtrl{
			Event: &message.UserCtrlEventStreamBegin{
				StreamID: 0,
			},
		}); err != nil {
			return err
		}

		result := h.newConnectSuccessResult()

		l.Infof("Connect: ResponseBody = %#v", result)
		if err := stream.ReplyConnect(chunkStreamID, timestamp, result); err != nil {
			return err
		}
		l.Info("Connected")

		h.entry.ChangeState(&serverControlConnectedHandler{entry: h.entry})

		return nil

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *serverControlNotConnectedHandler) newConnectSuccessResult() *message.NetConnectionConnectResult {
	return &message.NetConnectionConnectResult{
		Properties: message.NetConnectionConnectResultProperties{
			FMSVer:       "GO-RTMP/0,0,0,0", // TODO: fix
			Capabilities: 31,                // TODO: fix
			Mode:         1,                 // TODO: fix
		},
		Information: message.NetConnectionConnectResultInformation{
			Level:       "status",
			Code:        message.NetConnectionConnectCodeSuccess,
			Description: "Connection succeeded.",
			Data: map[string]interface{}{
				"type":    "go-rtmp",
				"version": "master", // TODO: fix
			},
		},
	}
}

func (h *serverControlNotConnectedHandler) newConnectErrorResult() *message.NetConnectionConnectResult {
	return &message.NetConnectionConnectResult{
		Properties: message.NetConnectionConnectResultProperties{
			FMSVer:       "GO-RTMP/0,0,0,0", // TODO: fix
			Capabilities: 31,                // TODO: fix
			Mode:         1,                 // TODO: fix
		},
		Information: message.NetConnectionConnectResultInformation{
			Level:       "error",
			Code:        message.NetConnectionConnectCodeFailed,
			Description: "Connection failed.",
			Data: map[string]interface{}{
				"type":    "go-rtmp",
				"version": "master", // TODO: fix
			},
		},
	}
}
