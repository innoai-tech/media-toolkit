package livestream

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"sync"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/rtspv2"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
	"github.com/pkg/errors"
)

var (
	StreamExitNoVideoOnStream = errors.New("stream exit no video on stream")
	StreamExitRtspDisconnect  = errors.New("stream exit rtsp disconnect")
	StreamExitIdleTimeout     = errors.New("stream exit idle timeout")
)

type StreamSubject interface {
	Info() Stream
	Status() Status
	Subscribe(ctx context.Context, o StreamObserver) (io.Closer, error)
}

type Status struct {
	Active    bool           `json:"active"`
	Observers map[string]int `json:"observers"`
}

type Packet struct {
	Stream
	At     time.Time
	Codecs []av.CodecData
	av.Packet
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

	observers sync.Map

	worker syncutil.ValueMutex[*worker]
	codecs syncutil.ValueMutex[[]av.CodecData]
}

func (s *streamSubject) Close() error {
	if w := s.worker.Get(); w != nil {
		defer func() {
			s.worker.Set(nil)
		}()
		return w.Close()
	}
	return nil
}

func (s *streamSubject) Name() string {
	return "StreamSubject"
}

func (s *streamSubject) Subscribe(ctx context.Context, o StreamObserver) (io.Closer, error) {
	if c, ok := o.(CanUniqueKey); ok {
		if ss, ok := s.observers.Load(c.UniqueKey()); ok {
			return ss.(StreamObserver), nil
		}
	}

	l := logr.FromContextOrDiscard(ctx).WithValues(
		"stream_id", s.stream.ID, "stream_name", s.stream.Name,
	)

	// try to serve worker when observer add
	if w := s.worker.Get(); w == nil {
		w := s.worker.Set(&worker{
			CloseNotifier: syncutil.NewCloseNotifier(),
		})

		go func() {
			for {
				l.Info("starting...")
				err := w.serve(s.stream, s)
				if err == nil {
					break
				}
				l.Error(err, "worker exit error")
				if err == StreamExitRtspDisconnect {
					time.Sleep(3 * time.Second)
					l.Info("will reconnect 3s later")
					continue
				}
				break
			}
			s.worker.Set(nil)
		}()
	}

	ss := &streamObserverWrapper{
		StreamObserver: o,
		streamSubject:  s,
	}

	go func() {
		for {
			select {
			case _ = <-ss.Done():
				l.V(1).Info(fmt.Sprintf("stream observer `%s` removed", o.Name()))
				_ = ss.Close()
				return
			}
		}
	}()

	l.V(1).Info(fmt.Sprintf("stream observer `%s` added.", o.Name()))
	s.observers.Store(ss.UniqueKey(), ss)

	return ss, nil
}

type streamObserverWrapper struct {
	StreamObserver
	streamSubject *streamSubject
}

func (s *streamObserverWrapper) UniqueKey() any {
	if can, ok := s.StreamObserver.(CanUniqueKey); ok {
		return can.UniqueKey()
	}
	return s
}

func (s *streamObserverWrapper) Close() error {
	defer func() {
		s.streamSubject.observers.Delete(s.UniqueKey())
	}()
	return s.StreamObserver.Close()
}

func (s *streamSubject) WritePacket(pkt *Packet) {
	s.observers.Range(func(_, value any) bool {
		go value.(StreamObserver).WritePacket(pkt)
		return true
	})
}

func (s *streamSubject) HasObserver() bool {
	observerCount := 0
	s.observers.Range(func(_, value any) bool {
		observerCount++
		return true
	})
	return observerCount > 0
}

func (s *streamSubject) Status() Status {
	status := Status{
		Observers: map[string]int{},
	}

	s.observers.Range(func(_, value any) bool {
		so := value.(StreamObserver)
		status.Observers[so.Name()] += 1
		return true
	})

	status.Active = s.worker.Get() != nil

	return status
}

func (s *streamSubject) Info() Stream {
	return s.stream
}

type worker struct {
	syncutil.CloseNotifier
}

func (s *worker) dial(rtspURL string) (*rtspv2.RTSPClient, error) {
	return rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:              rtspURL,
		DialTimeout:      5 * time.Second,
		ReadWriteTimeout: 5 * time.Second,
		DisableAudio:     true,
		Debug:            false,
	})
}

func (s *worker) serve(stream Stream, observer StreamNotifier) error {
	s.Reset()

	c, err := s.dial(stream.Rtsp)
	if err != nil {
		return err
	}

	defer func() {
		s.SendDone(err)
		c.Close()
	}()

	idleTimeout := 20 * time.Second
	keyTimeout := 20 * time.Second

	keyTimer := time.NewTimer(keyTimeout)
	idleTimer := time.NewTimer(idleTimeout)

	var audioOnly bool

	codecData := c.CodecData
	if codecData != nil {
		if len(codecData) == 1 && codecData[0].Type().IsAudio() {
			audioOnly = true
		}
	}

	var packetStartedAt *time.Time
	var packetStartedDur *time.Duration

	timelines := make(map[int8]time.Duration)

	var lastAvp *av.Packet

	for {
		select {
		case <-s.Done():
			return nil
		case <-idleTimer.C:
			return StreamExitIdleTimeout
		case <-keyTimer.C:
			return StreamExitNoVideoOnStream
		case signals := <-c.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				codecData = c.CodecData
			case rtspv2.SignalStreamRTPStop:
				return StreamExitRtspDisconnect
			}
		case avp := <-c.OutgoingPacketQueue:
			if audioOnly || avp.IsKeyFrame {
				keyTimer.Reset(keyTimeout)
			}

			if can, ok := observer.(interface{ HasObserver() bool }); ok {
				if can.HasObserver() {
					idleTimer.Reset(idleTimeout)
				}
			}

			// time force fix
			timelines[avp.Idx] += avp.Duration
			avp.Time = timelines[avp.Idx]

			if packetStartedAt == nil {
				t := time.Now()
				packetStartedAt = &t
				packetStartedDur = &avp.Time
			}

			if lastAvp != nil && avp.Time < lastAvp.Time {
				// ignore invalid frame
				continue
			}

			lastAvp = avp

			observer.WritePacket(&Packet{
				Stream: stream,
				At:     packetStartedAt.Add(avp.Time - *packetStartedDur),
				Codecs: codecData,
				Packet: *avp,
			})
		}
	}
}
