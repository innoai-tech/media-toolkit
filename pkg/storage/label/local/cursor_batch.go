package local

import (
	"bytes"
	"sync"

	"github.com/cockroachdb/pebble"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label/index"
)

var batchPool = sync.Pool{
	New: func() any {
		return &cursorBatch{
			start:     bytes.NewBuffer(make([]byte, 0, 1024)),
			rowPrefix: bytes.NewBuffer(make([]byte, 0, 1024)),
		}
	},
}

type cursorBatch struct {
	iter      *pebble.Iterator
	query     *index.Query
	start     *bytes.Buffer
	rowPrefix *bytes.Buffer
	seeked    bool

	currEntity *index.Entry
}

func (c *cursorBatch) Iterator() index.ReadBatchIterator {
	return c
}

func (c *cursorBatch) Next() bool {
	for k, v := c.nextItem(); k != nil; k, v = c.nextItem() {
		if !bytes.HasPrefix(k, c.rowPrefix.Bytes()) {
			break
		}

		if len(c.query.RangeValuePrefix) > 0 && !bytes.HasPrefix(k, c.start.Bytes()) {
			break
		}

		if len(c.query.ValueEqual) > 0 && !bytes.Equal(v, c.query.ValueEqual) {
			continue
		}

		rangeValue := make([]byte, len(k)-c.rowPrefix.Len())
		copy(rangeValue, k[c.rowPrefix.Len():])

		value := make([]byte, len(v))
		copy(value, v)

		c.currEntity = &index.Entry{
			TableName:  c.query.TableName,
			HashValue:  c.query.HashValue,
			RangeValue: rangeValue,
			Value:      value,
		}
		return true
	}
	return false
}

func (c *cursorBatch) Entry() *index.Entry {
	return c.currEntity
}

func (c *cursorBatch) reset(cur *pebble.Iterator, q *index.Query) {
	c.currEntity = nil
	c.seeked = false
	c.iter = cur
	c.query = q
	c.rowPrefix.Reset()
	c.start.Reset()
}

var (
	separator byte = 0
)

func (c *cursorBatch) nextItem() ([]byte, []byte) {
	if !c.seeked {
		if len(c.query.RangeValuePrefix) > 0 {
			c.start.WriteString(c.query.HashValue)
			c.start.WriteByte(separator)
			c.start.Write(c.query.RangeValuePrefix)
		} else {
			c.start.WriteString(c.query.HashValue)
			c.start.WriteByte(separator)
		}

		c.rowPrefix.WriteString(c.query.HashValue)
		c.rowPrefix.WriteByte(separator)

		c.seeked = true
		if c.iter.SeekGE(c.start.Bytes()) {
			return c.iter.Key(), c.iter.Value()
		}
		return nil, nil
	}
	if c.iter.Next() && c.iter.Valid() {
		return c.iter.Key(), c.iter.Value()
	}
	return nil, nil
}
