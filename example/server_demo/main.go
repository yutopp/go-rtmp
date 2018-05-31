package main

import (
	"bytes"
	"errors"
	flvtag "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"
	"log"
	"net"
)

type Handler struct{}

func (h *Handler) OnConnect(timestamp uint32, args []interface{}) error {
	log.Printf("OnConnect: %+v", args)
	return nil
}

func (h *Handler) OnPublish(timestamp uint32, args []interface{}) error {
	log.Printf("OnPublish: %+v", args)
	return nil
}

func (h *Handler) OnPlay(timestamp uint32, args []interface{}) error {
	return errors.New("Not supported")
}

func (h *Handler) OnAudio(timestamp uint32, payload []byte) error {
	buf := bytes.NewBuffer(payload)
	audio, err := flvtag.DecodeAudioData(buf)
	if err != nil {
		return err
	}

	log.Printf("FLV Audio Data: %+v", audio)
	return nil
}

func (h *Handler) OnVideo(timestamp uint32, payload []byte) error {
	buf := bytes.NewBuffer(payload)
	video, err := flvtag.DecodeVideoData(buf)
	if err != nil {
		return err
	}

	log.Printf("FLV Video Data: %+v", video)
	return nil
}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1935")
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	listner, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	srv := &rtmp.Server{}
	if err := srv.Serve(listner, func() rtmp.Handler {
		return &Handler{}
	}); err != nil {
		log.Panicf("Failed: %+v", err)
	}
}

/*
func handler(m rtmpmsg.Message, timestamp uint64, s rtmp.Stream) error {
	log.Printf("MESSAGE: %+v", m)

	switch msg := m.(type) {
	case *rtmpmsg.AudioMessage:
		buf := bytes.NewBuffer(msg.Payload)
		audio, err := flvtag.DecodeAudioData(buf)
		if err != nil {
			return err
		}

		log.Printf("FLV Audio Data: %+v", audio)

	case *rtmpmsg.VideoMessage:
		buf := bytes.NewBuffer(msg.Payload)
		video, err := flvtag.DecodeVideoData(buf)
		if err != nil {
			return err
		}

		log.Printf("FLV Video Data: %+v", video)
	}

	return nil
}
*/
