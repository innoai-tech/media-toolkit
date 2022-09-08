package format

import (
	"context"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/format/mp4"
	"github.com/innoai-tech/media-toolkit/pkg/storage/mime"
	"github.com/pion/mediadevices"
	"golang.org/x/sync/errgroup"
	"io"
	"time"
)

type Info struct {
	ID        string
	MediaType string
	At        time.Time
	StartedAt time.Time
}

type Recorder interface {
	Record() (*Info, error)
	Close() error
}

func NewRecorder(f io.WriteSeeker, videoTrack mediadevices.Track, codecName string) (Recorder, error) {
	encodedReader, err := videoTrack.NewEncodedReader(codecName)
	if err != nil {
		return nil, err
	}

	return &recorder{
		videoTrack:    videoTrack,
		encodedReader: encodedReader,
		muxer:         mp4.NewMuxer(f),
		release: func() {
		},
		p: Packetizer{},
	}, nil
}

type recorder struct {
	videoTrack    mediadevices.Track
	muxer         *mp4.Muxer
	encodedReader mediadevices.EncodedReadCloser
	startedAt     time.Time
	p             Packetizer
	release       func()
}

func (r *recorder) Close() error {
	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return r.encodedReader.Close()
	})

	g.Go(func() error {
		return r.muxer.WriteTrailer()
	})

	return g.Wait()
}

func (r *recorder) Record() (*Info, error) {
	buf, release, err := r.encodedReader.Read()
	if err != nil {
		return nil, err
	}
	r.release = release
	if len(buf.Data) == 0 {
		r.release()
	}

	first := r.p.Time == 0
	pkt := r.p.Packetize(buf.Data, buf.Samples)

	if first {
		r.startedAt = time.Now().Add(-r.p.Time)

		codecData, err := h264parser.NewCodecDataFromSPSAndPPS(pkt.SPS, pkt.PPS)
		if err != nil {
			return nil, err
		}

		if err := r.muxer.WriteHeader([]av.CodecData{codecData}); err != nil {
			return nil, err
		}
	}

	if err := r.muxer.WritePacket(av.Packet{
		Idx:        0,
		IsKeyFrame: pkt.IsKeyFrame,
		Time:       pkt.Time,
		Data:       pkt.Data,
	}); err != nil {
		return nil, err
	}

	return &Info{
		ID:        r.videoTrack.ID(),
		MediaType: mime.MediaTypeVideoMP4,
		At:        r.startedAt.Add(r.p.Time),
		StartedAt: r.startedAt,
	}, nil
}
