package blob

import (
	"bytes"
	"github.com/innoai-tech/media-toolkit/pkg/types"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

var (
	ErrUserNotMatch = errors.New("user not match")
)

func ParseExternalKey(ref string, expectUserID string) (Info, error) {
	userIdx := strings.Index(ref, ":")
	if userIdx == -1 || userIdx+1 >= len(ref) {
		return Info{}, errors.Errorf("invalid ref `%s`", ref)
	}
	userID := ref[:userIdx]
	// When empty, will skip check
	if expectUserID != "" && expectUserID != userID {
		return Info{}, errors.WithStack(ErrUserNotMatch)
	}
	hexParts := ref[userIdx+1:]
	partsBytes := unsafeGetBytes(hexParts)

	// Parse from
	h, i := readOneHexPart(partsBytes)
	if i == 0 || i+1 >= len(partsBytes) {
		return Info{}, errors.Wrap(errors.Errorf("invalid ref `%s`", ref), "decoding start")
	}
	from, err := strconv.ParseInt(unsafeGetString(h), 16, 64)
	if err != nil {
		return Info{}, errors.Wrap(err, "parsing start")
	}
	partsBytes = partsBytes[i+1:]

	// Parse through
	h, i = readOneHexPart(partsBytes)
	if i == 0 || i+1 >= len(partsBytes) {
		return Info{}, errors.Wrap(errors.Errorf("invalid blobid `%s`", ref), "decoding through")
	}
	through, err := strconv.ParseInt(unsafeGetString(h), 16, 64)
	if err != nil {
		return Info{}, errors.Wrap(err, "parsing through")
	}
	partsBytes = partsBytes[i+1:]

	// Get alg
	h, i = readOneHexPart(partsBytes)
	if i == 0 || i+1 >= len(partsBytes) {
		return Info{}, errors.Wrap(errors.Errorf("invalid ref `%s`", ref), "decoding alg")
	}
	alg := unsafeGetString(h)
	partsBytes = partsBytes[i+1:]

	// Get hex
	h, i = readOneHexPart(partsBytes)
	if i == 0 {
		h = partsBytes
	}
	if len(h) != 64 {
		return Info{}, errors.Wrap(errors.Errorf("invalid ref `%s`", ref), "decoding hex")
	}
	hex := unsafeGetString(h)

	return Info{
		Ref: Ref{
			UserID: userID,
			Alg:    alg,
			Hex:    hex,
			TimeRange: TimeRange{
				From:    types.Time(from),
				Through: types.Time(through),
			},
		},
	}, nil
}

func readOneHexPart(hex []byte) (part []byte, i int) {
	for i < len(hex) {
		if hex[i] != ':' && hex[i] != '/' {
			i++
			continue
		}
		return hex[:i], i
	}
	return nil, 0
}

func unsafeGetBytes(s string) []byte {
	var buf []byte
	p := unsafe.Pointer(&buf)
	*(*string)(p) = s
	(*reflect.SliceHeader)(p).Cap = len(s)
	return buf
}

func unsafeGetString(buf []byte) string {
	return *((*string)(unsafe.Pointer(&buf)))
}

type Opt func(b *Info)

func WithUserId(userID string) Opt {
	return func(b *Info) {
		b.UserID = userID
	}
}

func WithFromThough(from types.Time, through ...types.Time) Opt {
	return func(b *Info) {
		b.From = from

		if len(through) == 0 {
			b.Through = from
		} else {
			b.Through = through[0]
		}
	}
}

func WithLabels(lbs Labels) Opt {
	return func(b *Info) {
		if b.Labels == nil {
			b.Labels = map[string][]string{}
		}
		for k, vv := range lbs {
			b.Labels[k] = append(b.Labels[k], vv...)
		}
	}
}

func FromDigest(dgst digest.Digest, options ...Opt) Info {
	b := Info{}
	b.UserID = DefaultUser
	b.Alg = string(dgst.Algorithm())
	b.Hex = dgst.Hex()
	for _, o := range options {
		o(&b)
	}
	return b
}

func FromReader(r io.Reader, options ...Opt) (Info, error) {
	now := types.Now()
	d, err := digest.FromReader(r)
	if err != nil {
		return Info{}, err
	}
	info := FromDigest(d, append([]Opt{WithFromThough(now, now)}, options...)...)
	return info, nil
}

func FromString(s string, options ...Opt) Info {
	return FromBytes([]byte(s), options...)
}

func FromBytes(data []byte, options ...Opt) Info {
	b, _ := FromReader(bytes.NewBuffer(data), options...)
	return b
}

type Info struct {
	Ref
	Labels Labels `json:"labels"`
	RefKey string `json:"ref"`
}

type Infos []Info

func (infos Infos) Len() int {
	return len(infos)
}

func (infos Infos) Less(i, j int) bool {
	return infos[i].From > infos[j].From
}

func (infos Infos) Swap(i, j int) {
	infos[i], infos[j] = infos[j], infos[i]
}
