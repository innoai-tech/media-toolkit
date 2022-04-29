package index

import (
	"crypto/sha256"

	"github.com/pkg/errors"
)

var (
	ErrInvalidRangeValue       = errors.New("invalid range value")
	ErrRangeValuePartNotExists = errors.New("range value part not exists")
)

// EncodeRangeValue a complete key including type marker (which goes at the end)
func EncodeRangeValue(rv RangeValue, ss ...[]byte) []byte {
	output := buildRangeValue(2, ss...)
	output[len(output)-2] = rv.RangeKey()
	return output
}

var rangeValues = []RangeValue{
	RangeValueMetricName{},
	RangeValueLabelValue{},
	RangeValueLabelValueBlob{},
}

func DecodeRangeValue(rv []byte) (RangeValue, error) {
	if len(rv) < 3 || !(rv[len(rv)-1] == 0 && rv[len(rv)-3] == 0) {
		return nil, errors.Wrapf(ErrInvalidRangeValue, "%q", rv)
	}

	rk := rv[len(rv)-2]

	for _, r := range rangeValues {
		if r.RangeKey() == rk {
			return r.New(rv), nil
		}
	}

	return RangeValueUnknown(rv), nil
}

type RangeValue interface {
	New(rv []byte) RangeValue
	RangeKey() byte
}

type RangeValueUnknown []byte

func (RangeValueUnknown) RangeKey() byte {
	return 0
}

func (RangeValueUnknown) New(rv []byte) RangeValue {
	return RangeValueUnknown(rv)
}

type RangeValueMetricName []byte

func (RangeValueMetricName) RangeKey() byte {
	return '1'
}

func (RangeValueMetricName) New(rv []byte) RangeValue {
	return RangeValueMetricName(rv)
}

func (r RangeValueMetricName) EncodeRangeValue(blobID []byte) []byte {
	return EncodeRangeValue(r, blobID, nil)
}

func (r RangeValueMetricName) BlobID() string {
	return string(readPart(r, 0))
}

type RangeValueLabelValue []byte

func (RangeValueLabelValue) RangeKey() byte {
	return '2'
}

func (RangeValueLabelValue) New(rv []byte) RangeValue {
	return RangeValueLabelValue(rv)
}

func (r RangeValueLabelValue) EncodeRangeValue(metricName []byte, labelName []byte, labelValue []byte) []byte {
	return EncodeRangeValue(r, metricName, labelName, sha256bytes(labelValue), nil)
}

func (r RangeValueLabelValue) MetricName() string {
	return string(readPart(r, 0))
}

func (r RangeValueLabelValue) LabelName() string {
	return string(readPart(r, 1))
}

type RangeValueLabelValueBlob []byte

func (RangeValueLabelValueBlob) RangeKey() byte {
	return '3'
}

func (RangeValueLabelValueBlob) New(rv []byte) RangeValue {
	return RangeValueLabelValueBlob(rv)
}

func (r RangeValueLabelValueBlob) EncodeRangeValue(labelValue []byte, blobID []byte) []byte {
	return EncodeRangeValue(r, sha256bytes(labelValue), blobID, nil)
}

func (r RangeValueLabelValueBlob) BlobID() string {
	return string(readPart(r, 1))
}

func sha256bytes(v []byte) []byte {
	h := sha256.Sum256(v)
	return encodeBase64Bytes(h[:])
}

func readPart(rangeValue []byte, idx int) []byte {
	count := 0
	left := 0

	for i, b := range rangeValue {
		if b == 0 {
			if count == idx {
				return rangeValue[left:i]
			}
			count++
			left = i + 1
		}
	}
	return rangeValue[left:]
}
