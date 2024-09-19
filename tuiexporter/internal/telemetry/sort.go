package telemetry

import "sort"

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
		sort.Slice(svcSpans, func(i, j int) bool {
			// default sort is received_at asc
			return svcSpans[i].ReceivedAt.Before(svcSpans[j].ReceivedAt)
		})
	case SORT_TYPE_LATENCY_DESC:
		sort.Slice(svcSpans, func(i, j int) bool {
			istart := svcSpans[i].Span.StartTimestamp().AsTime()
			iend := svcSpans[i].Span.EndTimestamp().AsTime()
			iduration := iend.Sub(istart)
			jstart := svcSpans[j].Span.StartTimestamp().AsTime()
			jend := svcSpans[j].Span.EndTimestamp().AsTime()
			jduration := jend.Sub(jstart)
			return iduration > jduration
		})
	case SORT_TYPE_LATENCY_ASC:
		sort.Slice(svcSpans, func(i, j int) bool {
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
