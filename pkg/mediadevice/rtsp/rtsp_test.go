package rtsp

import (
	"context"
	"fmt"
	"github.com/innoai-tech/media-toolkit/pkg/util/ioutil"
	"image"
	"image/jpeg"
	"io"
	"path/filepath"
	"sync"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/innoai-tech/media-toolkit/pkg/format"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/core"
	"github.com/innoai-tech/media-toolkit/pkg/util/fsutil"
	testingx "github.com/octohelm/x/testing"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/x264"
	"github.com/pion/mediadevices/pkg/io/video"
)

var rootProject = testingx.ProjectRoot()

func TestRTSP(t *testing.T) {
	streams, _ := core.LoadStreams(filepath.Join(rootProject, ".tmp/streams.json"))

	videoSource, err := Open(context.Background(), streams[0].Rtsp, streams[0].ID)
	testingx.Expect(t, err, testingx.Be[error](nil))
	defer videoSource.Close()

	t.Run("Save as mp4", func(t *testing.T) {
		x264Params, _ := x264.NewParams()
		x264Params.Preset = x264.PresetMedium
		x264Params.BitRate = 4_000_000 // 4mbps

		codecSelector := mediadevices.NewCodecSelector(
			mediadevices.WithVideoEncoders(&x264Params),
		)

		videoTrack := mediadevices.NewVideoTrack(videoSource, codecSelector)

		f, err := fsutil.CreateOrOpen(".tmp/x.mp4")
		testingx.Expect(t, err, testingx.Be[error](nil))
		defer f.Close()

		er, err := videoTrack.NewEncodedReader(x264Params.RTPCodec().MimeType)
		testingx.Expect(t, err, testingx.Be[error](nil))

		r := format.NewRecorder(f, er, "2")
		defer r.Close()

		var info *format.Info

		for count := 0; count < 200; count++ {
			info, err = r.Record()
			if err != nil {
				testingx.Expect(t, err, testingx.Be[error](nil))
			}
		}

		spew.Dump(info)
	})

	t.Run("Save as images", func(t *testing.T) {
		r := video.NewBroadcaster(videoSource, nil).NewReader(true)
		wg := sync.WaitGroup{}

		for count := 0; count < 100; count++ {
			if count%25 == 0 {
				continue
			}

			img, _, err := r.Read()
			if err != nil {
				spew.Dump(err)
				break
			}

			wg.Add(1)

			go func(img image.Image, i int) {
				defer wg.Done()

				_ = ioutil.Pipe(
					func(w io.Writer) error {
						return jpeg.Encode(w, img, nil)
					},
					func(r io.Reader) error {
						f, _ := fsutil.CreateOrOpen(fmt.Sprintf(".tmp/frame/%d.jpeg", i))
						defer f.Close()
						_, err := io.Copy(f, r)
						return err
					},
				)
			}(img, count)
		}

		wg.Wait()
	})
}
