package telemetry

import (
	"slices"
	"sort"
	"time"
)

const (
	SORT_TYPE_NONE         SortType = "none"
	SORT_TYPE_LATENCY_DESC SortType = "latency-desc"
	SORT_TYPE_LATENCY_ASC  SortType = "latency-asc"
)

// SortType is sort type
type SortType string

func (t SortType) IsNone() bool {
	return t == SORT_TYPE_NONE
}

func (t SortType) IsDesc() bool {
	return t == SORT_TYPE_LATENCY_DESC
}

func (t SortType) GetHeaderLabel() string {
	switch t {
	case SORT_TYPE_LATENCY_DESC:
		return "Latency"
	case SORT_TYPE_LATENCY_ASC:
		return "Latency"
	}
	return "N/A"
}

func sortSvcSpans(svcSpans SvcSpans, sortType SortType) {
	switch sortType {
	case SORT_TYPE_NONE:
		sortByTime(svcSpans, (*SpanData).sortTime)
	case SORT_TYPE_LATENCY_DESC:
		sort.SliceStable(svcSpans, func(i, j int) bool {
			istart := svcSpans[i].Span.StartTimestamp().AsTime()
			iend := svcSpans[i].Span.EndTimestamp().AsTime()
			iduration := iend.Sub(istart)
			jstart := svcSpans[j].Span.StartTimestamp().AsTime()
			jend := svcSpans[j].Span.EndTimestamp().AsTime()
			jduration := jend.Sub(jstart)
			return iduration > jduration
		})
	case SORT_TYPE_LATENCY_ASC:
		sort.SliceStable(svcSpans, func(i, j int) bool {
			istart := svcSpans[i].Span.StartTimestamp().AsTime()
			iend := svcSpans[i].Span.EndTimestamp().AsTime()
			iduration := iend.Sub(istart)
			jstart := svcSpans[j].Span.StartTimestamp().AsTime()
			jend := svcSpans[j].Span.EndTimestamp().AsTime()
			jduration := jend.Sub(jstart)
			return iduration < jduration
		})
	}
}

// sortByTime sorts items oldest first by timeOf.
// Items with equal times keep their order, which is the received order.
func sortByTime[T any](items []T, timeOf func(T) time.Time) {
	type keyed struct {
		item T
		at   time.Time
	}
	keys := make([]keyed, len(items))
	for i, item := range items {
		keys[i] = keyed{item: item, at: timeOf(item)}
	}
	slices.SortStableFunc(keys, func(a, b keyed) int {
		return a.at.Compare(b.at)
	})
	for i, k := range keys {
		items[i] = k.item
	}
}
