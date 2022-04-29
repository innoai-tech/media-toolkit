package blob

import (
	"time"

	"github.com/innoai-tech/media-toolkit/pkg/types"
)

func Date(year int, month time.Month, day int) types.Time {
	return types.TimeFromUnixNano(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).UnixNano())
}

func UnixDay(t types.Time) int64 {
	return t.UnixNano() / int64(24*time.Hour)
}
