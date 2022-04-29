package blob

import (
	"time"
	_ "time/tzdata"

	"github.com/innoai-tech/media-toolkit/pkg/types"
)

type Time = types.Time

type TimeRange struct {
	From    types.Time `json:"from"`
	Through types.Time `json:"through"`
}

func SinceFrom(from types.Time, d time.Duration) TimeRange {
	return TimeRange{
		From:    from,
		Through: from.Add(d),
	}
}

func LastFrom(from types.Time, d time.Duration) TimeRange {
	return TimeRange{
		From:    from.Add(-d),
		Through: from,
	}
}

func Last(d time.Duration) TimeRange {
	return LastFrom(types.Now(), d)
}
