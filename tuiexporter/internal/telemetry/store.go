package telemetry

import (
	"strings"
	"sync"
	"time"

	"github.com/rivo/tview"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const MAX_SERVICE_SPAN_COUNT = 1000

type SpanData struct {
	Span         *ptrace.Span
	ResourceSpan *ptrace.ResourceSpans
	ScopeSpans   *ptrace.ScopeSpans
	ReceivedAt   time.Time
}

// IsRoot returns true if the span is a root span
func (sd *SpanData) IsRoot() bool {
	return sd.Span.ParentSpanID().IsEmpty()
}

// SvcSpans is a slice of service spans
// This is a slice of one span of a single service
type SvcSpans []*SpanData

// Store is a store of trace spans
type Store struct {
	mut                 sync.Mutex
	filterSvc           string
	svcspans            SvcSpans
	svcspansFiltered    SvcSpans
	cache               *TraceCache
	updatedAt           time.Time
	maxServiceSpanCount int
}

// NewStore creates a new store
func NewStore() *Store {
	return &Store{
		mut:                 sync.Mutex{},
		svcspans:            []*SpanData{},
		svcspansFiltered:    []*SpanData{},
		cache:               NewTraceCache(),
		maxServiceSpanCount: MAX_SERVICE_SPAN_COUNT, // TODO: make this configurable
	}
}

// GetCache returns the cache
func (s *Store) GetCache() *TraceCache {
	return s.cache
}

// GetSvcSpans returns the service spans in the store
func (s *Store) GetSvcSpans() *SvcSpans {
	return &s.svcspans
}

// GetFilteredSvcSpans returns the filtered service spans in the store
func (s *Store) GetFilteredSvcSpans() *SvcSpans {
	return &s.svcspansFiltered
}

// UpdatedAt returns the last updated time
func (s *Store) UpdatedAt() time.Time {
	return s.updatedAt
}

// ApplyFilterService applies a filter to the traces
func (s *Store) ApplyFilterService(svc string) {
	s.filterSvc = svc
	s.svcspansFiltered = []*SpanData{}

	if svc == "" {
		s.svcspansFiltered = s.svcspans
		return
	}

	for _, span := range s.svcspans {
		sname, _ := span.ResourceSpan.Resource().Attributes().Get("service.name")
		if strings.Contains(sname.AsString(), svc) {
			s.svcspansFiltered = append(s.svcspansFiltered, span)
		}
	}
}

func (s *Store) updateFilterService() {
	s.ApplyFilterService(s.filterSvc)
}

// GetTraceIDByFilteredIdx returns the trace at the given index
func (s *Store) GetTraceIDByFilteredIdx(idx int) string {
	if idx >= 0 && idx < len(s.svcspansFiltered) {
		return s.svcspansFiltered[idx].Span.TraceID().String()
	}
	return ""
}

// GetFilteredServiceSpansByIdx returns the spans for a given service at the given index
func (s *Store) GetFilteredServiceSpansByIdx(idx int) []*SpanData {
	if idx < 0 || idx >= len(s.svcspansFiltered) {
		return nil
	}
	span := s.svcspansFiltered[idx]
	traceID := span.Span.TraceID().String()
	sname, _ := span.ResourceSpan.Resource().Attributes().Get("service.name")
	spans, _ := s.cache.GetSpansByTraceIDAndSvc(traceID, sname.AsString())

	return spans
}

// AddSpan adds a span to the store
func (s *Store) AddSpan(traces *ptrace.Traces) {
	s.mut.Lock()
	defer func() {
		s.updatedAt = time.Now()
		s.mut.Unlock()
	}()

	for rsi := 0; rsi < traces.ResourceSpans().Len(); rsi++ {
		rs := traces.ResourceSpans().At(rsi)

		for ssi := 0; ssi < rs.ScopeSpans().Len(); ssi++ {
			ss := rs.ScopeSpans().At(ssi)

			for si := 0; si < ss.Spans().Len(); si++ {
				span := ss.Spans().At(si)
				// attribute service.name is required
				// see: https://opentelemetry.io/docs/specs/semconv/resource/#service
				sname, _ := rs.Resource().Attributes().Get("service.name")
				sd := &SpanData{
					Span:         &span,
					ResourceSpan: &rs,
					ScopeSpans:   &ss,
					ReceivedAt:   time.Now(),
				}
				newtracesvc := s.cache.UpdateCache(sname.AsString(), sd)
				if newtracesvc {
					s.svcspans = append(s.svcspans, sd)
				}
			}
		}
	}

	// data rotation
	if len(s.svcspans) > s.maxServiceSpanCount {
		deleteSpans := s.svcspans[:len(s.svcspans)-s.maxServiceSpanCount]

		s.cache.DeleteCache(deleteSpans)

		s.svcspans = s.svcspans[len(s.svcspans)-s.maxServiceSpanCount:]
	}

	s.updateFilterService()
}

// implementations for tview Virtual Table
// see: https://github.com/rivo/tview/wiki/VirtualTable
// TODO: these should be implemented in component package?

func (s *SpanData) GetCell(column int) *tview.TableCell {
	text := "N/A"

	switch column {
	case 0:
		text = s.Span.TraceID().String()
	case 1:
		if serviceName, ok := s.ResourceSpan.Resource().Attributes().Get("service.name"); ok {
			text = serviceName.AsString()
		}
	case 2:
		text = s.ReceivedAt.Local().Format("2006-01-02 15:04:05")
	}

	return tview.NewTableCell(text)
}

func (t SvcSpans) GetCell(row, column int) *tview.TableCell {
	if row >= 0 && row < len(t) {
		return t[row].GetCell(column)
	}
	return nil
}

func (t SvcSpans) GetRowCount() int {
	return len(t)
}

func (t SvcSpans) GetColumnCount() int {
	// 0: TraceID
	// 1: ServiceName
	// 2: ReceivedAt
	return 3
}

// readonly table
func (t SvcSpans) SetCell(row, column int, cell *tview.TableCell) {}
func (t SvcSpans) RemoveRow(row int)                              {}
func (t SvcSpans) RemoveColumn(column int)                        {}
func (t SvcSpans) InsertRow(row int)                              {}
func (t SvcSpans) InsertColumn(column int)                        {}
func (t SvcSpans) Clear()                                         {}
