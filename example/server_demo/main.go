package main

import (
	"github.com/sirupsen/logrus"
	"log"
	"net"

	"github.com/yutopp/go-rtmp"
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
		HandlerFactory: func(conn *rtmp.Conn) rtmp.Handler {
			l := logrus.StandardLogger()
			//l.SetLevel(logrus.DebugLevel)
			conn.SetLogger(l)

			return &Handler{}
		},
		Conn: &rtmp.ConnConfig{
			MaxBitrateKbps: 6 * 1024,
			ControlState: rtmp.StreamControlStateConfig{
				DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
			},
		},
	})
	if err := srv.Serve(listner); err != nil {
		log.Panicf("Failed: %+v", err)
	}
}
