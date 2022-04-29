package config

import (
	"github.com/prometheus/common/model"
	"time"
)

type Config struct {
	Schema  SchemaConfig
	Storage StorageConfig
}

var DefaultConfig = Config{
	Storage: StorageConfig{
		Root: ".tmp/mediadb",
		Compactor: CompactorConfig{
			After:        3,
			DiscardAfter: 7,
		},
	},
	Schema: SchemaConfig{
		Configs: []PeriodConfig{
			{
				From:      model.TimeFromUnixNano(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()),
				Schema:    "v1",
				RowShards: 16,
				IndexTables: PeriodicTableConfig{
					Period: 24 * time.Hour,
				},
			},
		},
	},
}
