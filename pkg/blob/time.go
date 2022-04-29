package blob

import (
	"github.com/prometheus/common/model"
	"time"
)

type Time = model.Time

type TimeRange struct {
	From    model.Time `json:"from"`
	Through model.Time `json:"through"`
}

func SinceFrom(from Time, d time.Duration) TimeRange {
	return TimeRange{
		From:    from,
		Through: from.Add(d),
	}
}

func LastFrom(from Time, d time.Duration) TimeRange {
	return TimeRange{
		From:    from.Add(-d),
		Through: from,
	}
}

func Last(d time.Duration) TimeRange {
	return LastFrom(model.Now(), d)
}
