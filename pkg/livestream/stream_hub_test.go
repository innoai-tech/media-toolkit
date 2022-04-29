package livestream_test

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	imagewriter "github.com/innoai-tech/media-toolkit/pkg/livestream/writer/image"
	videowriter "github.com/innoai-tech/media-toolkit/pkg/livestream/writer/video"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	. "github.com/innoai-tech/media-toolkit/pkg/testutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var rootProject = ProjectRoot()

func loadStreams() []livestream.Stream {
	jsonRaw, _ := os.ReadFile(filepath.Join(rootProject, ".tmp/streams.json"))
	streams := make([]livestream.Stream, 0)
	_ = json.Unmarshal(jsonRaw, &streams)
	return streams
}

func TestStreamHub(t *testing.T) {
	l := stdr.New(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))

	t.Run("Hub", WithT(func() {
		streams := loadStreams()
		hub := livestream.NewStreamHub()

		s, err := storage.NewStore(filepath.Join(rootProject, ".tmp/screenshots"))
		So(err, ShouldBeNil)

		for _, s := range streams {
			hub.AddStream(s)
		}

		It("Should add all stream", func() {
			So(hub.List(), ShouldHaveLength, len(streams))
		})

		It("Hub", func() {
			ctx := logr.NewContext(context.Background(), l)

			iw := imagewriter.NewStreamWriter(s, imagewriter.Options{
				Filename: "{{ .Timestamp }}",
			})

			vw := videowriter.NewStreamWriter(s, videowriter.Options{
				Filename: "test",
			})

			mw := livestream.MultiStreamWriter(vw, iw)

			c, err := hub.ConnStream(ctx, "0", mw)
			So(err, ShouldBeNil)

			time.Sleep(5 * time.Second)

			_ = c.Close()
			_ = mw.Close()
		})
	}))

}
