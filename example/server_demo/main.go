package main

import (
	"bytes"
	"log"
	"net"

	"github.com/yutopp/go-rtmp"
	rtmpmsg "github.com/yutopp/go-rtmp/message"

	flvtag "github.com/yutopp/go-flv/tag"
)

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
	if err := srv.Serve(listner, handler); err != nil {
		log.Panicf("Failed: %+v", err)
	}
}

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
