package label_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label/index"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label/local"
	. "github.com/octohelm/x/testing"
	"github.com/prometheus/prometheus/model/labels"
	"golang.org/x/exp/rand"
)

var dayFrom = blob.Date(2022, 5, 15)
var sampleBlobs = make([]blob.Info, 10000)
var c = config.DefaultConfig

func TestStore(t *testing.T) {
	workingDir := t.TempDir()
	indexClient, err := local.NewIndexClient(local.DBConfig{
		Directory: workingDir,
	})
	Expect(t, err, Be[error](nil))
	defer func() {
		_ = indexClient.Shutdown(context.Background())
	}()

	schema, err := index.CreateSchema(c.Schema.Configs[0])
	Expect(t, err, Be[error](nil))

	t.Run("Write", func(t *testing.T) {
		w := label.NewWriter(c.Schema, indexClient, schema)
		err = w.Put(context.Background(), label.MetricLabel, sampleBlobs)
		Expect(t, err, Be[error](nil))
	})

	t.Run("Read", func(t *testing.T) {
		r := label.NewIndexStore(c.Schema, indexClient, schema)

		t.Run("GetBlobRefs", func(t *testing.T) {
			t.Run("Without Filter", func(t *testing.T) {
				blobRefs, err := r.GetBlobRefs(context.Background(), blob.SinceFrom(dayFrom, 24*time.Hour), blob.DefaultUser, label.MetricLabel)
				Expect(t, err, Be[error](nil))
				Expect(t, len(blobRefs), Be(10000))
			})

			t.Run("With Label Filter", func(t *testing.T) {
				filter, _ := labels.NewMatcher(labels.MatchNotEqual, "mediaType", "")
				blobRefs, err := r.GetBlobRefs(context.Background(), blob.SinceFrom(dayFrom, 24*time.Hour), blob.DefaultUser, label.MetricLabel, filter)
				Expect(t, err, Be[error](nil))
				Expect(t, len(blobRefs), Be(10000))
			})

			t.Run("With LabelValue Filter", func(t *testing.T) {
				filters := labels.MustNewMatcher(labels.MatchEqual, "mediaType", "text/plain")
				blobRefs, err := r.GetBlobRefs(context.Background(), blob.SinceFrom(dayFrom, 24*time.Hour), blob.DefaultUser, label.MetricLabel, filters)
				Expect(t, err, Be[error](nil))
				Expect(t, len(blobRefs), Be(5000))
			})
		})

		t.Run("GetBlob", func(t *testing.T) {
			filters := labels.MustNewMatcher(labels.MatchEqual, "mediaType", "text/plain")
			blobs, err := r.GetBlobs(context.Background(), blob.SinceFrom(dayFrom, 24*time.Hour), blob.DefaultUser, label.MetricLabel, filters)
			Expect(t, err, Be[error](nil))
			Expect(t, len(blobs), Be(5000))
			Expect(t, len(blobs[0].Labels), Be(3))
		})
	})
}

func init() {
	for i := range sampleBlobs {
		s := i % 2
		if s == 0 {
			b := blob.FromString(
				strconv.Itoa(rand.Int()),
				blob.WithFromThough(dayFrom),
				blob.WithLabels(map[string][]string{
					"instanceID": {strconv.Itoa(i % 3)},
					"mediaType":  {"text/plain"},
					"tag":        {"face", "secure"},
				}),
			)
			sampleBlobs[i] = b
		} else {
			b := blob.FromString(
				strconv.Itoa(rand.Int()),
				blob.WithFromThough(blob.Date(2022, 5, 16)),
				blob.WithLabels(map[string][]string{
					"instanceID": {strconv.Itoa(i % 3)},
					"mediaType":  {"text/html"},
					"tag":        {"wine"},
				}),
			)
			sampleBlobs[i] = b
		}
	}
}
