package index

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
	"github.com/pkg/errors"
)

type BlobStoreSchema interface {
	GetMetricLabelValues(timeRange blob.TimeRange, userID string, metricName string, blobID string) ([]Query, error)

	GetReadQueriesForMetric(timeRange blob.TimeRange, userID string, metricName string) ([]Query, error)
	GetReadQueriesForMetricLabel(timeRange blob.TimeRange, userID string, metricName string, labelName string) ([]Query, error)
	GetReadQueriesForMetricLabelValue(timeRange blob.TimeRange, userID string, metricName string, labelName string, labelValue string) ([]Query, error)

	GetCacheKeysAndLabelWriteEntries(timeRange blob.TimeRange, userID string, metricName string, blobID string, labels blob.Labels) ([]string, [][]Entry, error)
}

type BlobStoreEntries interface {
	GetMetricLabelValues(bucket *Bucket, metricName string, blobID string) ([]Query, error)

	GetReadQueriesForMetric(bucket *Bucket, metricName string) ([]Query, error)
	GetReadQueriesForMetricLabel(bucket *Bucket, metricName string, labelName string) ([]Query, error)
	GetReadQueriesForMetricLabelValue(bucket *Bucket, metricName string, labelName string, labelValue string) ([]Query, error)

	GetLabelWriteEntries(bucket *Bucket, metricName string, labels blob.Labels, blobID string) ([]Entry, error)
}

var (
	errInvalidSchemaVersion = errors.New("invalid schema version")
	errInvalidTablePeriod   = errors.New("the table period must be a multiple of 24h (1h for schema v1)")
)

var (
	ErrNotSupported = errors.New("not supported")
)

func CreateSchema(cfg config.PeriodConfig) (BlobStoreSchema, error) {
	buckets, bucketsPeriod := dailyBuckets(cfg), 24*time.Hour

	// Ensure the tables period is a multiple of the bucket period
	if cfg.IndexTables.Period > 0 && cfg.IndexTables.Period%bucketsPeriod != 0 {
		return nil, errInvalidTablePeriod
	}

	switch cfg.Schema {
	case "v1":
		return newBlobStoreSchema(&v1Entries{rowShards: cfg.RowShards}, buckets), nil
	default:
		return nil, errInvalidSchemaVersion
	}
}

type rangeBucketFunc func(from, through blob.Time, userID string, each func(bucket *Bucket) error) error

func newBlobStoreSchema(entries BlobStoreEntries, rangeBucket rangeBucketFunc) *blobStoreSchema {
	return &blobStoreSchema{
		rangeBucket: rangeBucket,
		entries:     entries,
	}
}

type blobStoreSchema struct {
	entries     BlobStoreEntries
	rangeBucket rangeBucketFunc
}

func (s *blobStoreSchema) makeQuerySliceBuckets(timeRange blob.TimeRange, userID string, each func(bucket *Bucket) ([]Query, error)) ([]Query, error) {
	var ret []Query

	err := s.rangeBucket(timeRange.From, timeRange.Through, userID, func(bucket *Bucket) error {
		queries, err := each(bucket)
		if err != nil {
			return err
		}
		ret = append(ret, queries...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *blobStoreSchema) GetMetricLabelValues(timeRange blob.TimeRange, userID string, metricName string, blobID string) ([]Query, error) {
	return s.makeQuerySliceBuckets(timeRange, userID, func(bucket *Bucket) ([]Query, error) {
		return s.entries.GetMetricLabelValues(bucket, metricName, blobID)
	})
}

func (s *blobStoreSchema) GetReadQueriesForMetric(timeRange blob.TimeRange, userID string, metricName string) ([]Query, error) {
	return s.makeQuerySliceBuckets(timeRange, userID, func(bucket *Bucket) ([]Query, error) {
		return s.entries.GetReadQueriesForMetric(bucket, metricName)
	})
}

func (s *blobStoreSchema) GetReadQueriesForMetricLabel(timeRange blob.TimeRange, userID string, metricName string, labelName string) ([]Query, error) {
	return s.makeQuerySliceBuckets(timeRange, userID, func(bucket *Bucket) ([]Query, error) {
		return s.entries.GetReadQueriesForMetricLabel(bucket, metricName, labelName)
	})
}

func (s *blobStoreSchema) GetReadQueriesForMetricLabelValue(timeRange blob.TimeRange, userID string, metricName string, labelName string, labelValue string) ([]Query, error) {
	return s.makeQuerySliceBuckets(timeRange, userID, func(bucket *Bucket) ([]Query, error) {
		return s.entries.GetReadQueriesForMetricLabelValue(bucket, metricName, labelName, labelValue)
	})
}

func (s *blobStoreSchema) GetCacheKeysAndLabelWriteEntries(timeRange blob.TimeRange, userID string, metricName string, blobID string, labels blob.Labels) ([]string, [][]Entry, error) {
	var keys []string
	var indexEntries [][]Entry

	err := s.rangeBucket(timeRange.From, timeRange.Through, userID, func(bucket *Bucket) error {
		key := strings.Join([]string{bucket.TableName, bucket.HashKey, blobID}, "-")

		// This is just encoding to remove invalid characters so that we can put them in memcache.
		// We're not hashing them as the length of the key is well within memcache bounds. tableName + userid + day + 32Byte(seriesID)
		key = hex.EncodeToString([]byte(key))
		keys = append(keys, key)

		entries, err := s.entries.GetLabelWriteEntries(bucket, metricName, labels, blobID)
		if err != nil {
			return err
		}
		indexEntries = append(indexEntries, entries)

		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return keys, indexEntries, nil
}
