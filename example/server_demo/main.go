package main

import (
	"github.com/yutopp/go-rtmp"
	"log"
	"net"
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

	srv := rtmp.NewServer(&rtmp.ServerConfig{
		HandlerFactory: func(_ *rtmp.Conn) rtmp.Handler {
			return &Handler{}
		},
		Conn: nil,
	})
	if err := srv.Serve(listner); err != nil {
		log.Panicf("Failed: %+v", err)
	}
}
