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

func (h *serverControlNotConnectedHandler) HandleCommand(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	cmdMsg *message.CommandMessage,
	stream *Stream,
) error {
	l := h.entry.Logger()

	switch cmd := cmdMsg.Command.(type) {
	case *message.NetConnectionConnect:
		l.Info("Connect")

		if err := h.entry.handler.OnConnect(timestamp, cmd); err != nil {
			cmdRespMsg := h.newConnectErrorMessage()

			l.Infof("Reject a connect request: Response = %#v", cmdRespMsg.Command)
			if writeErr := stream.WriteCommandMessage(chunkStreamID, timestamp, encTy, cmdRespMsg); writeErr != nil {
				return errors.Wrapf(err, "Write failed: Err = %+v", writeErr)
			}

			return err
		}

		l.Infof("Set win ack size: Size = %+v", h.entry.streamer.SelfState().AckWindowSize())
		if err := stream.WriteWinAckSize(ctrlMsgChunkStreamID, timestamp, &message.WinAckSize{
			Size: h.entry.streamer.SelfState().AckWindowSize(),
		}); err != nil {
			return err
		}

		l.Infof("Set peer bandwidth: Size = %+v, Limit = %+v",
			h.entry.streamer.SelfState().BandwidthWindowSize(),
			h.entry.streamer.SelfState().BandwidthLimitType(),
		)
		if err := stream.WriteSetPeerBandwidth(ctrlMsgChunkStreamID, timestamp, &message.SetPeerBandwidth{
			Size:  h.entry.streamer.SelfState().BandwidthWindowSize(),
			Limit: h.entry.streamer.SelfState().BandwidthLimitType(),
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

		cmdRespMsg := h.newConnectSuccessMessage()
		l.Infof("Connect: Response = %#v", cmdRespMsg.Command)
		if err := stream.WriteCommandMessage(chunkStreamID, timestamp, encTy, cmdRespMsg); err != nil {
			return err
		}
		l.Info("Connected")

		h.entry.ChangeState(&serverControlConnectedHandler{entry: h.entry})

		return nil

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *serverControlNotConnectedHandler) HandleData(
	chunkStreamID int,
	timestamp uint32,
	encTy message.EncodingType,
	dataMsg *message.DataMessage,
	stream *Stream,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlNotConnectedHandler) newConnectSuccessMessage() *message.CommandMessage {
	return &message.CommandMessage{
		CommandName:   "_result",
		TransactionID: 1, // 7.2.1.2, flow.6
		Command: &message.NetConnectionConnectResult{
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
		},
	}
}

func (h *serverControlNotConnectedHandler) newConnectErrorMessage() *message.CommandMessage {
	return &message.CommandMessage{
		CommandName:   "_error",
		TransactionID: 1, // 7.2.1.2, flow.6
		Command: &message.NetConnectionConnectResult{
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
		},
	}
}
