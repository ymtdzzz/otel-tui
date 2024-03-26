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

// Traces is a slice of root spans
type Traces []*SpanData

// Store is a store of trace spans
type Store struct {
	mut            sync.Mutex
	filterSvc      string
	traces         Traces
	tracesFiltered Traces
	cache          *TraceCache
	updatedAt      time.Time
}

// NewStore creates a new store
func NewStore() *Store {
	return &Store{
		mut:            sync.Mutex{},
		traces:         []*SpanData{},
		tracesFiltered: []*SpanData{},
		cache:          NewTraceCache(),
	}
}

// GetCache returns the cache
func (s *Store) GetCache() *TraceCache {
	return s.cache
}

// GetTraces returns the traces in the store
func (s *Store) GetTraces() *Traces {
	return &s.traces
}

// GetFilteredTraces returns the filtered traces in the store
func (s *Store) GetFilteredTraces() *Traces {
	return &s.tracesFiltered
}

// UpdatedAt returns the last updated time
func (s *Store) UpdatedAt() time.Time {
	return s.updatedAt
}

// ApplyFilterService applies a filter to the traces
func (s *Store) ApplyFilterService(svc string) {
	s.filterSvc = svc
	s.tracesFiltered = []*SpanData{}

	if svc == "" {
		s.tracesFiltered = s.traces
		return
	}

	for _, span := range s.traces {
		sname, _ := span.ResourceSpan.Resource().Attributes().Get("service.name")
		if strings.Contains(sname.AsString(), svc) {
			s.tracesFiltered = append(s.tracesFiltered, span)
		}
	}
}

func (s *Store) updateFilterService() {
	s.ApplyFilterService(s.filterSvc)
}

// GetTraceIDByFilteredIdx returns the trace at the given index
func (s *Store) GetTraceIDByFilteredIdx(idx int) string {
	if idx >= 0 && idx < len(s.tracesFiltered) {
		return s.tracesFiltered[idx].Span.TraceID().String()
	}
	return ""
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
				sname, _ := rs.Resource().Attributes().Get("service.name")
				sd := &SpanData{
					Span:         &span,
					ResourceSpan: &rs,
					ScopeSpans:   &ss,
					ReceivedAt:   time.Now(),
				}
				newtracesvc := s.cache.UpdateCache(sname.AsString(), sd)
				if newtracesvc {
					s.traces = append(s.traces, sd)
				}
			}
		}
	}

	// data rotation
	if len(s.traces) > MAX_SERVICE_SPAN_COUNT {
		deleteSpans := s.traces[:len(s.traces)-MAX_SERVICE_SPAN_COUNT]

		s.cache.DeleteCache(deleteSpans)

		s.traces = s.traces[len(s.traces)-MAX_SERVICE_SPAN_COUNT:]
	}

	s.updateFilterService()
}

// GetFilteredServiceSpansByIdx returns the spans for a given service at the given index
func (s *Store) GetFilteredServiceSpansByIdx(idx int) []*SpanData {
	if idx < 0 || idx >= len(s.tracesFiltered) {
		return nil
	}
	span := s.tracesFiltered[idx]
	traceID := span.Span.TraceID().String()
	sname, _ := span.ResourceSpan.Resource().Attributes().Get("service.name")
	spans, _ := s.cache.GetSpansByTraceIDAndSvc(traceID, sname.AsString())

	return spans
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

func (t Traces) GetCell(row, column int) *tview.TableCell {
	if row >= 0 && row < len(t) {
		return t[row].GetCell(column)
	}
	return nil
}

func (t Traces) GetRowCount() int {
	return len(t)
}

func (t Traces) GetColumnCount() int {
	// 0: TraceID
	// 1: ServiceName
	// 2: ReceivedAt
	return 3
}

// readonly table
func (t Traces) SetCell(row, column int, cell *tview.TableCell) {}
func (t Traces) RemoveRow(row int)                              {}
func (t Traces) RemoveColumn(column int)                        {}
func (t Traces) InsertRow(row int)                              {}
func (t Traces) InsertColumn(column int)                        {}
func (t Traces) Clear()                                         {}
