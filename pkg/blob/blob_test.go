package blob

import (
	"testing"

	. "github.com/octohelm/x/testing"
)

func TestBlob(t *testing.T) {
	b := FromString("x")
	Expect(t, b.Hex, Be("2d711642b726b04401627ca9fbac32f5c8530fb1903cc4db02258717921a4881"))

	parsed, err := ParseExternalKey(b.ExternalKey(""), "0")
	Expect(t, err, Be[error](nil))
	Expect(t, parsed, Equal(b))
}
