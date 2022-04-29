package storage

import (
	"context"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
	"github.com/innoai-tech/media-toolkit/pkg/storage/content"
	contentlocal "github.com/innoai-tech/media-toolkit/pkg/storage/content/local"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label/index"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label/local"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/labels"
)

var (
	ErrNotFound       = errors.New("blob not found")
	ErrLabelImmutable = errors.New("immutable label")
)

func New(c config.Config) (Store, error) {
	indexClient, err := local.NewIndexClient(local.DBConfig{
		Directory: c.Storage.Root,
	})
	if err != nil {
		return nil, err
	}
	contentStore, err := contentlocal.NewStore(c.Storage.Root)
	if err != nil {
		return nil, err
	}

	return &store{
		c:           c,
		indexClient: indexClient,
		Store:       contentStore,
	}, nil
}

type store struct {
	c config.Config
	content.Store
	indexClient index.Client
}

func (s *store) Stop() {
	s.indexClient.Stop()
}

func (s *store) schemaFor(timeRange blob.TimeRange) (index.BlobStoreSchema, error) {
	pc, err := s.c.Schema.SchemaForTime(timeRange.From)
	if err != nil {
		return nil, err
	}

	return index.CreateSchema(pc)
}

func (s *store) labelIndexStoreFor(ctx context.Context, timeRange blob.TimeRange) (label.IndexStore, error) {
	schema, err := s.schemaFor(timeRange)
	if err != nil {
		return nil, err
	}

	return label.NewIndexStore(s.c.Schema, s.indexClient, schema), nil
}

func (s *store) labelWriterFor(ctx context.Context, timeRange blob.TimeRange) (label.Writer, error) {
	schema, err := s.schemaFor(timeRange)
	if err != nil {
		return nil, err
	}
	return label.NewWriter(s.c.Schema, s.indexClient, schema), nil
}

func (s *store) Query(ctx context.Context, timeRange blob.TimeRange, userID string, matchers ...*labels.Matcher) ([]blob.Info, error) {
	indexStore, err := s.labelIndexStoreFor(ctx, timeRange)
	if err != nil {
		return nil, err
	}
	return indexStore.GetBlobs(ctx, timeRange, userID, label.MetricLabel, matchers...)
}

func (s *store) Info(ctx context.Context, ref blob.Ref) (*blob.Info, error) {
	indexStore, err := s.labelIndexStoreFor(ctx, ref.TimeRange)
	if err != nil {
		return nil, err
	}
	blobs, err := indexStore.RefsToBlobs(ctx, []blob.Ref{ref}, label.MetricLabel)
	if err != nil {
		return nil, err
	}
	if len(blobs) == 0 {
		return nil, ErrNotFound
	}
	return &blobs[0], nil
}

func (s *store) Delete(ctx context.Context, ref blob.Ref) error {
	labelWriter, err := s.labelWriterFor(ctx, ref.TimeRange)
	if err != nil {
		return err
	}
	return labelWriter.DelOne(ctx, ref.TimeRange, label.MetricLabel, ref)
}

func (s *store) PutLabel(ctx context.Context, ref blob.Ref, labelName string, labelValue string) error {
	if len(labelName) != 0 && labelName[0] == '_' {
		return ErrLabelImmutable
	}

	labelWriter, err := s.labelWriterFor(ctx, ref.TimeRange)
	if err != nil {
		return err
	}

	return labelWriter.PutLabels(ctx, ref.TimeRange, label.MetricLabel, ref, blob.Labels{labelName: {labelValue}})
}

func (s *store) DeleteLabel(ctx context.Context, ref blob.Ref, labelName string, labelValue string) error {
	if len(labelName) != 0 && labelName[0] == '_' {
		return ErrLabelImmutable
	}

	labelWriter, err := s.labelWriterFor(ctx, ref.TimeRange)
	if err != nil {
		return err
	}

	return labelWriter.DelLabels(ctx, ref.TimeRange, label.MetricLabel, ref, blob.Labels{labelName: {labelValue}})
}

func (s *store) Writer(ctx context.Context, opts ...blob.Opt) (Writer, error) {
	w, err := s.Store.Writer(ctx, opts...)
	if err != nil {
		return nil, err
	}

	labelWriter, err := s.labelWriterFor(ctx, w.Info().TimeRange)
	if err != nil {
		return nil, err
	}

	return &writer{Writer: w, labelWriter: labelWriter}, nil
}

type writer struct {
	Writer
	labelWriter label.Writer
}

func (w *writer) Commit(ctx context.Context, size int64, expected digest.Digest, opts ...blob.Opt) error {
	err := w.Writer.Commit(ctx, size, expected, opts...)
	if err != nil {
		return err
	}
	return w.labelWriter.Put(ctx, label.MetricLabel, []blob.Info{w.Writer.Info()})
}
