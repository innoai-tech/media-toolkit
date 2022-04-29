package label

import (
	"github.com/innoai-tech/media-toolkit/pkg/types"
	"strings"
	"unicode/utf8"

	"github.com/innoai-tech/media-toolkit/pkg/blob"
)

func filterBlobsByTime(from, through types.Time, blobs []blob.Info) []blob.Info {
	filtered := make([]blob.Info, 0, len(blobs))
	for _, b := range blobs {
		if b.Through < from || through < b.From {
			continue
		}
		filtered = append(filtered, b)
	}
	return filtered
}

func filterBlobRefsByTime(from, through types.Time, blobs []blob.Ref) []blob.Ref {
	filtered := make([]blob.Ref, 0, len(blobs))
	for _, b := range blobs {
		if b.Through < from || through < b.From {
			continue
		}
		filtered = append(filtered, b)
	}
	return filtered
}

func uniqueStrings(cs []string) []string {
	if len(cs) == 0 {
		return []string{}
	}

	result := make([]string, 1, len(cs))
	result[0] = cs[0]
	i, j := 0, 1
	for j < len(cs) {
		if result[i] == cs[j] {
			j++
			continue
		}
		result = append(result, cs[j])
		i++
		j++
	}
	return result
}

func intersectStrings(left, right []string) []string {
	var (
		i, j   = 0, 0
		result = []string{}
	)
	for i < len(left) && j < len(right) {
		if left[i] == right[j] {
			result = append(result, left[i])
		}

		if left[i] < right[j] {
			i++
		} else {
			j++
		}
	}
	return result
}

// Bitmap used by func isRegexMetaCharacter to check whether a character needs to be escaped.
var regexMetaCharacterBytes [16]byte

// isRegexMetaCharacter reports whether byte b needs to be escaped.
func isRegexMetaCharacter(b byte) bool {
	return b < utf8.RuneSelf && regexMetaCharacterBytes[b%16]&(1<<(b/16)) != 0
}

func init() {
	for _, b := range []byte(`.+*?()|[]{}^$`) {
		regexMetaCharacterBytes[b%16] |= 1 << (b / 16)
	}
}

// FindSetMatches returns list of values that can be equality matched on.
// copied from Prometheus querier.go, removed check for Prometheus wrapper.
func FindSetMatches(pattern string) []string {
	escaped := false
	sets := []*strings.Builder{{}}
	for i := 0; i < len(pattern); i++ {
		if escaped {
			switch {
			case isRegexMetaCharacter(pattern[i]):
				sets[len(sets)-1].WriteByte(pattern[i])
			case pattern[i] == '\\':
				sets[len(sets)-1].WriteByte('\\')
			default:
				return nil
			}
			escaped = false
		} else {
			switch {
			case isRegexMetaCharacter(pattern[i]):
				if pattern[i] == '|' {
					sets = append(sets, &strings.Builder{})
				} else {
					return nil
				}
			case pattern[i] == '\\':
				escaped = true
			default:
				sets[len(sets)-1].WriteByte(pattern[i])
			}
		}
	}
	matches := make([]string, 0, len(sets))
	for _, s := range sets {
		if s.Len() > 0 {
			matches = append(matches, s.String())
		}
	}
	return matches
}
