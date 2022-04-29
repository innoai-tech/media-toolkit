package livestream

import (
	"context"
	_ "embed"
	"github.com/deepch/vdk/av"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sync"
	"time"

	"github.com/deepch/vdk/format/rtspv2"
)

var (
	StreamExitNoVideoOnStream = errors.New("stream exit no video on stream")
	StreamExitRtspDisconnect  = errors.New("stream exit rtsp disconnect")
	StreamExitIdleTimeout     = errors.New("stream exit rtsp disconnect")
)

type StreamConn interface {
	Dial(rtspURL string) error
	Serve(sw StreamWriter) error
	IsClosed() bool
	Close() error
}

func NewStreamConn(ctx context.Context, id string) StreamConn {
	return &streamConn{
		id:   id,
		done: make(chan struct{}),
		l:    logr.FromContextOrDiscard(ctx),
	}
}

type streamConn struct {
	id        string
	l         logr.Logger
	keyTimer  *time.Timer
	idleTimer *time.Timer
	c         *rtspv2.RTSPClient
	done      chan struct{}
	rw        sync.RWMutex
}

func (w *streamConn) Dial(rtspURL string) error {
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
	w.c = c
	return nil
}

func (w *streamConn) Serve(sw StreamWriter) error {
	idleTimeout := 20 * time.Second
	keyTimeout := 20 * time.Second

	w.keyTimer = time.NewTimer(keyTimeout)
	w.idleTimer = time.NewTimer(idleTimeout)

	if codecData := w.c.CodecData; codecData != nil {
		sw.SetCodecData(codecData)
	}

	audioOnly := false

	if len(w.c.CodecData) == 1 && w.c.CodecData[0].Type().IsAudio() {
		audioOnly = true
	}

	var packetStartedAt *time.Time
	var packetStartedDur *time.Duration

	for {
		select {
		case <-w.done:
			return nil
		case <-w.idleTimer.C:
			return StreamExitIdleTimeout
		case <-w.keyTimer.C:
			return StreamExitNoVideoOnStream
		case signals := <-w.c.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				sw.SetCodecData(w.c.CodecData)
			case rtspv2.SignalStreamRTPStop:
				return StreamExitRtspDisconnect
			}
		case packetAV := <-w.c.OutgoingPacketQueue:
			if audioOnly || packetAV.IsKeyFrame {
				w.keyTimer.Reset(keyTimeout)
			}

			t := time.Now()

			if packetStartedAt == nil {
				packetStartedAt = &t
				packetStartedDur = &packetAV.Time
			}

			if sw.Enabled() {
				w.idleTimer.Reset(idleTimeout)
			}

			sw.WritePacket(&Packet{
				ctx:    logr.NewContext(context.Background(), w.l),
				ID:     w.id,
				At:     packetStartedAt.Add(packetAV.Time - *packetStartedDur),
				Packet: *packetAV,
			})
		}
	}
}

func (w *streamConn) IsClosed() bool {
	w.rw.Lock()
	defer w.rw.Unlock()

	return w.c == nil
}

func (w *streamConn) Close() error {
	w.rw.Lock()
	defer w.rw.Unlock()

	if w.c != nil {
		close(w.done)
		w.c.Close()
		w.c = nil
	}

	return nil
}

type Packet struct {
	ctx context.Context
	ID  string
	At  time.Time
	av.Packet
}

func (c Packet) WithContext(ctx context.Context) *Packet {
	c.ctx = ctx
	return &c
}

func (c *Packet) Context() context.Context {
	return c.ctx
}
