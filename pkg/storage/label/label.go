package label

import (
	"github.com/innoai-tech/media-toolkit/pkg/blob"
)

type Store interface {
	IndexStore
	writer
}

var (
	MetricLabel = "_label"
)

type TimeRange = blob.TimeRange
