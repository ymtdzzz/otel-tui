package telemetry

import (
	"strings"
	"sync"
	"time"

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

// Flush clears the store including the cache
func (s *Store) Flush() {
	s.mut.Lock()
	defer func() {
		s.updatedAt = time.Now()
		s.mut.Unlock()
	}()

	s.svcspans = SvcSpans{}
	s.svcspansFiltered = SvcSpans{}
	s.cache.flush()
	s.updatedAt = time.Now()
}
