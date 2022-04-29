package types

import (
	"testing"

	. "github.com/octohelm/x/testing"
	"github.com/prometheus/prometheus/model/labels"
)

func TestFilter(t *testing.T) {
	f := Filter{}
	err := f.UnmarshalText([]byte(`{ _media_type = "image/jpeg", _device_id = "1" }`))
	Expect(t, err, Be[error](nil))
	Expect(t, f.Matchers, Equal([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "_media_type", "image/jpeg"),
		labels.MustNewMatcher(labels.MatchEqual, "_device_id", "1"),
	}))

	data, _ := f.MarshalText()
	Expect(t, string(data), Be(`{_media_type="image/jpeg",_device_id="1"}`))
}
