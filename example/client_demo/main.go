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

	client.Connect()
}
