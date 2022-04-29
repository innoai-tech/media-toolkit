package local

import "github.com/innoai-tech/media-toolkit/pkg/storage/label/index"

type TableWrites struct {
	puts    []index.Entry
	deletes []index.Entry
}

type WriteBatch struct {
	Writes map[string]*TableWrites
}

func (b *WriteBatch) getOrCreateTableWrites(tableName string) *TableWrites {
	writes, ok := b.Writes[tableName]
	if !ok {
		writes = &TableWrites{}
		b.Writes[tableName] = writes
	}
	return writes
}

func (b *WriteBatch) Delete(entry index.Entry) {
	t := b.getOrCreateTableWrites(entry.TableName)
	t.deletes = append(t.deletes, entry)
}

func (b *WriteBatch) Add(entry index.Entry) {
	t := b.getOrCreateTableWrites(entry.TableName)
	t.puts = append(t.puts, entry)
}
