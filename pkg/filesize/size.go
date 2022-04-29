package filesize

type FileSize int64

const (
	Byte FileSize = 1
	KiB  FileSize = 1024 * Byte
	MiB  FileSize = 1024 * KiB
	GiB  FileSize = 1024 * MiB
	TiB  FileSize = 1024 * GiB
	PiB  FileSize = 1024 * TiB
)
