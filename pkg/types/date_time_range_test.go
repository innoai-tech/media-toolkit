package types

import (
	. "github.com/octohelm/x/testing"
	"testing"
	"time"
)

func Date(year int, month time.Month, day int) Time {
	return TimeFromUnixNano(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).UnixNano())
}

func TestDateTimeRange(t *testing.T) {
	b := DateTimeRange{
		From: Date(2022, 2, 1),
		To:   Date(2022, 2, 5),
	}

	buf, err := b.MarshalText()
	Expect(t, err, Be[error](nil))
	Expect(t, string(buf), Be("2022-02-01T00:00:00Z..2022-02-05T00:00:00Z"))

	b2 := DateTimeRange{}
	err = b2.UnmarshalText([]byte("2022-02-01T00:00:00Z..2022-02-05T00:00:00Z"))
	Expect(t, err, Be[error](nil))

	Expect(t, b2, Equal(b))
}
