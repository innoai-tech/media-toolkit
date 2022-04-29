package blob

import (
	"path/filepath"
	"strconv"

	"github.com/opencontainers/go-digest"
)

var (
	DefaultUser = "0"
)

// openapi:strfmt blob-ref-string
type RefString Ref

func (s RefString) MarshalText() ([]byte, error) {
	return []byte(Ref(s).ExternalKey("")), nil
}

func (s *RefString) UnmarshalText(b []byte) error {
	r, err := ParseExternalKey(string(b), "")
	if err != nil {
		return err
	}
	*s = RefString(r.Ref)
	return nil
}

func (s RefString) Ref() Ref {
	return Ref(s)
}

type Ref struct {
	TimeRange
	UserID string `json:"userID"`
	Alg    string `json:"-"`
	Hex    string `json:"-"`
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
