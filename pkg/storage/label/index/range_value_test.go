package index

import (
	. "github.com/octohelm/x/testing"
	"testing"
)

func TestRangeValue(t *testing.T) {
	rangeValue := RangeValueLabelValue(nil).EncodeRangeValue([]byte("_label"), []byte("x"), []byte("v"))
	Expect(t, string(rangeValue), Be("_label\x00x\x00TJRIXgwhrmxBzh3+e2v6zupato5AokdvUCCOUm9QYIA\x00\x002\x00"))

	rv, err := DecodeRangeValue(rangeValue)
	Expect(t, err, Be[error](nil))
	Expect(t, rv.(RangeValueLabelValue).MetricName(), Be("_label"))
	Expect(t, rv.(RangeValueLabelValue).LabelName(), Be("x"))
	Expect(t, string(readPart(rv.(RangeValueLabelValue), 2)), Be("TJRIXgwhrmxBzh3+e2v6zupato5AokdvUCCOUm9QYIA"))
}
