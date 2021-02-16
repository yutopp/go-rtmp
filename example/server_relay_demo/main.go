package main

import (
	"fmt"
	"io"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/yutopp/go-rtmp"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1935")
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	relayService := NewRelayService()

	srv := rtmp.NewServer(&rtmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			l := log.StandardLogger()
			//l.SetLevel(logrus.DebugLevel)

			h := &Handler{
				relayService: relayService,
			}

			return conn, &rtmp.ConnConfig{
				Handler: h,

				ControlState: rtmp.StreamControlStateConfig{
					DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
				},

				Logger: l,
			}
		},
	})
	if err := srv.Serve(listener); err != nil {
		log.Panicf("Failed: %+v", err)
	}
}

type RelayService struct {
	streams map[string]*Pubsub
	m       sync.Mutex
}

func NewRelayService() *RelayService {
	return &RelayService{
		streams: make(map[string]*Pubsub),
	}
}

func (s *RelayService) GetPubsub(key string) (*Pubsub, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.streams[key]; !ok {
		return nil, fmt.Errorf("Already published: %s", key)
	}

	pubsub := &Pubsub{}

	s.streams[key] = pubsub

	return pubsub, nil
}

type Pubsub struct {
}

func (pb *Pubsub) Pub() *Pub {
	return nil
}

type Pub struct {
}

type Sub struct {
}
