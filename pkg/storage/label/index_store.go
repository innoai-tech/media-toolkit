package label

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label/index"
	"github.com/prometheus/prometheus/model/labels"
	"sort"
	"sync"
)

type IndexStore interface {
	GetBlobRefs(ctx context.Context, timeRange blob.TimeRange, userID string, metricName string, matchers ...*labels.Matcher) ([]blob.Ref, error)
	GetBlobs(ctx context.Context, timeRange blob.TimeRange, userID string, metricName string, matchers ...*labels.Matcher) ([]blob.Info, error)
	RefsToBlobs(ctx context.Context, refs []blob.Ref, metricName string) ([]blob.Info, error)
}

func NewIndexStore(schemaCfg config.SchemaConfig, index index.Client, schema index.BlobStoreSchema) IndexStore {
	return &indexStore{
		schema:    schema,
		index:     index,
		schemaCfg: schemaCfg,
	}
}

type indexStore struct {
	schema    index.BlobStoreSchema
	index     index.Client
	schemaCfg config.SchemaConfig
}

func (c *indexStore) GetBlobs(ctx context.Context, timeRange blob.TimeRange, userID string, metricName string, matchers ...*labels.Matcher) ([]blob.Info, error) {
	refs, err := c.GetBlobRefs(ctx, timeRange, userID, metricName, matchers...)
	if err != nil {
		return nil, err
	}
	return c.RefsToBlobs(ctx, refs, metricName)
}

func (c *indexStore) RefsToBlobs(ctx context.Context, refs []blob.Ref, metricName string) ([]blob.Info, error) {
	queries := make([]index.Query, 0)
	for _, ref := range refs {
		q, err := c.schema.GetMetricLabelValues(ref.TimeRange, ref.UserID, metricName, c.schemaCfg.ExternalKey(ref))
		if err != nil {
			return nil, err
		}
		queries = append(queries, q...)
	}

	entries, err := c.lookupEntriesByQueries(ctx, queries)
	if err != nil {
		return nil, err
	}

	blobSet := make(map[string]*blob.Info)
	deletedBlobs := make(map[string]struct{})

	for i := range entries {
		e := entries[i]

		rk, err := index.DecodeRangeValue(e.RangeValue)
		if err != nil {
			return nil, err
		}
		labelName := rk.(index.RangeValueLabelValue).LabelName()
		blobID := e.HashValue
		if labelName == blob.LabelDeleted {
			deletedBlobs[blobID] = struct{}{}
			continue
		}

		b, ok := blobSet[blobID]
		if !ok {
			bb, err := blob.ParseExternalKey(blobID, "")
			if err != nil {
				return nil, err
			}
			bb.Labels = map[string][]string{}
			blobSet[blobID] = &bb
			b = &bb
		}
		b.Labels[labelName] = append(b.Labels[labelName], string(e.Value))
	}

	blobs := make([]blob.Info, 0, len(blobSet))

	for id := range blobSet {
		if _, ok := deletedBlobs[id]; !ok {
			blobs = append(blobs, *blobSet[id])
		}
	}

	// TODO sort

	return blobs, nil
}

func (c *indexStore) GetBlobRefs(ctx context.Context, timeRange blob.TimeRange, userID string, metricName string, matchers ...*labels.Matcher) ([]blob.Ref, error) {
	l := logr.FromContextOrDiscard(ctx)
	blobIDs, err := c.lookupBlobMatchers(ctx, timeRange, userID, metricName, matchers)
	if err != nil {
		return nil, err
	}
	l.V(1).Info("blob-ids", len(blobIDs))
	return c.convertBlobIDsToBlobRefs(ctx, userID, blobIDs)
}

func (c *indexStore) lookupBlobMatchers(ctx context.Context, timeRange blob.TimeRange, userID string, metricName string, matchers []*labels.Matcher) ([]string, error) {
	incomingIDs := make(chan []string)
	incomingErrors := make(chan error)

	if len(matchers) == 0 {
		matchers = []*labels.Matcher{nil}
	}

	for _, matcher := range matchers {
		go func(matcher *labels.Matcher) {
			ids, err := c.lookupIdsByMatcher(ctx, timeRange, userID, metricName, matcher, nil)
			if err != nil {
				incomingErrors <- err
				return
			}
			incomingIDs <- ids
		}(matcher)
	}

	var ids []string
	var initialized bool

	var preIntersectionCount int

	for i := 0; i < len(matchers); i++ {
		select {
		case incoming := <-incomingIDs:
			preIntersectionCount += len(incoming)
			if !initialized {
				ids = incoming
				initialized = true
			} else {
				ids = intersectStrings(ids, incoming)
			}
		case err := <-incomingErrors:
			// The idea is that if we have 2 matchers, and if one returns a lot of
			// series and the other returns only 10 (a few), we don't lookup the first one at all.
			// We just manually filter through the 10 series again using "filterChunksByMatchers",
			// saving us from looking up and intersecting a lot of series.
			// TODO
			return nil, err
		}
	}

	return ids, nil
}

func (c *indexStore) lookupIdsByMatcher(ctx context.Context, timeRange TimeRange, userID string, metricName string, matcher *labels.Matcher, filter func([]index.Query) []index.Query) ([]string, error) {
	var err error
	var queries []index.Query

	if matcher == nil {
		queries, err = c.schema.GetReadQueriesForMetric(timeRange, userID, metricName)
	} else if matcher.Type == labels.MatchEqual {
		queries, err = c.schema.GetReadQueriesForMetricLabelValue(timeRange, userID, metricName, matcher.Name, matcher.Value)
	} else {
		queries, err = c.schema.GetReadQueriesForMetricLabel(timeRange, userID, metricName, matcher.Name)
	}
	if err != nil {
		return nil, err
	}

	if filter != nil {
		queries = filter(queries)
	}

	entries, err := c.lookupEntriesByQueries(ctx, queries)
	if err != nil {
		return nil, err
	}

	ids, err := parseIndexEntries(ctx, entries, matcher)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func parseIndexEntries(_ context.Context, entries []index.Entry, matcher *labels.Matcher) ([]string, error) {
	// Nothing to do if there are no entries.
	if len(entries) == 0 {
		return nil, nil
	}

	matchSet := map[string]struct{}{}
	if matcher != nil && matcher.Type == labels.MatchRegexp {
		set := FindSetMatches(matcher.Value)
		for _, v := range set {
			matchSet[v] = struct{}{}
		}
	}

	ids := make([]string, 0, len(entries))
	for i := range entries {
		entry := entries[i]
		rk, err := index.DecodeRangeValue(entry.RangeValue)
		if err != nil {
			return nil, err
		}

		blobID := rk.(interface{ BlobID() string }).BlobID()

		// If the matcher is like a set (=~"a|b|c|d|...") and
		// the label value is not in that set move on.
		if len(matchSet) > 0 {
			if _, ok := matchSet[string(entry.Value)]; !ok {
				continue
			}

			// If its in the set, then add it to set, we don't need to run
			// matcher on it again.
			ids = append(ids, blobID)
			continue
		}

		if matcher != nil && !matcher.Matches(string(entry.Value)) {
			continue
		}

		ids = append(ids, blobID)
	}
	// Return ids sorted and deduped because they will be merged with other sets.
	sort.Strings(ids)
	ids = uniqueStrings(ids)
	return ids, nil
}

func (c *indexStore) lookupEntriesByQueries(ctx context.Context, queries []index.Query) ([]index.Entry, error) {
	// Nothing to do if there are no queries.
	if len(queries) == 0 {
		return nil, nil
	}

	wg := sync.WaitGroup{}
	chEntry := make(chan *index.Entry)
	entries := make([]index.Entry, 0)

	go func() {
		wg.Add(1)
		defer wg.Done()

		for entry := range chEntry {
			entries = append(entries, *entry)
		}
	}()

	err := c.index.QueryPages(ctx, queries, func(resp index.ReadBatchResult, query index.Query) error {
		iter := resp.Iterator()
		for iter.Next() {
			chEntry <- iter.Entry()
		}
		return nil
	})
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "error querying storage", "err", err)
	}

	close(chEntry)
	wg.Wait()

	return entries, err
}

func (c *indexStore) convertBlobIDsToBlobRefs(ctx context.Context, userID string, blobIds []string) ([]blob.Ref, error) {
	refs := make([]blob.Ref, 0, len(blobIds))
	for _, blobID := range blobIds {
		b, err := blob.ParseExternalKey(blobID, userID)
		if err != nil {
			return nil, err
		}
		refs = append(refs, b.Ref)
	}
	return refs, nil
}
