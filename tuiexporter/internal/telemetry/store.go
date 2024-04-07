package telemetry

import (
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	MAX_SERVICE_SPAN_COUNT = 1000
	MAX_LOG_COUNT          = 1000
)

// SpanData is a struct to represent a span
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

// LogData is a struct to represent a log
type LogData struct {
	Log         *plog.LogRecord
	ResourceLog *plog.ResourceLogs
	ScopeLog    *plog.ScopeLogs
	ReceivedAt  time.Time
}

// Store is a store of trace spans
type Store struct {
	mut                 sync.Mutex
	filterSvc           string
	filterLog           string
	svcspans            SvcSpans
	svcspansFiltered    SvcSpans
	tracecache          *TraceCache
	logs                []*LogData
	logsFiltered        []*LogData
	logcache            *LogCache
	updatedAt           time.Time
	maxServiceSpanCount int
	maxLogCount         int
}

// NewStore creates a new store
func NewStore() *Store {
	return &Store{
		mut:                 sync.Mutex{},
		svcspans:            SvcSpans{},
		svcspansFiltered:    SvcSpans{},
		tracecache:          NewTraceCache(),
		logs:                []*LogData{},
		logsFiltered:        []*LogData{},
		logcache:            NewLogCache(),
		maxServiceSpanCount: MAX_SERVICE_SPAN_COUNT, // TODO: make this configurable
		maxLogCount:         MAX_LOG_COUNT,          // TODO: make this configurable
	}
}

// GetCache returns the cache
func (s *Store) GetCache() *TraceCache {
	return s.tracecache
}

// GetSvcSpans returns the service spans in the store
func (s *Store) GetSvcSpans() *SvcSpans {
	return &s.svcspans
}

// GetFilteredSvcSpans returns the filtered service spans in the store
func (s *Store) GetFilteredSvcSpans() *SvcSpans {
	return &s.svcspansFiltered
}

// GetFilteredLogs returns the filtered logs in the store
func (s *Store) GetFilteredLogs() *[]*LogData {
	return &s.logsFiltered
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

// ApplyFilterLogs applies a filter to the logs
func (s *Store) ApplyFilterLogs(filter string) {
	s.filterLog = filter
	s.logsFiltered = []*LogData{}

	if filter == "" {
		s.logsFiltered = s.logs
		return
	}

	for _, log := range s.logs {
		sname, _ := log.ResourceLog.Resource().Attributes().Get("service.name")
		target := sname.AsString() + " " + log.Log.Body().AsString()
		if strings.Contains(target, filter) {
			s.logsFiltered = append(s.logsFiltered, log)
		}
	}
}

func (s *Store) updateFilterLogs() {
	s.ApplyFilterLogs(s.filterLog)
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
	spans, _ := s.tracecache.GetSpansByTraceIDAndSvc(traceID, sname.AsString())

	return spans
}

// GetFilteredLogByIdx returns the log at the given index
func (s *Store) GetFilteredLogByIdx(idx int) *LogData {
	if idx < 0 || idx >= len(s.logsFiltered) {
		return nil
	}
	return s.logsFiltered[idx]
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
				newtracesvc := s.tracecache.UpdateCache(sname.AsString(), sd)
				if newtracesvc {
					s.svcspans = append(s.svcspans, sd)
				}
			}
		}
	}

	// data rotation
	if len(s.svcspans) > s.maxServiceSpanCount {
		deleteSpans := s.svcspans[:len(s.svcspans)-s.maxServiceSpanCount]

		s.tracecache.DeleteCache(deleteSpans)

		s.svcspans = s.svcspans[len(s.svcspans)-s.maxServiceSpanCount:]
	}

	s.updateFilterService()
}

// AddLog adds a log to the store
func (s *Store) AddLog(logs *plog.Logs) {
	s.mut.Lock()
	defer func() {
		s.updatedAt = time.Now()
		s.mut.Unlock()
	}()

	for rli := 0; rli < logs.ResourceLogs().Len(); rli++ {
		rl := logs.ResourceLogs().At(rli)

		for sli := 0; sli < rl.ScopeLogs().Len(); sli++ {
			sl := rl.ScopeLogs().At(sli)

			for li := 0; li < sl.LogRecords().Len(); li++ {
				lr := sl.LogRecords().At(li)
				ld := &LogData{
					Log:         &lr,
					ResourceLog: &rl,
					ScopeLog:    &sl,
					ReceivedAt:  time.Now(),
				}
				s.logs = append(s.logs, ld)
				s.logcache.UpdateCache(ld)
			}
		}
	}

	// data rotation
	if len(s.logs) > s.maxLogCount {
		deleteLogs := s.logs[:len(s.logs)-s.maxLogCount]
		s.logs = s.logs[len(s.logs)-s.maxLogCount:]

		s.logcache.DeleteCache(deleteLogs)
	}

	s.updateFilterLogs()
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
	s.tracecache.flush()
	s.logs = []*LogData{}
	s.logsFiltered = []*LogData{}
	s.logcache.flush()
	s.updatedAt = time.Now()
}
