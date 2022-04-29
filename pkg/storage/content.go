package storage

import (
	"context"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage/content"
	"github.com/prometheus/prometheus/model/labels"
)

type Writer = content.Writer
type ReaderAt = content.ReaderAt

type Ingester = content.Ingester

type Provider interface {
	content.Provider
}

type Store interface {
	content.Store
	Manager
	Stop()
}

type Manager interface {
	Query(ctx context.Context, timeRange blob.TimeRange, userID string, matchers ...*labels.Matcher) ([]blob.Info, error)
	Info(ctx context.Context, ref blob.Ref) (*blob.Info, error)
	PutLabel(ctx context.Context, ref blob.Ref, labelName string, labelValue string) error
	DeleteLabel(ctx context.Context, ref blob.Ref, labelName string, labelValue string) error
}
