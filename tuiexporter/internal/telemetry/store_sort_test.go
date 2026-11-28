package telemetry

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

var sortBaseTime = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

func sortTestTime(sec int) time.Time {
	return sortBaseTime.Add(time.Duration(sec) * time.Second)
}

// rootSpans builds one root span per name, each in its own trace. Names must be
// unique and at most 8 bytes, since they also make the IDs.
// start is the span start in seconds after sortBaseTime, or -1 to leave it unset.
func rootSpans(t *testing.T, svc string, spans map[string]int, order ...string) *ptrace.Traces {
	t.Helper()
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr("service.name", svc)
	ss := rs.ScopeSpans().AppendEmpty()
	for _, name := range order {
		span := ss.Spans().AppendEmpty()
		span.SetName(name)
		var traceID pcommon.TraceID
		var spanID pcommon.SpanID
		copy(traceID[:], name)
		copy(spanID[:], name)
		span.SetTraceID(traceID)
		span.SetSpanID(spanID)
		if start := spans[name]; start >= 0 {
			span.SetStartTimestamp(pcommon.NewTimestampFromTime(sortTestTime(start)))
			span.SetEndTimestamp(pcommon.NewTimestampFromTime(sortTestTime(start).Add(time.Duration(len(name)) * time.Millisecond)))
		}
	}
	return &traces
}

func spanNames(spans []*SpanData) []string {
	names := make([]string, 0, len(spans))
	for _, s := range spans {
		names = append(names, s.Span.Name())
	}
	return names
}

func TestStoreSortsTracesByStartTime(t *testing.T) {
	clock := clockwork.NewFakeClockAt(sortTestTime(100))
	store := NewStore(clock)

	store.AddSpan(rootSpans(t, "svc", map[string]int{"b": 20, "d": 40}, "d", "b"))
	clock.Advance(time.Second)
	store.AddSpan(rootSpans(t, "svc", map[string]int{"a": 10, "c": 30}, "c", "a"))

	assert.Equal(t, []string{"a", "b", "c", "d"}, spanNames(*store.GetFilteredSvcSpans()))
	// The store keeps received order, so rotation drops the oldest received spans.
	assert.Equal(t, []string{"d", "b", "c", "a"}, spanNames(store.svcspans))
}

func TestStoreSortsTracesWithoutStartTimeByReceivedAt(t *testing.T) {
	clock := clockwork.NewFakeClockAt(sortTestTime(25))
	store := NewStore(clock)

	store.AddSpan(rootSpans(t, "svc", map[string]int{"a": 10, "c": 30}, "c", "a"))
	store.AddSpan(rootSpans(t, "svc", map[string]int{"b": -1}, "b"))

	assert.Equal(t, []string{"a", "b", "c"}, spanNames(*store.GetFilteredSvcSpans()))
}

func TestStoreSortsFilteredTracesByStartTime(t *testing.T) {
	store := NewStore(clockwork.NewFakeClockAt(sortTestTime(100)))
	store.AddSpan(rootSpans(t, "svc", map[string]int{"keep-2": 20, "drop": 5, "keep-1": 10}, "keep-2", "drop", "keep-1"))

	store.ApplyFilterTraces("keep", SORT_TYPE_NONE)

	assert.Equal(t, []string{"keep-1", "keep-2"}, spanNames(*store.GetFilteredSvcSpans()))
}

func TestStoreLatencySortKeepsReceivedOrderForRotation(t *testing.T) {
	store := NewStore(clockwork.NewFakeClockAt(sortTestTime(100)))
	store.maxServiceSpanCount = 2
	store.ApplyFilterTraces("", SORT_TYPE_LATENCY_DESC)

	// Span latency is len(name) milliseconds.
	store.AddSpan(rootSpans(t, "svc", map[string]int{"b": 1, "aaa": 2}, "b", "aaa"))
	assert.Equal(t, []string{"aaa", "b"}, spanNames(*store.GetFilteredSvcSpans()))

	store.AddSpan(rootSpans(t, "svc", map[string]int{"cc": 3}, "cc"))

	// "b" was received first, so rotation drops it, not the slowest span.
	assert.Equal(t, []string{"aaa", "cc"}, spanNames(*store.GetFilteredSvcSpans()))
}

// logRecords builds one log record per body. ts and observed are seconds after
// sortBaseTime, or -1 to leave the field unset.
type logTimes struct{ ts, observed int }

func logRecords(times map[string]logTimes, order ...string) *plog.Logs {
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr("service.name", "svc")
	sl := rl.ScopeLogs().AppendEmpty()
	for _, body := range order {
		lr := sl.LogRecords().AppendEmpty()
		lr.Body().SetStr(body)
		if ts := times[body].ts; ts >= 0 {
			lr.SetTimestamp(pcommon.NewTimestampFromTime(sortTestTime(ts)))
		}
		if observed := times[body].observed; observed >= 0 {
			lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(sortTestTime(observed)))
		}
	}
	return &logs
}

func logBodies(logs []*LogData) []string {
	bodies := make([]string, 0, len(logs))
	for _, l := range logs {
		bodies = append(bodies, l.Log.Body().AsString())
	}
	return bodies
}

func TestStoreSortsLogsByTimestamp(t *testing.T) {
	clock := clockwork.NewFakeClockAt(sortTestTime(25))
	store := NewStore(clock)

	store.AddLog(logRecords(map[string]logTimes{
		"log-4": {ts: 40, observed: 1},
		"log-2": {ts: -1, observed: 20}, // no Timestamp, falls back to ObservedTimestamp
	}, "log-4", "log-2"))
	store.AddLog(logRecords(map[string]logTimes{
		"log-1": {ts: 10, observed: 50},
		"log-3": {ts: -1, observed: -1}, // no timestamps, falls back to received at (25)
	}, "log-3", "log-1"))

	assert.Equal(t, []string{"log-1", "log-2", "log-3", "log-4"}, logBodies(*store.GetFilteredLogs()))
	assert.Equal(t, []string{"log-4", "log-2", "log-3", "log-1"}, logBodies(store.logs))

	store.ApplyFilterLogs("log-")
	assert.Equal(t, []string{"log-1", "log-2", "log-3", "log-4"}, logBodies(*store.GetFilteredLogs()))
}

func TestStoreSortsLogsWithSameTimestampByReceivedOrder(t *testing.T) {
	store := NewStore(clockwork.NewFakeClockAt(sortTestTime(100)))

	store.AddLog(logRecords(map[string]logTimes{"x": {ts: 10, observed: -1}, "y": {ts: 10, observed: -1}}, "y", "x"))
	store.AddLog(logRecords(map[string]logTimes{"z": {ts: 10, observed: -1}}, "z"))

	assert.Equal(t, []string{"y", "x", "z"}, logBodies(*store.GetFilteredLogs()))
}

func TestStoreLogRotationDropsOldestReceived(t *testing.T) {
	store := NewStore(clockwork.NewFakeClockAt(sortTestTime(100)))
	store.maxLogCount = 2

	store.AddLog(logRecords(map[string]logTimes{"late": {ts: 30, observed: -1}}, "late"))
	store.AddLog(logRecords(map[string]logTimes{"early": {ts: 10, observed: -1}}, "early"))
	store.AddLog(logRecords(map[string]logTimes{"mid": {ts: 20, observed: -1}}, "mid"))

	assert.Equal(t, []string{"early", "mid"}, logBodies(*store.GetFilteredLogs()))
}

// gauges builds one gauge per name. Each value lists the data point timestamps
// in seconds after sortBaseTime.
func gauges(points map[string][]int, order ...string) *pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("service.name", "svc")
	sm := rm.ScopeMetrics().AppendEmpty()
	for _, name := range order {
		m := sm.Metrics().AppendEmpty()
		m.SetName(name)
		g := m.SetEmptyGauge()
		for _, sec := range points[name] {
			dp := g.DataPoints().AppendEmpty()
			dp.SetTimestamp(pcommon.NewTimestampFromTime(sortTestTime(sec)))
			dp.SetIntValue(int64(sec))
		}
	}
	return &metrics
}

func metricNames(metrics []*MetricData) []string {
	names := make([]string, 0, len(metrics))
	for _, m := range metrics {
		names = append(names, m.Metric.Name())
	}
	return names
}

func TestStoreSortsMetricsByNewestDataPoint(t *testing.T) {
	clock := clockwork.NewFakeClockAt(sortTestTime(25))
	store := NewStore(clock)

	store.AddMetric(gauges(map[string][]int{
		"m-4": {5, 40},
		"m-2": {20, 1},
	}, "m-4", "m-2"))
	store.AddMetric(gauges(map[string][]int{
		"m-1": {10},
		"m-3": nil, // no data points, falls back to received at (25)
	}, "m-3", "m-1"))

	assert.Equal(t, []string{"m-1", "m-2", "m-3", "m-4"}, metricNames(*store.GetFilteredMetrics()))
	assert.Equal(t, []string{"m-4", "m-2", "m-3", "m-1"}, metricNames(store.metrics))

	store.ApplyFilterMetrics("m-")
	assert.Equal(t, []string{"m-1", "m-2", "m-3", "m-4"}, metricNames(*store.GetFilteredMetrics()))
}

func TestStoreSortsEveryMetricTypeByNewestDataPoint(t *testing.T) {
	// Each metric has an old data point at 1s and its newest one at the given second.
	appendPoints := func(appendTs func(pcommon.Timestamp), newest int) {
		appendTs(pcommon.NewTimestampFromTime(sortTestTime(newest)))
		appendTs(pcommon.NewTimestampFromTime(sortTestTime(1)))
	}
	fill := map[string]func(m pmetric.Metric, newest int){
		"gauge": func(m pmetric.Metric, newest int) {
			dps := m.SetEmptyGauge().DataPoints()
			appendPoints(func(ts pcommon.Timestamp) { dps.AppendEmpty().SetTimestamp(ts) }, newest)
		},
		"sum": func(m pmetric.Metric, newest int) {
			dps := m.SetEmptySum().DataPoints()
			appendPoints(func(ts pcommon.Timestamp) { dps.AppendEmpty().SetTimestamp(ts) }, newest)
		},
		"histogram": func(m pmetric.Metric, newest int) {
			dps := m.SetEmptyHistogram().DataPoints()
			appendPoints(func(ts pcommon.Timestamp) { dps.AppendEmpty().SetTimestamp(ts) }, newest)
		},
		"exponential histogram": func(m pmetric.Metric, newest int) {
			dps := m.SetEmptyExponentialHistogram().DataPoints()
			appendPoints(func(ts pcommon.Timestamp) { dps.AppendEmpty().SetTimestamp(ts) }, newest)
		},
		"summary": func(m pmetric.Metric, newest int) {
			dps := m.SetEmptySummary().DataPoints()
			appendPoints(func(ts pcommon.Timestamp) { dps.AppendEmpty().SetTimestamp(ts) }, newest)
		},
	}
	received := []string{"summary", "exponential histogram", "histogram", "sum", "gauge"}
	newest := map[string]int{"gauge": 10, "sum": 20, "histogram": 30, "exponential histogram": 40, "summary": 50}

	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("service.name", "svc")
	sm := rm.ScopeMetrics().AppendEmpty()
	for _, name := range received {
		m := sm.Metrics().AppendEmpty()
		m.SetName(name)
		fill[name](m, newest[name])
	}

	store := NewStore(clockwork.NewFakeClockAt(sortTestTime(100)))
	store.AddMetric(&metrics)

	assert.Equal(t, []string{"gauge", "sum", "histogram", "exponential histogram", "summary"}, metricNames(*store.GetFilteredMetrics()))
}
