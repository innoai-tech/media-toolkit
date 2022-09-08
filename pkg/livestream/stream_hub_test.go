package livestream_test

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/core"
	imageobserver "github.com/innoai-tech/media-toolkit/pkg/livestream/observer/image"
	videoobserver "github.com/innoai-tech/media-toolkit/pkg/livestream/observer/video"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
	. "github.com/octohelm/x/testing"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

var rootProject = ProjectRoot()

func TestStreamHub(t *testing.T) {
	l := stdr.New(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))

	streams, _ := core.LoadStreams(filepath.Join(rootProject, ".tmp/streams.json"))
	hub := livestream.NewStreamHub()

	s, err := storage.New(config.DefaultConfig)
	Expect(t, err, Be[error](nil))

	for _, s := range streams {
		hub.AddStream(logr.NewContext(context.Background(), l), s)
	}

	t.Run("Should add all stream", func(t *testing.T) {
		Expect(t, hub.List(), HaveLen[[]core.Stream](len(streams)))
	})

	t.Run("Take pic", func(t *testing.T) {
		ctx := logr.NewContext(context.Background(), l)

		wg := sync.WaitGroup{}

		for i := 0; i < 3; i++ {
			wg.Add(1)

			go func(i int) {
				defer wg.Done()
				o := imageobserver.New(s)

				_, err := hub.Subscribe(ctx, "1", o)
				Expect(t, err, Be[error](nil))

				err = <-o.Done()
				Expect(t, err, Be[error](nil))
				fmt.Println("done", i)
			}(i)
		}

		wg.Wait()
	})

	t.Run("Take video", func(t *testing.T) {
		ctx := logr.NewContext(context.Background(), l)

		wg := sync.WaitGroup{}

		for i := 0; i < 3; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				vo := videoobserver.New(s, func(o *videoobserver.Options) {
					o.MaxDuration = 5 * time.Second
				})

				_, err := hub.Subscribe(ctx, "1", vo)
				Expect(t, err, Be[error](nil))

				err = <-vo.Done()
				Expect(t, err, Be[error](nil))
			}()
		}

		wg.Wait()
	})
}
