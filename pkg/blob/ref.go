package blob

import (
	"github.com/opencontainers/go-digest"
	"path/filepath"
	"strconv"
)

var (
	DefaultUser = "0"
)

type Ref struct {
	TimeRange
	UserID string `json:"userID"`
	Alg    string `json:"alg"`
	Hex    string `json:"hex"`
}

func (r Ref) Digest() digest.Digest {
	return digest.Digest(r.Alg + ":" + r.Hex)
}

// ExternalKey <user_id>:<start>:<end>:<alg>:<hex>
func (r Ref) ExternalKey(schemaVersion string) string {
	b := make([]byte, 0, 6+6+64)
	b = append(b, []byte(r.UserID)...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(r.From), 16)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(r.Through), 16)
	b = append(b, ':')
	b = append(b, []byte(r.Alg)...)
	b = append(b, ':')
	b = append(b, []byte(r.Hex)...)
	return string(b)
}

// BlobPath blobs/<unix_day>/<alg>/<hex>
func (r Ref) BlobPath(root string) string {
	return filepath.Join(
		root,
		"blobs",
		strconv.FormatInt(UnixDay(r.From), 10),
		r.Alg,
		r.Hex,
	)
}
