package main

import (
	"bytes"
	"sync"

	flvtag "github.com/yutopp/go-flv/tag"
)

type Pubsub struct {
	srv  *RelayService
	name string

	pub  *Pub
	subs []*Sub

	m sync.Mutex
}

func NewPubsub(srv *RelayService, name string) *Pubsub {
	return &Pubsub{
		srv:  srv,
		name: name,

		subs: make([]*Sub, 0),
	}
}

func (pb *Pubsub) Deregister() error {
	pb.m.Lock()
	defer pb.m.Unlock()

	for _, sub := range pb.subs {
		_ = sub.Close()
	}

	return pb.srv.RemovePubsub(pb.name)
}

func (pb *Pubsub) Pub() *Pub {
	pub := &Pub{
		pb: pb,
	}

	pb.pub = pub

	return pub
}

func (pb *Pubsub) Sub() *Sub {
	pb.m.Lock()
	defer pb.m.Unlock()

	sub := &Sub{}

	// TODO: Implement more efficient resource management
	pb.subs = append(pb.subs, sub)

	return sub
}

type Pub struct {
	pb *Pubsub

	avcSeqHeader *flvtag.FlvTag
	lastKeyFrame *flvtag.FlvTag
}

// TODO: Should check codec types and so on.
// In this example, checks only sequence headers and assume that AAC and AVC.
func (p *Pub) Publish(flv *flvtag.FlvTag) error {
	switch flv.Data.(type) {
	case *flvtag.AudioData, *flvtag.ScriptData:
		for _, sub := range p.pb.subs {
			_ = sub.onEvent(cloneView(flv))
		}

	case *flvtag.VideoData:
		d := flv.Data.(*flvtag.VideoData)
		if d.AVCPacketType == flvtag.AVCPacketTypeSequenceHeader {
			p.avcSeqHeader = flv
		}

		if d.FrameType == flvtag.FrameTypeKeyFrame {
			p.lastKeyFrame = flv
		}

		for _, sub := range p.pb.subs {
			if !sub.initialized {
				if p.avcSeqHeader != nil {
					_ = sub.onEvent(cloneView(p.avcSeqHeader))
				}
				if p.lastKeyFrame != nil {
					_ = sub.onEvent(cloneView(p.lastKeyFrame))
				}
				sub.initialized = true
				continue
			}

			_ = sub.onEvent(cloneView(flv))
		}

	default:
		panic("unexpected")
	}

	return nil
}

func (p *Pub) Close() error {
	return p.pb.Deregister()
}

type Sub struct {
	initialized bool
	closed      bool

	lastTimestamp uint32
	eventCallback func(*flvtag.FlvTag) error
}

func (s *Sub) onEvent(flv *flvtag.FlvTag) error {
	if s.closed {
		return nil
	}

	if flv.Timestamp != 0 && s.lastTimestamp == 0 {
		s.lastTimestamp = flv.Timestamp
	}
	flv.Timestamp -= s.lastTimestamp

	return s.eventCallback(flv)
}

func (s *Sub) Close() error {
	if s.closed {
		return nil
	}

	s.closed = true

	return nil
}

func cloneView(flv *flvtag.FlvTag) *flvtag.FlvTag {
	// Need to clone the view because Binary data will be consumed
	v := *flv

	switch flv.Data.(type) {
	case *flvtag.AudioData:
		dCloned := *v.Data.(*flvtag.AudioData)
		v.Data = &dCloned

		dCloned.Data = bytes.NewBuffer(dCloned.Data.(*bytes.Buffer).Bytes())

	case *flvtag.VideoData:
		dCloned := *v.Data.(*flvtag.VideoData)
		v.Data = &dCloned

		dCloned.Data = bytes.NewBuffer(dCloned.Data.(*bytes.Buffer).Bytes())

	case *flvtag.ScriptData:
		dCloned := *v.Data.(*flvtag.ScriptData)
		v.Data = &dCloned

	default:
		panic("unreachable")
	}

	return &v
}
