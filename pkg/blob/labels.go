package blob

import (
	"sort"
	"strconv"
	"strings"
)

type Labels map[string][]string

func (labels Labels) String() string {
	b := strings.Builder{}
	b.WriteRune('{')
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		for j, v := range labels[k] {
			if i > 0 || j > 0 {
				b.WriteRune(',')
			}
			b.WriteString(k)
			b.WriteRune('=')
			b.WriteString(strconv.Quote(v))
		}
	}

	b.WriteRune('}')

	return b.String()
}

const LabelDeleted = "__deleted__"
