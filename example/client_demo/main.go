package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/yutopp/go-rtmp"
	rtmpmsg "github.com/yutopp/go-rtmp/message"
)

func main() {
	client, err := rtmp.Dial("rtmp", "localhost:1935", &rtmp.ConnConfig{
		Logger: log.StandardLogger(),
	})
	if err != nil {
		log.Fatalf("Failed to dial: %+v", err)
	}
	defer client.Close()
	log.Infof("Client created")

	if err := client.Connect(nil); err != nil {
		log.Fatalf("Failed to connect: Err=%+v", err)
	}
	log.Infof("connected")

	stream, err := client.CreateStream(nil)
	if err != nil {
		log.Fatalf("Failed to create stream: Err=%+v", err)
	}
	defer stream.Close()

	if err := stream.Publish(&rtmpmsg.NetStreamPublish{
		PublishingName: "testtesttesttest",
		PublishingType: "live",
	}); err != nil {
		log.Fatalf("Failed to send publish message: Err=%+v", err)
	}

	log.Infof("stream created")
}
