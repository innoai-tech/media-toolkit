package types

import (
	"bytes"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// openapi:strfmt label-filter-encoded
type Filter struct {
	Matchers []*labels.Matcher
}

func (f Filter) MarshalText() ([]byte, error) {
	b := bytes.NewBuffer(nil)
	b.WriteRune('{')
	for i := range f.Matchers {
		if i > 0 {
			b.WriteRune(',')
		}
		b.WriteString(f.Matchers[i].String())
	}
	b.WriteRune('}')

	return b.Bytes(), nil
}

func (f *Filter) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	matchers, err := parser.ParseMetricSelector(string(data))
	if err != nil {
		return err
	}
	for i := range matchers {
		m := matchers[i]
		if m.Name != labels.MetricName {
			f.Matchers = append(f.Matchers, m)
		}
	}
	return nil
}
