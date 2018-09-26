package main

import (
	log "github.com/sirupsen/logrus"
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
		OnConnect: func() *rtmp.ConnConfig {
			l := log.StandardLogger()
			//l.SetLevel(logrus.DebugLevel)

			h := &Handler{}

			return &rtmp.ConnConfig{
				Handler: h,

				MaxBitrateKbps: 6 * 1024,
				ControlState: rtmp.StreamControlStateConfig{
					DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
				},

				Logger: l,
			}
		},
	})
	if err := srv.Serve(listner); err != nil {
		log.Panicf("Failed: %+v", err)
	}
}
