package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/yutopp/go-rtmp"
)

func main() {
	client, err := rtmp.Dial("rtmp://localhost:1935", &rtmp.ConnConfig{
		Logger: log.StandardLogger(),
	})
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}
	defer client.Close()
	log.Infof("Client created")

	if err := client.Connect(); err != nil {
		log.Infof("Failed to connect: Err=%+v", err)
	}
	log.Infof("connected")

	stream, err := client.CreateStream()
	if err != nil {
		log.Infof("Failed to create stream: Err=%+v", err)
	}
	//defer stream.Close()
	_ = stream

	log.Infof("stream created")
}
