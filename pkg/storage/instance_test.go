package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
	"github.com/innoai-tech/media-toolkit/pkg/types"
	. "github.com/octohelm/x/testing"
	"github.com/pkg/errors"
)

func TestStore(t *testing.T) {
	s, err := New(config.DefaultConfig)
	Expect(t, err, Be[error](nil))
	defer func() {
		_ = s.Shutdown(context.Background())
	}()

	t.Run("Writer", func(t *testing.T) {
		w, err := s.Writer(context.Background())
		Expect(t, err, Be[error](nil))
		b := bytes.NewBufferString("1234")
		_, _ = io.Copy(w, b)

		s := blob.FromString("1234")

		err = w.Commit(
			context.Background(),
			int64(b.Len()),
			s.Digest(),
			blob.WithLabels(map[string][]string{
				"_media_type": {"text/plain"},
				"_tag":        {"test"},
			}),
		)
		Expect(t, err, Be[error](nil))
	})

	t.Run("Export", func(t *testing.T) {
		blobs, err := s.Query(context.Background(), blob.SinceFrom(types.Now(), 1*time.Hour), blob.DefaultUser)
		Expect(t, err, Be[error](nil))
		Expect(t, len(blobs) >= 1, Be(true))
		f, _ := os.CreateTemp(t.TempDir(), "")
		defer f.Close()
		_, err = ExportDataset(context.Background(), s, f, blobs)
		Expect(t, err, Be[error](nil))

		_ = os.RemoveAll(".tmp/dataset.tar.gz")
		_ = os.Rename(f.Name(), ".tmp/dataset.tar.gz")
	})

	t.Run("Reader", func(t *testing.T) {
		blobs, err := s.Query(context.Background(), blob.SinceFrom(types.Now(), 1*time.Hour), blob.DefaultUser)
		Expect(t, err, Be[error](nil))
		Expect(t, len(blobs) >= 1, Be(true))

		fmt.Println(blobs)

		t.Run("Provider", func(t *testing.T) {
			info, err := s.Info(context.Background(), blobs[0].Ref)
			Expect(t, err, Be[error](nil))

			buf := bytes.NewBuffer(nil)
			r, err := s.ReaderAt(context.Background(), info.Ref)
			Expect(t, err, Be[error](nil))
			_, _ = io.Copy(buf, io.NewSectionReader(r, 0, r.Size()))
			Expect(t, buf.String(), Be("1234"))

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
				Expect(t, errors.Is(err, ErrNotFound), Be(true))
			})
		})

	})
}
