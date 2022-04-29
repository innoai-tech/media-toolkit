package label

import (
	"bytes"
	"context"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label/index"
	"github.com/prometheus/common/model"
	"strconv"
)

type IndexWriter interface {
	NewWriteBatch() index.WriteBatch
	BatchWrite(context.Context, index.WriteBatch) error
}

type Writer interface {
	Put(ctx context.Context, metricName string, blobs []blob.Info) error
	PutOne(ctx context.Context, timeRange TimeRange, metricName string, b blob.Info) error

	PutLabels(ctx context.Context, timeRange TimeRange, metricName string, ref blob.Ref, labels blob.Labels) error
	DelLabels(ctx context.Context, timeRange TimeRange, metricName string, ref blob.Ref, labels blob.Labels) error

	DelOne(ctx context.Context, timeRange TimeRange, metricName string, ref blob.Ref) error
}

func NewWriter(schemaCfg config.SchemaConfig, indexWriter IndexWriter, schema index.BlobStoreSchema) Writer {
	return &writer{
		schemaCfg:   schemaCfg,
		indexWriter: indexWriter,
		schema:      schema,
	}
}

type writer struct {
	schemaCfg   config.SchemaConfig
	schema      index.BlobStoreSchema
	indexWriter IndexWriter
}

func (c *writer) Put(ctx context.Context, metricName string, blobs []blob.Info) error {
	for _, b := range blobs {
		if err := c.PutOne(ctx, b.TimeRange, metricName, b); err != nil {
			return err
		}
	}
	return nil
}

func (c *writer) PutOne(ctx context.Context, timeRange TimeRange, metricName string, b blob.Info) error {
	writeReqs, err := c.calculateIndexEntries(ctx, timeRange, metricName, b.Ref, b.Labels)
	if err != nil {
		return err
	}
	if err := c.indexWriter.BatchWrite(ctx, writeReqs); err != nil {
		return err
	}
	return nil
}

func (c *writer) DelOne(ctx context.Context, timeRange TimeRange, metricName string, ref blob.Ref) error {
	return c.PutLabels(ctx, timeRange, metricName, ref, blob.Labels{
		blob.LabelDeleted: {strconv.FormatInt(model.Now().Unix(), 10)},
	})
}

func (c *writer) PutLabels(ctx context.Context, timeRange TimeRange, metricName string, ref blob.Ref, labels blob.Labels) error {
	writeReqs, err := c.calculateIndexEntries(ctx, timeRange, metricName, ref, labels)
	if err != nil {
		return err
	}
	if err := c.indexWriter.BatchWrite(ctx, writeReqs); err != nil {
		return err
	}
	return nil
}

func (c *writer) DelLabels(ctx context.Context, timeRange TimeRange, metricName string, ref blob.Ref, labels blob.Labels) error {
	writeReqs, err := c.calculateIndexEntriesForLabelDelete(ctx, timeRange, metricName, ref, labels)
	if err != nil {
		return err
	}
	if err := c.indexWriter.BatchWrite(ctx, writeReqs); err != nil {
		return err
	}
	return nil
}

func (c *writer) calculateIndexEntries(ctx context.Context, timeRange TimeRange, metricName string, ref blob.Ref, labels blob.Labels) (index.WriteBatch, error) {
	entries := make([]index.Entry, 0)

	keys, labelEntries, err := c.schema.GetCacheKeysAndLabelWriteEntries(timeRange, ref.UserID, metricName, c.schemaCfg.ExternalKey(ref), labels)
	if err != nil {
		return nil, err
	}

	for i, _ := range keys {
		entries = append(entries, labelEntries[i]...)
	}

	result := c.indexWriter.NewWriteBatch()

	for i := range entries {
		result.Add(entries[i])
	}

	return result, nil
}

func (c *writer) calculateIndexEntriesForLabelDelete(ctx context.Context, timeRange TimeRange, metricName string, ref blob.Ref, labels blob.Labels) (index.WriteBatch, error) {
	entries := make([]index.Entry, 0)

	keys, labelEntries, err := c.schema.GetCacheKeysAndLabelWriteEntries(timeRange, ref.UserID, metricName, c.schemaCfg.ExternalKey(ref), labels)
	if err != nil {
		return nil, err
	}

	for i, _ := range keys {
		entries = append(entries, labelEntries[i]...)
	}

	result := c.indexWriter.NewWriteBatch()

	for i := range entries {
		e := entries[i]

		// skip range value metric
		if bytes.Equal(e.Value, []byte{0}) {
			continue
		}

		result.Delete(e)
	}

	return result, nil
}
