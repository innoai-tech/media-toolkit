package livestream

import (
	_ "embed"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/pkg/errors"

	"github.com/deepch/vdk/format/rtspv2"
)

var (
	StreamExitNoVideoOnStream = errors.New("stream exit no video on stream")
	StreamExitRtspDisconnect  = errors.New("stream exit rtsp disconnect")
	StreamExitIdleTimeout     = errors.New("stream exit rtsp disconnect")
)

type StreamSubject interface {
	Info() Stream
	Subscribe(o StreamObserver) (io.Closer, error)
}

type Packet struct {
	Stream
	av.Packet
	At time.Time
}

func (p Packet) Clone() Packet {
	data := make([]byte, len(p.Data))
	copy(data, p.Data)
	p.Data = data
	return p
}

func NewStreamSubject(stream Stream) StreamSubject {
	return &streamSubject{
		stream: stream,
	}
}

type streamSubject struct {
	stream Stream

	keyTimer  *time.Timer
	idleTimer *time.Timer
	c         *rtspv2.RTSPClient

	codecData []av.CodecData

	observers     sync.Map
	observerCount int32

	done chan struct {
	}

	rw sync.RWMutex
}

func (s *streamSubject) Info() Stream {
	return s.stream
}

func (s *streamSubject) Subscribe(o StreamObserver) (io.Closer, error) {
	if !s.IsActive() {
		go func() {
			_ = s.Start()
		}()
	}

	ss := &streamSubjection{
		StreamObserver: o,
		streamSubject:  s,
	}

	atomic.AddInt32(&s.observerCount, 1)
	s.observers.Store(ss, true)

	if codecData := s.codecData; codecData != nil {
		ss.SetCodecData(codecData)
	}

	return ss, nil
}

type streamSubjection struct {
	StreamObserver
	streamSubject *streamSubject
}

func (s *streamSubjection) Close() error {
	atomic.AddInt32(&s.streamSubject.observerCount, -1)
	s.streamSubject.observers.Delete(s)
	return s.StreamObserver.Close()
}

func (s *streamSubject) SetCodecData(data []av.CodecData) {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.codecData = data

	s.observers.Range(func(key, value any) bool {
		go key.(StreamObserver).SetCodecData(data)
		return true
	})
}

func (s *streamSubject) WritePacket(pkt Packet) {
	s.observers.Range(func(key, value any) bool {
		go key.(StreamObserver).WritePacket(pkt)
		return true
	})
}

func (s *streamSubject) HasObserver() bool {
	return true
}

func (s *streamSubject) Start() error {
	if err := s.Dial(s.stream.Rtsp); err != nil {
		return err
	}

	idleTimeout := 20 * time.Second
	keyTimeout := 20 * time.Second

	s.keyTimer = time.NewTimer(keyTimeout)
	s.idleTimer = time.NewTimer(idleTimeout)

	var audioOnly bool

	if codecData := s.c.CodecData; codecData != nil {
		if len(codecData) == 1 && codecData[0].Type().IsAudio() {
			audioOnly = true
		}
		s.SetCodecData(codecData)
	}

	var packetStartedAt *time.Time
	var packetStartedDur *time.Duration

	defer func() {
		s.observers.Range(func(key, value any) bool {
			_ = key.(StreamObserver).Close()
			return true
		})
	}()

	for {
		select {
		case <-s.done:
			return nil
		case <-s.idleTimer.C:
			return StreamExitIdleTimeout
		case <-s.keyTimer.C:
			return StreamExitNoVideoOnStream
		case signals := <-s.c.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				s.SetCodecData(s.c.CodecData)
			case rtspv2.SignalStreamRTPStop:
				return StreamExitRtspDisconnect
			}
		case packetAV := <-s.c.OutgoingPacketQueue:
			if audioOnly || packetAV.IsKeyFrame {
				s.keyTimer.Reset(keyTimeout)
			}

			t := time.Now()

			if packetStartedAt == nil {
				packetStartedAt = &t
				packetStartedDur = &packetAV.Time
			}

			if s.HasObserver() {
				s.idleTimer.Reset(idleTimeout)
			}

			s.WritePacket(Packet{
				Stream: s.stream,
				At:     packetStartedAt.Add(packetAV.Time - *packetStartedDur),
				Packet: *packetAV,
			})
		}
	}
}

func (s *streamSubject) Dial(rtspURL string) error {
	c, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:              rtspURL,
		DialTimeout:      3 * time.Second,
		ReadWriteTimeout: 3 * time.Second,
		DisableAudio:     true,
		Debug:            false,
	})
	if err != nil {
		return err
	}
	s.c = c
	return nil
}

func (s *streamSubject) IsActive() bool {
	s.rw.Lock()
	defer s.rw.Unlock()

	return s.c != nil
}

func (s *streamSubject) Close() error {
	s.rw.Lock()
	defer s.rw.Unlock()

	if s.c != nil {
		close(s.done)
		s.c.Close()
		s.c = nil
	}

	return nil
}
