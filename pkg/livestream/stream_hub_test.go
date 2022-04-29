package livestream_test

import (
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/stdr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	imageobserver "github.com/innoai-tech/media-toolkit/pkg/livestream/observer/image"
	videoobserver "github.com/innoai-tech/media-toolkit/pkg/livestream/observer/video"
	. "github.com/octohelm/x/testing"
)

var rootProject = ProjectRoot()

func TestStreamHub(t *testing.T) {
	l := stdr.New(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))

	streams, _ := livestream.LoadStreams(filepath.Join(rootProject, ".tmp/streams.json"))
	hub := livestream.NewStreamHub()

	s, err := storage.New(config.DefaultConfig)
	Expect(t, err, Be[error](nil))

	for _, s := range streams {
		hub.AddStream(s)
	}

	t.Run("Should add all stream", func(t *testing.T) {
		Expect(t, hub.List(), HaveLen[[]livestream.Stream](len(streams)))
	})

	t.Run("TakePicture", func(t *testing.T) {
		iw := imageobserver.New(s, imageobserver.Options{
			Filename: "{{ .Timestamp }}",
		})
		c, err := hub.Subscribe("0", iw)
		Expect(t, err, Be[error](nil))
		time.Sleep(3 * time.Second)
		_ = c.Close()
	})

	t.Run("TakeVideo", func(t *testing.T) {
		vo := videoobserver.New(s, videoobserver.Options{
			Filename: "{{ .Timestamp }}",
			Log:      &l,
		})
		ss, err := hub.Subscribe("1", vo)
		Expect(t, err, Be[error](nil))
		time.Sleep(5 * time.Second)
		_ = ss.Close()
	})
}
