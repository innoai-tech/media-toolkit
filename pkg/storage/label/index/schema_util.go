package index

import (
	"encoding/base64"
)

func makeQuerySlice(n uint32, create func(i uint32) Query) []Query {
	list := make([]Query, n)
	for i := range list {
		list[i] = create(uint32(i))
	}
	return list
}

func buildRangeValue(extra int, ss ...[]byte) []byte {
	length := extra
	for _, s := range ss {
		length += len(s) + 1
	}
	output, i := make([]byte, length), 0
	for _, s := range ss {
		i += copy(output[i:], s) + 1
	}
	return output
}

func rangeValuePrefix(ss ...[]byte) []byte {
	return buildRangeValue(0, ss...)
}

func encodeBase64Bytes(bytes []byte) []byte {
	encodedLen := base64.RawStdEncoding.EncodedLen(len(bytes))
	encoded := make([]byte, encodedLen)
	base64.RawStdEncoding.Encode(encoded, bytes)
	return encoded
}
