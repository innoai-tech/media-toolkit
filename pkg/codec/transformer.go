package codec

import (
	"context"
	"encoding/binary"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pkg/errors"
	"image"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/x264"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
	"golang.org/x/sync/errgroup"
)

func AsVideoTransformer(dec Decoder) (*VideoTransformer, error) {
	x264Params, err := x264.NewParams()
	if err != nil {
		return nil, err
	}

	x264Params.Preset = x264.PresetMedium
	x264Params.BitRate = 4_000_000 // 4mbps

	info := dec.Info()

	t := &VideoTransformer{
		dec:    dec,
		params: x264Params,
	}

	enc, err := x264Params.BuildVideoEncoder(video.ReaderFunc(func() (img image.Image, release func(), err error) {
		if t.frame == nil {
			return nil, nil, errors.New("no frame!")
		}
		return t.frame, func() {}, nil
	}), prop.Media{
		Video: prop.Video{
			Width:       info.Width,
			Height:      info.Height,
			FrameFormat: frame.FormatI420,
		},
	})

	if err != nil {
		return nil, err
	}

	t.enc = enc

	return t, nil
}

type VideoTransformer struct {
	params    x264.Params
	dec       Decoder
	enc       codec.ReadCloser
	codecData av.CodecData
	frame     *image.YCbCr
}

func (t *VideoTransformer) Info() Info {
	return t.dec.Info()
}

func (t *VideoTransformer) Close() error {
	t.release()
	eg, _ := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		return t.dec.Close()
	})
	eg.Go(func() error {
		return t.enc.Close()
	})
	return eg.Wait()
}

func (t *VideoTransformer) release() {
	t.frame = nil
}

func (t *VideoTransformer) CodecData() av.CodecData {
	return t.codecData
}

type Result struct {
	IsKeyFrame bool
	Data       []byte
	Frame      *image.YCbCr
}

func (t *VideoTransformer) Transform(data []byte) (bool, *Result, error) {
	imgFrame, err := t.dec.DecodeToImage(data)
	if err != nil {
		if err == EAGAIN {
			return false, nil, nil
		}
		return false, nil, err
	}
	t.frame = imgFrame

	f, release, err := t.enc.Read()
	if err != nil {
		return false, nil, err
	}
	defer release()

	var sps, pps []byte
	isKeyFrame := false
	nalRaw := make([]byte, 0, len(f))

	emitNalus(f, func(nalu []byte) {
		naluType := nalu[0] & 0x1f

		switch {
		case naluType >= 1 && naluType <= 5:
			isKeyFrame = naluType == 5
			nalRaw = append(nalRaw, append(binSize(len(nalu)), nalu...)...)
		case naluType == 7: // sps
			sps = nalu[0:]
		case naluType == 8: // pps
			pps = nalu[0:]
		case naluType == 6: // sei
			// skip
		}
	})

	if t.codecData == nil && len(sps) > 0 {
		codecData, err := h264parser.NewCodecDataFromSPSAndPPS(sps, pps)
		if err != nil {
			return false, nil, err
		}
		t.codecData = codecData
	}

	return true, &Result{
		Data:       nalRaw,
		IsKeyFrame: isKeyFrame,
		Frame:      imgFrame,
	}, nil
}

func emitNalus(nals []byte, emit func([]byte)) {
	nextInd := func(nalu []byte, start int) (indStart int, indLen int) {
		zeroCount := 0

		for i, b := range nalu[start:] {
			if b == 0 {
				zeroCount++
				continue
			} else if b == 1 {
				if zeroCount >= 2 {
					return start + i - zeroCount, zeroCount + 1
				}
			}
			zeroCount = 0
		}
		return -1, -1
	}

	nextIndStart, nextIndLen := nextInd(nals, 0)
	if nextIndStart == -1 {
		emit(nals)
	} else {
		for nextIndStart != -1 {
			prevStart := nextIndStart + nextIndLen
			nextIndStart, nextIndLen = nextInd(nals, prevStart)
			if nextIndStart != -1 {
				emit(nals[prevStart:nextIndStart])
			} else {
				// Emit until end of stream, no end indicator found
				emit(nals[prevStart:])
			}
		}
	}
}

func binSize(val int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(val))
	return buf
}
