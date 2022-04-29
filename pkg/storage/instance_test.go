package storage

import (
	"bytes"
	"context"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
	. "github.com/octohelm/x/testing"
	"github.com/prometheus/common/model"
	"io"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	s, err := New(config.DefaultConfig)
	Expect(t, err, Be[error](nil))
	defer s.Stop()

	t.Run("Writer", func(t *testing.T) {
		w, err := s.Writer(context.Background())
		Expect(t, err, Be[error](nil))
		_, _ = io.Copy(w, bytes.NewBufferString("1234"))
		err = w.Commit(context.Background(), 0, "", blob.WithLabels(map[string][]string{
			"_media_type": {"text/plain"},
		}))
		Expect(t, err, Be[error](nil))
	})

	t.Run("Reader", func(t *testing.T) {
		blobs, err := s.Query(context.Background(), blob.SinceFrom(model.Now(), 1*time.Hour), blob.DefaultUser)
		Expect(t, err, Be[error](nil))
		Expect(t, len(blobs) > 1, Be(true))

		t.Run("Provider", func(t *testing.T) {
			info, err := s.Info(context.Background(), blobs[0].Ref)
			Expect(t, err, Be[error](nil))

			t.Run("PutLabel", func(t *testing.T) {
				err = s.PutLabel(context.Background(), info.Ref, "tag", "v")
				Expect(t, err, Be[error](nil))

				updated, err := s.Info(context.Background(), info.Ref)
				Expect(t, err, Be[error](nil))
				Expect(t, updated.Labels["tag"], Equal([]string{"v"}))
			})

			t.Run("DeleteLabel", func(t *testing.T) {
				err = s.DeleteLabel(context.Background(), info.Ref, "tag", "v")
				Expect(t, err, Be[error](nil))

				updated2, err := s.Info(context.Background(), info.Ref)
				Expect(t, err, Be[error](nil))
				_, existsTag := updated2.Labels["tag"]
				Expect(t, existsTag, Be(false))
			})

			t.Run("DeleteRecord", func(t *testing.T) {
				err = s.Delete(context.Background(), info.Ref)
				Expect(t, err, Be[error](nil))

				_, err := s.Info(context.Background(), info.Ref)
				Expect(t, err, Be[error](ErrNotFound))
			})
		})

	})
}
