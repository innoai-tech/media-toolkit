package blob

import (
	"time"

	"github.com/prometheus/common/model"
)

func Date(year int, month time.Month, day int) model.Time {
	return model.TimeFromUnixNano(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).UnixNano())
}

func UnixDay(t model.Time) int64 {
	return t.UnixNano() / int64(24*time.Hour)
}
