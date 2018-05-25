package main

import (
	"log"
	"net"

	"github.com/yutopp/rtmp-go"
	rtmpMsg "github.com/yutopp/rtmp-go/message"

	"github.com/yutopp/flv-go"
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

func handler(m rtmpMsg.Message, timestamp uint64, s rtmp.Stream) error {
	log.Printf("MESSAGE: %+v", m)

	switch msg := m.(type) {
	case *rtmpMsg.AudioMessage:
		audio, err := flv.ParseAudioData(msg.Payload)
		if err != nil {
			return err
		}

		log.Printf("FLV Audio Data: %+v", audio)

	case *rtmpMsg.VideoMessage:
		video, err := flv.ParseVideoData(msg.Payload)
		if err != nil {
			return err
		}

		log.Printf("FLV Video Data: %+v", video)
	}

	return nil
}
