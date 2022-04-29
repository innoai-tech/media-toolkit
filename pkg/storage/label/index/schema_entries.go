package index

import (
	"encoding/binary"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"strings"
)

type v1Entries struct {
	rowShards uint32
}

func (s *v1Entries) GetMetricLabelValues(bucket *Bucket, metricName string, blobID string) ([]Query, error) {
	return []Query{{
		TableName:        bucket.TableName,
		HashValue:        blobID,
		RangeValuePrefix: rangeValuePrefix([]byte(metricName)),
	}}, nil
}

func (s *v1Entries) GetReadQueriesForMetric(bucket *Bucket, metricName string) ([]Query, error) {
	return makeQuerySlice(s.rowShards, func(shard uint32) Query {
		return Query{
			TableName: bucket.TableName,
			HashValue: bucket.HashValuePrefixFor(shard, metricName),
		}
	}), nil
}

func (s *v1Entries) GetReadQueriesForMetricLabel(bucket *Bucket, metricName string, labelName string) ([]Query, error) {
	return makeQuerySlice(s.rowShards, func(shard uint32) Query {
		return Query{
			TableName: bucket.TableName,
			HashValue: bucket.HashValuePrefixFor(shard, metricName, labelName),
		}
	}), nil
}

func (s *v1Entries) GetReadQueriesForMetricLabelValue(bucket *Bucket, metricName string, labelName string, labelValue string) ([]Query, error) {
	labelValueBytes := []byte(labelValue)

	return makeQuerySlice(s.rowShards, func(shard uint32) Query {
		return Query{
			TableName:        bucket.TableName,
			HashValue:        bucket.HashValuePrefixFor(shard, metricName, labelName),
			RangeValuePrefix: rangeValuePrefix(sha256bytes(labelValueBytes)),
			ValueEqual:       labelValueBytes,
		}
	}), nil
}

func (s *v1Entries) GetLabelWriteEntries(bucket *Bucket, metricName string, labels blob.Labels, blobID string) ([]Entry, error) {
	// read first 32 bits of the hex and use this to calculate the shard
	shard := uint32(0)
	i := strings.LastIndex(blobID, ":")
	if i > 0 {
		shard = binary.BigEndian.Uint32([]byte(blobID[i+1:])) % s.RowShard()
	}

	entries := []Entry{{
		TableName:  bucket.TableName,
		HashValue:  bucket.HashValuePrefixFor(shard, metricName),
		RangeValue: RangeValueMetricName(nil).EncodeRangeValue([]byte(blobID)),
		Value:      []byte{0},
	}}

	for labelName, vv := range labels {
		for _, labelValue := range vv {
			labelValueBytes := []byte(labelValue)

			entries = append(entries,
				// id | metricName label_name Hash(label_value) _ rk = label_value
				Entry{
					TableName:  bucket.TableName,
					HashValue:  blobID,
					RangeValue: RangeValueLabelValue(nil).EncodeRangeValue([]byte(metricName), []byte(labelName), labelValueBytes),
					Value:      labelValueBytes,
				},
				// HashValuePrefixFor(label_name) | Hash(label_value) blob_id _ rk = label_value
				Entry{
					TableName:  bucket.TableName,
					HashValue:  bucket.HashValuePrefixFor(shard, metricName, labelName),
					RangeValue: RangeValueLabelValueBlob(nil).EncodeRangeValue(labelValueBytes, []byte(blobID)),
					Value:      labelValueBytes,
				},
			)
		}
	}

	return entries, nil
}

func (s *v1Entries) RowShard() uint32 {
	if s.rowShards == 0 {
		s.rowShards = 16
	}
	return s.rowShards
}
