package index

import (
	"fmt"
	"github.com/innoai-tech/media-toolkit/pkg/types"
	"strings"
	"time"

	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
)

const (
	secondsInDay      = int64(24 * time.Hour / time.Second)
	millisecondsInDay = int64(24 * time.Hour / time.Millisecond)
)

type Bucket struct {
	TableName  string
	From       uint32
	Through    uint32
	HashKey    string
	BucketSize uint32
}

func (bucket *Bucket) HashValuePrefixFor(shard uint32, keys ...string) string {
	b := &strings.Builder{}
	_, _ = fmt.Fprintf(b, "%02d", shard)
	b.WriteRune(':')
	b.WriteString(bucket.HashKey)
	for i := range keys {
		b.WriteRune(':')
		b.WriteString(keys[i])
	}
	return b.String()
}

func dailyBuckets(cfg config.PeriodConfig) rangeBucketFunc {
	return func(from, through types.Time, userID string, each func(bucket *Bucket) error) error {
		var (
			fromDay    = from.Unix() / secondsInDay
			throughDay = through.Unix() / secondsInDay
		)

		for i := fromDay; i <= throughDay; i++ {
			relativeFrom := max(int64(0), int64(from)-i*millisecondsInDay)
			relativeThrough := min(millisecondsInDay, int64(through)-(i*millisecondsInDay))

			err := each(&Bucket{
				From:       uint32(relativeFrom),
				Through:    uint32(relativeThrough),
				TableName:  cfg.IndexTables.TableFor(types.TimeFromUnix(i * secondsInDay)),
				HashKey:    fmt.Sprintf("%s:d%d", userID, i),
				BucketSize: uint32(millisecondsInDay),
			})

			if err != nil {
				return err
			}
		}

		return nil
	}
}
