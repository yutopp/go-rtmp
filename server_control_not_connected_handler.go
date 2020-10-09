//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"github.com/pkg/errors"

	"github.com/edgeware/go-rtmp/internal"
	"github.com/edgeware/go-rtmp/message"
)

var _ stateHandler = (*serverControlNotConnectedHandler)(nil)

// serverControlNotConnectedHandler Handle control messages from a client which has not send connect at server side.
//   transitions:
//     | "connect" -> controlStreamStateConnected
//     | _         -> self
//
type serverControlNotConnectedHandler struct {
	sh *streamHandler
}

func (h *serverControlNotConnectedHandler) onMessage(
	chunkStreamID int,
	timestamp uint32,
	msg message.Message,
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlNotConnectedHandler) onData(
	chunkStreamID int,
	timestamp uint32,
	dataMsg *message.DataMessage,
	body interface{},
) error {
	return internal.ErrPassThroughMsg
}

func (h *serverControlNotConnectedHandler) onCommand(
	chunkStreamID int,
	timestamp uint32,
	cmdMsg *message.CommandMessage,
	body interface{},
) (err error) {
	l := h.sh.Logger()

	switch cmd := body.(type) {
	case *message.NetConnectionConnect:
		l.Info("Connect")
		defer func() {
			if err != nil {
				result := h.newConnectErrorResult()

				l.Infof("Connect(Error): ResponseBody = %#v, Err = %+v", result, err)
				if err1 := h.sh.stream.ReplyConnect(chunkStreamID, timestamp, result); err1 != nil {
					err = errors.Wrapf(err, "Failed to reply response: Err = %+v", err1)
				}
			}
		}()

		if err := h.sh.stream.userHandler().OnConnect(timestamp, cmd); err != nil {
			return err
		}

		l.Infof("Set win ack size: Size = %+v", h.sh.stream.streamer().SelfState().AckWindowSize())
		if err := h.sh.stream.WriteWinAckSize(ctrlMsgChunkStreamID, timestamp, &message.WinAckSize{
			Size: h.sh.stream.streamer().SelfState().AckWindowSize(),
		}); err != nil {
			return err
		}

		l.Infof("Set peer bandwidth: Size = %+v, Limit = %+v",
			h.sh.stream.streamer().SelfState().BandwidthWindowSize(),
			h.sh.stream.streamer().SelfState().BandwidthLimitType(),
		)
		if err := h.sh.stream.WriteSetPeerBandwidth(ctrlMsgChunkStreamID, timestamp, &message.SetPeerBandwidth{
			Size:  h.sh.stream.streamer().SelfState().BandwidthWindowSize(),
			Limit: h.sh.stream.streamer().SelfState().BandwidthLimitType(),
		}); err != nil {
			return err
		}

		l.Infof("Stream Begin: ID = %d", h.sh.stream.streamID)
		if err := h.sh.stream.WriteUserCtrl(ctrlMsgChunkStreamID, timestamp, &message.UserCtrl{
			Event: &message.UserCtrlEventStreamBegin{
				StreamID: h.sh.stream.streamID,
			},
		}); err != nil {
			return err
		}

		result := h.newConnectSuccessResult()

		l.Infof("Connect: ResponseBody = %#v", result)
		if err := h.sh.stream.ReplyConnect(chunkStreamID, timestamp, result); err != nil {
			return err
		}
		l.Info("Connected")

		h.sh.ChangeState(streamStateServerConnected)

		return nil

	default:
		return internal.ErrPassThroughMsg
	}
}

func (h *serverControlNotConnectedHandler) newConnectSuccessResult() *message.NetConnectionConnectResult {
	rPreset := h.sh.stream.conn.config.RPreset
	if rPreset == nil {
		rPreset = defaultResponsePreset
	}
	return &message.NetConnectionConnectResult{
		Properties: rPreset.GetServerConnectResultProperties(),
		Information: message.NetConnectionConnectResultInformation{
			Level:       "status",
			Code:        message.NetConnectionConnectCodeSuccess,
			Description: "Connection succeeded.",
			Data:        rPreset.GetServerConnectResultData(),
		},
	}
}

func (h *serverControlNotConnectedHandler) newConnectErrorResult() *message.NetConnectionConnectResult {
	rPreset := h.sh.stream.conn.config.RPreset
	if rPreset == nil {
		rPreset = defaultResponsePreset
	}
	return &message.NetConnectionConnectResult{
		Properties: rPreset.GetServerConnectResultProperties(),
		Information: message.NetConnectionConnectResultInformation{
			Level:       "error",
			Code:        message.NetConnectionConnectCodeFailed,
			Description: "Connection failed.",
			Data:        rPreset.GetServerConnectResultData(),
		},
	}
}
