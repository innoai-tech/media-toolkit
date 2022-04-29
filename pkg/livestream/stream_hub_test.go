package livestream_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"

	"github.com/go-logr/stdr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	imageobserver "github.com/innoai-tech/media-toolkit/pkg/livestream/observer/image"
	videoobserver "github.com/innoai-tech/media-toolkit/pkg/livestream/observer/video"
	. "github.com/octohelm/x/testing"
)

type fake struct {
}

func (f fake) Done() <-chan error {
	return nil
}

func (f fake) Name() string {
	return "fake"
}

func (f fake) WritePacket(pkt livestream.Packet) {
}

func (f fake) Close() error {
	time.Sleep(3 * time.Second)
	return nil
}

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

	t.Run("Fake", func(t *testing.T) {
		ctx := logr.NewContext(context.Background(), l)
		allStart := sync.WaitGroup{}
		allDone := sync.WaitGroup{}

		for i := 0; i < 3; i++ {
			allStart.Add(1)
			allDone.Add(1)

			go func(i int) {
				c, err := hub.Subscribe(ctx, "1", &fake{})
				Expect(t, err, Be[error](nil))
				allStart.Done()
				time.Sleep(1 * time.Second)
				_ = c.Close()
				allDone.Done()
			}(i)
		}

		allStart.Wait()
		status, err := hub.Status(ctx, "1")
		Expect(t, err, Be[error](nil))
		Expect(t, status.Active, Be(true))
		Expect(t, status.Observers["fake"], Be(3))

		allDone.Wait()
		status, err = hub.Status(ctx, "1")
		Expect(t, err, Be[error](nil))
		//Expect(t, status.Active, Be(true))
		Expect(t, status.Observers["fake"], Be(0))
	})

	t.Run("TakePicture", func(t *testing.T) {
		ctx := logr.NewContext(context.Background(), l)
		allStart := sync.WaitGroup{}
		allDone := sync.WaitGroup{}

		for i := 0; i < 3; i++ {
			allStart.Add(1)
			allDone.Add(1)

			go func(i int) {
				iw := imageobserver.New(ctx, s)
				c, err := hub.Subscribe(ctx, "1", iw)
				Expect(t, err, Be[error](nil))
				allStart.Done()
				time.Sleep(1 * time.Second)
				_ = c.Close()
				allDone.Done()
			}(i)
		}

		allStart.Wait()
		status, err := hub.Status(ctx, "1")
		Expect(t, err, Be[error](nil))
		Expect(t, status.Observers["Image"], Be(3))

		allDone.Wait()

		status, err = hub.Status(ctx, "1")
		Expect(t, err, Be[error](nil))
		Expect(t, status.Observers["Image"], Be(0))
	})

	t.Run("TakeVideo", func(t *testing.T) {
		ctx := logr.NewContext(context.Background(), l)

		vo := videoobserver.New(ctx, s)
		ss, err := hub.Subscribe(ctx, "1", vo)
		Expect(t, err, Be[error](nil))
		time.Sleep(5 * time.Second)
		err = ss.Close()
		Expect(t, err, Be[error](nil))
	})
}
