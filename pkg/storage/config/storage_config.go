package config

type StorageConfig struct {
	Root      string
	Compactor CompactorConfig
}

type CompactorConfig struct {
	After        uint
	DiscardAfter uint
}
