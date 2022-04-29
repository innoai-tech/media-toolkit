package types

import (
	"bytes"
	"time"

	"github.com/pkg/errors"
)

// openapi:strfmt date-time-range
type DateTimeRange struct {
	From Time
	To   Time
}

func (tr DateTimeRange) IsZero() bool {
	return tr.From == 0 && tr.To == 0
}

func (tr DateTimeRange) MarshalText() ([]byte, error) {
	data := make([]byte, 0)

	if tr.From != 0 {
		data = append(data, []byte(tr.From.Time().In(time.UTC).Format(time.RFC3339))...)
	}
	data = append(data, []byte("..")...)

	if tr.To != 0 {
		data = append(data, []byte(tr.To.Time().In(time.UTC).Format(time.RFC3339))...)
	}

	return data, nil
}

func (tr *DateTimeRange) UnmarshalText(data []byte) error {
	parts := bytes.Split(data, []byte(".."))
	if len(parts) != 2 {
		return errors.Errorf("invalid data time range value %q", data)
	}

	from, err := time.Parse(time.RFC3339, string(parts[0]))
	if err != nil {
		return err
	}

	to, err := time.Parse(time.RFC3339, string(parts[1]))
	if err != nil {
		return err
	}

	tr.From = TimeFromUnixNano(from.UnixNano())
	tr.To = TimeFromUnixNano(to.UnixNano())

	return nil
}
