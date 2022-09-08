package codec

import (
	"fmt"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/codec/h265parser"
	"github.com/deepch/vdk/format/mp4"
	"github.com/innoai-tech/media-toolkit/pkg/util/fsutil"
	"image/jpeg"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"
)

func Test(t *testing.T) {
	t.Run("h265 to h264", func(t *testing.T) {
		ConvertFramesToMP4(t, "h265", 0, 100)
	})

	t.Run("h264 to h264", func(t *testing.T) {
		ConvertFramesToMP4(t, "h264", 0, 100)
	})
}

func ConvertFramesToMP4(t testing.TB, from string, start, end int) {
	frameRoot := fmt.Sprintf(".tmp/%s", from)

	confRecord, err := os.ReadFile(path.Join(frameRoot, "record.frame"))
	if err != nil {
		t.Fatal(err)
	}

	var d Decoder

	switch from {
	case "h264":
		c, err := h264parser.NewCodecDataFromAVCDecoderConfRecord(confRecord)
		if err != nil {
			t.Fatal(err)
		}
		d, err = NewH264Decoder(confRecord, c.Width(), c.Height())
		if err != nil {
			t.Fatal(err)
		}
	case "h265":
		c, err := h265parser.NewCodecDataFromAVCDecoderConfRecord(confRecord)
		if err != nil {
			t.Fatal(err)
		}
		d, err = NewH265Decoder(confRecord, c.Width(), c.Height())
		if err != nil {
			t.Fatal(err)
		}
	}

	vt, _ := AsVideoTransformer(d)
	defer vt.Close()

	vfile, _ := fsutil.CreateOrOpen(fmt.Sprintf(".tmp/%s.mp4", from))
	defer vfile.Close()

	ifile, _ := fsutil.CreateOrOpen(fmt.Sprintf(".tmp/%s.jpeg", from))
	defer ifile.Close()

	m := mp4.NewMuxer(vfile)
	init := false

	for i := start; i < end; i++ {
		raw, err := os.ReadFile(filepath.Join(frameRoot, fmt.Sprintf("%03d.frame", i)))
		if err != nil {
			t.Fatal(err)
		}

		ready, ret, err := vt.Transform(raw)
		if err != nil {
			t.Fatal(fmt.Sprintf("[%d]: %s", i, err))
		}

		if !ready {
			continue
		}

		if !init {
			err = m.WriteHeader([]av.CodecData{vt.CodecData()})
			if err != nil {
				t.Fatal(err)
			}
			init = true

			if ret.Frame != nil {
				_ = jpeg.Encode(ifile, ret.Frame, nil)
			}
		}

		err = m.WritePacket(av.Packet{
			Idx:        0,
			IsKeyFrame: ret.IsKeyFrame,
			Data:       ret.Data,
			Time:       time.Millisecond * time.Duration(i) * 40,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	if err := m.WriteTrailer(); err != nil {
		t.Fatal(err)
	}
}
