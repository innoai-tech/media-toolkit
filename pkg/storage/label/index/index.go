package index

import (
	"bytes"
	"context"
)

type Query struct {
	TableName        string
	HashValue        string
	RangeValuePrefix []byte
	RangeValueStart  []byte
	ValueEqual       []byte
}

type Entry struct {
	TableName  string
	HashValue  string
	RangeValue []byte // For writes, RangeValue will always be set.
	Value      []byte
}

func (e *Entry) Key() []byte {
	return bytes.Join([][]byte{[]byte(e.HashValue), e.RangeValue}, []byte{0})
}

type Client interface {
	Stop()
	NewWriteBatch() WriteBatch
	BatchWrite(context.Context, WriteBatch) error
	QueryPages(ctx context.Context, queries []Query, callback QueryPagesCallback) error
}

type WriteBatch interface {
	Add(entry Entry)
	Delete(entry Entry)
}

type QueryPagesCallback func(ReadBatchResult, Query) error

type ReadBatchResult interface {
	Iterator() ReadBatchIterator
}

type ReadBatchIterator interface {
	Next() bool
	Entry() *Entry
}
