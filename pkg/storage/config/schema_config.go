package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/prometheus/common/model"
)

type SchemaConfig struct {
	Configs []PeriodConfig
}

func (cfg SchemaConfig) ExternalKey(ref blob.Ref) string {
	p, _ := cfg.SchemaForTime(ref.From)
	return ref.ExternalKey(p.Schema)
}

func (cfg SchemaConfig) SchemaForTime(t model.Time) (PeriodConfig, error) {
	for i := range cfg.Configs {
		if t >= cfg.Configs[i].From && (i+1 == len(cfg.Configs) || t < cfg.Configs[i+1].From) {
			return cfg.Configs[i], nil
		}
	}
	return PeriodConfig{}, fmt.Errorf("no schema config found for time %v", t)
}

type PeriodConfig struct {
	From        model.Time // used when working with config
	IndexTables PeriodicTableConfig
	Schema      string
	RowShards   uint32
	ObjectType  string // type of object client to use; if omitted, defaults to store.
}

type PeriodicTableConfig struct {
	Prefix string
	Period time.Duration
}

func (cfg *PeriodicTableConfig) tableForPeriod(i int64) string {
	return cfg.Prefix + strconv.Itoa(int(i))
}

func (cfg *PeriodicTableConfig) TableFor(t model.Time) string {
	if cfg.Period == 0 { // non-periodic
		return cfg.Prefix
	}
	periodSecs := int64(cfg.Period / time.Second)
	return cfg.tableForPeriod(t.Unix() / periodSecs)
}
