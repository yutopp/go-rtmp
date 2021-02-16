package main

import (
	"bytes"
	"context"

	flvtag "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"
	rtmpmsg "github.com/yutopp/go-rtmp/message"
)

func onEventCallback(conn *rtmp.Conn, streamID uint32) func(flv *flvtag.FlvTag) error {
	return func(flv *flvtag.FlvTag) error {
		buf := new(bytes.Buffer)

		switch flv.Data.(type) {
		case *flvtag.VideoData:
			d := flv.Data.(*flvtag.VideoData)

			// Consume flv payloads (d)
			if err := flvtag.EncodeVideoData(buf, d); err != nil {
				return err
			}

			// TODO: Fix these values
			ctx := context.Background()
			chunkStreamID := 6
			return conn.Write(ctx, chunkStreamID, flv.Timestamp, &rtmp.ChunkMessage{
				StreamID: streamID,
				Message: &rtmpmsg.VideoMessage{
					Payload: buf,
				},
			})

		default:

		}

		return nil
	}
}
