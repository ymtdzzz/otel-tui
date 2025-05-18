package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestSpanDataIsRoot(t *testing.T) {
	_, testdata := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{2}})
	parentSpan := testdata.Spans[0]
	childSpan := testdata.Spans[1]
	parentSpan.SetParentSpanID(pcommon.SpanID{})
	childSpan.SetParentSpanID(parentSpan.SpanID())

	tests := []struct {
		name string
		span *SpanData
		want bool
	}{
		{
			name: "true",
			span: &SpanData{Span: parentSpan},
			want: true,
		},
		{
			name: "false",
			span: &SpanData{Span: childSpan},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.span.IsRoot())
		})
	}
}

func TestStoreGetter(t *testing.T) {
	store := NewStore()
	assert.Equal(t, store.tracecache, store.GetTraceCache())
	assert.Equal(t, store.metriccache, store.GetMetricCache())
	assert.Equal(t, store.logcache, store.GetLogCache())
	assert.Equal(t, &store.svcspans, store.GetSvcSpans())
	assert.Equal(t, &store.svcspansFiltered, store.GetFilteredSvcSpans())
	assert.Equal(t, &store.metricsFiltered, store.GetFilteredMetrics())
	assert.Equal(t, &store.logsFiltered, store.GetFilteredLogs())
	assert.Equal(t, store.updatedAt, store.UpdatedAt())
}

func TestStoreSpanFilters(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1
	//  | | └- span: span-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  └- resource: test-service-2
	//  | └- scope: test-scope-2-1
	//  |   └- span: span-2-1-1
	//  └- resource: [Empty]
	//    └- scope: test-scope-3-1
	//      └- span: span-3-1-1
	store := NewStore()
	payload, testdata := test.GenerateOTLPTracesPayload(t, 1, 3, []int{2, 1, 1}, [][]int{{2, 1}, {1}, {1}})
	traceID := testdata.Spans[0].TraceID().String()
	testdata.RSpans[2].Resource().Attributes().Clear()
	store.AddSpan(&payload)

	store.ApplyFilterTraces("0-0", SORT_TYPE_NONE)
	assert.Equal(t, 3, len(store.svcspansFiltered))
	assert.Equal(t, traceID, store.GetTraceIDByFilteredIdx(0))
	assert.Equal(t, traceID, store.GetTraceIDByFilteredIdx(1))
	assert.Equal(t, "", store.GetTraceIDByFilteredIdx(3))
	// spans in test-service-1
	assert.Equal(t, "span-0-0-0", store.GetFilteredServiceSpansByIdx(0)[0].Span.Name())
	assert.Equal(t, "span-0-0-1", store.GetFilteredServiceSpansByIdx(0)[1].Span.Name())
	// spans in test-service-2
	assert.Equal(t, "span-1-0-0", store.GetFilteredServiceSpansByIdx(1)[0].Span.Name())
	store.ApplyFilterTraces("service-2", SORT_TYPE_NONE)
	assert.Equal(t, 1, len(store.svcspansFiltered))
	assert.Equal(t, traceID, store.GetTraceIDByFilteredIdx(0))
	assert.Equal(t, "", store.GetTraceIDByFilteredIdx(1))
	assert.Equal(t, "span-1-0-0", store.GetFilteredServiceSpansByIdx(0)[0].Span.Name())

	tests := []struct {
		name string
		idx  int
		want []*SpanData
	}{
		{
			name: "invalid index",
			idx:  2,
			want: nil,
		},
		{
			name: "valid index",
			idx:  0,
			want: []*SpanData{
				{
					Span:         testdata.Spans[3],  // span-2-1-1
					ResourceSpan: testdata.RSpans[1], // test-service-2
					ScopeSpans:   testdata.SSpans[2], // test-scope-2-1
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run("GetFilteredServiceSpansByIdx_"+tt.name, func(t *testing.T) {
			got := store.GetFilteredServiceSpansByIdx(tt.idx)
			assert.Equal(t, len(tt.want), len(got))
			for idx, want := range tt.want {
				assert.Equal(t, want.Span, got[idx].Span)
				assert.Equal(t, want.ResourceSpan, got[idx].ResourceSpan)
				assert.Equal(t, want.ScopeSpans, got[idx].ScopeSpans)
			}
		})
	}

	// spans in unknown service
	store.ApplyFilterTraces("unknown", SORT_TYPE_NONE)
	assert.Equal(t, "span-2-0-0", store.GetFilteredServiceSpansByIdx(0)[0].Span.Name())
}

func TestStoreMetricFilters(t *testing.T) {
	// metric: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- metric: metric-1-1-1
	//  | |   └- datapoint: dp-1-1-1-1
	//  | |   └- datapoint: dp-1-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- metric: metric-1-2-1
	//  |     └- datapoint: dp-1-2-1-1
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- metric: metric-2-1-1
	//        └- datapoint: dp-2-1-1-1
	store := NewStore()
	payload, testdata := test.GenerateOTLPGaugeMetricsPayload(t, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	store.AddMetric(&payload)

	store.ApplyFilterMetrics("service-2")
	assert.Equal(t, 1, len(store.metricsFiltered))
	store.ApplyFilterMetrics("metric 0")
	assert.Equal(t, 2, len(store.metricsFiltered))

	tests := []struct {
		name string
		idx  int
		want *MetricData
	}{
		{
			name: "invalid index",
			idx:  2,
			want: nil,
		},
		{
			name: "valid index",
			idx:  1,
			want: &MetricData{
				Metric:         testdata.Metrics[1],  // metric-1-2-1
				ResourceMetric: testdata.RMetrics[0], // test-service-1
				ScopeMetric:    testdata.SMetrics[1], // test-scope-1-2
			},
		},
	}

	for _, tt := range tests {
		t.Run("GetFilteredMetricByIdx_"+tt.name, func(t *testing.T) {
			got := store.GetFilteredMetricByIdx(tt.idx)
			if tt.want != nil {
				assert.Equal(t, tt.want.Metric, got.Metric)
				assert.Equal(t, tt.want.ResourceMetric, got.ResourceMetric)
				assert.Equal(t, tt.want.ScopeMetric, got.ScopeMetric)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestStoreLogFilters(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1
	//  | | | └- log: log-1-1-1-1
	//  | | | └- log: log-1-1-1-2
	//  | | └- span: span-1-1-2
	//  | |   └- log: log-1-1-2-1
	//  | |   └- log: log-1-1-2-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  |     └- log: log-1-2-3-1
	//  |     └- log: log-1-2-3-2
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- span: span-2-1-1
	//        └- log: log-2-1-1-1
	//        └- log: log-2-1-1-2
	store := NewStore()
	payload, testdata := test.GenerateOTLPLogsPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	store.AddLog(&payload)

	store.ApplyFilterLogs("service-2")
	assert.Equal(t, 2, len(store.logsFiltered))
	store.ApplyFilterLogs("log body 1-0-0-0")
	assert.Equal(t, 1, len(store.logsFiltered))

	tests := []struct {
		name string
		idx  int
		want *LogData
	}{
		{
			name: "invalid index",
			idx:  1,
			want: nil,
		},
		{
			name: "valid index",
			idx:  0,
			want: &LogData{
				Log:         testdata.Logs[6],  // span-2-1-1
				ResourceLog: testdata.RLogs[1], // test-service-2
				ScopeLog:    testdata.SLogs[2], // test-scope-2-1
			},
		},
	}

	for _, tt := range tests {
		t.Run("GetFilteredLogByIdx_"+tt.name, func(t *testing.T) {
			got := store.GetFilteredLogByIdx(tt.idx)
			if tt.want != nil {
				assert.Equal(t, tt.want.Log, got.Log)
				assert.Equal(t, tt.want.ResourceLog, got.ResourceLog)
				assert.Equal(t, tt.want.ScopeLog, got.ScopeLog)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestStoreAddSpanWithoutRotation(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1
	//  | | └- span: span-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- span: span-2-1-1
	store := NewStore()
	store.maxServiceSpanCount = 2 // no rotation
	before := store.updatedAt
	payload, testdata := test.GenerateOTLPTracesPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	store.AddSpan(&payload)

	assert.Equal(t, "", store.filterSvc)
	assert.True(t, before.Before(store.updatedAt))

	// assert svcspans
	assert.Equal(t, 2, len(store.svcspans))
	assert.Equal(t, testdata.Spans[0], store.svcspans[0].Span)          // span-1-1-1
	assert.Equal(t, testdata.RSpans[0], store.svcspans[0].ResourceSpan) // test-service-1
	assert.Equal(t, testdata.SSpans[0], store.svcspans[0].ScopeSpans)   // test-scope-1-1
	assert.Equal(t, testdata.Spans[3], store.svcspans[1].Span)          // span-2-1-1
	assert.Equal(t, testdata.RSpans[1], store.svcspans[1].ResourceSpan) // test-service-2
	assert.Equal(t, testdata.SSpans[2], store.svcspans[1].ScopeSpans)   // test-scope-2-1

	// assert svcspansFiltered
	assert.Equal(t, 2, len(store.svcspansFiltered))
	assert.Equal(t, testdata.Spans[0], store.svcspansFiltered[0].Span) // span-1-1-1
	assert.Equal(t, testdata.Spans[3], store.svcspansFiltered[1].Span) // span-2-1-1

	// assert cache spanid2span
	assert.Equal(t, 4, len(store.tracecache.spanid2span))
	for _, span := range testdata.Spans {
		got := store.tracecache.spanid2span[span.SpanID().String()]
		assert.Equal(t, span, got.Span)
	}

	// assert cache traceid2spans
	{
		gotsd := store.tracecache.traceid2spans[testdata.Spans[0].TraceID().String()]
		assert.Equal(t, 4, len(gotsd))
		gotspans := []*ptrace.Span{}
		for _, sd := range gotsd {
			gotspans = append(gotspans, sd.Span)
		}
		assert.ElementsMatch(t, testdata.Spans, gotspans)
	}

	// assert cache tracesvc2spans
	{
		assert.Equal(t, 1, len(store.tracecache.tracesvc2spans))
		gotsds := store.tracecache.tracesvc2spans[testdata.Spans[0].TraceID().String()]
		assert.Equal(t, 2, len(gotsds))
		assert.Equal(t, testdata.Spans[0], gotsds["test-service-1"][0].Span) // span-1-1-1
		assert.Equal(t, testdata.Spans[1], gotsds["test-service-1"][1].Span) // span-1-1-2
		assert.Equal(t, testdata.Spans[2], gotsds["test-service-1"][2].Span) // span-1-1-3
		assert.Equal(t, testdata.Spans[3], gotsds["test-service-2"][0].Span) // span-2-1-1
	}
}

func TestStoreAddSpanWithRotation(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1
	//  | | └- span: span-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- span: span-2-1-1
	// traceid: 2
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-1
	store := NewStore()
	store.maxServiceSpanCount = 1
	payload1, _ := test.GenerateOTLPTracesPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	payload2, testdata2 := test.GenerateOTLPTracesPayload(t, 2, 1, []int{1}, [][]int{{1}})
	store.AddSpan(&payload1)
	store.AddSpan(&payload2)

	// assert rotation
	// assert svcspans
	assert.Equal(t, 1, len(store.svcspans))
	assert.Equal(t, testdata2.Spans[0], store.svcspans[0].Span)          // trace 2, span-1-1-1
	assert.Equal(t, testdata2.RSpans[0], store.svcspans[0].ResourceSpan) // trace 2, test-service-1
	assert.Equal(t, testdata2.SSpans[0], store.svcspans[0].ScopeSpans)   // trace 2, test-scope-1-1

	// assert svcspansFiltered
	assert.Equal(t, 1, len(store.svcspansFiltered))
	assert.Equal(t, testdata2.Spans[0], store.svcspansFiltered[0].Span) // trace 2, span-1-1-1

	// assert cache spanid2span
	assert.Equal(t, 1, len(store.tracecache.spanid2span))
	{
		want := testdata2.Spans[0]
		got := store.tracecache.spanid2span[want.SpanID().String()]
		assert.Equal(t, want, got.Span)
	}

	// assert cache traceid2spans
	{
		gotsd := store.tracecache.traceid2spans[testdata2.Spans[0].TraceID().String()]
		assert.Equal(t, 1, len(gotsd))
		assert.Equal(t, testdata2.Spans[0], gotsd[0].Span)
	}

	// assert cache tracesvc2spans
	{
		assert.Equal(t, 1, len(store.tracecache.tracesvc2spans))
		gotsds := store.tracecache.tracesvc2spans[testdata2.Spans[0].TraceID().String()]
		assert.Equal(t, 1, len(gotsds))
		assert.Equal(t, testdata2.Spans[0], gotsds["test-service-1"][0].Span) // trace 2, span-1-1-1
	}
}

func TestStoreAddSpanServiceSpanCalculation(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-2
	//      └- span: span-1-1-3
	// traceid: 1 (the same trace)
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-1
	store := NewStore()
	store.maxServiceSpanCount = 1
	payload1, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{3}, [][]int{{3, 0, 0}})
	payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).SetParentSpanID(
		payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SpanID(),
	)
	payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(2).SetParentSpanID(
		payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).SpanID(),
	)
	payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().RemoveIf(func(s ptrace.Span) bool {
		return s.Name() == "span-0-0-0"
	})
	payload2, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{3}, [][]int{{3, 0, 0}})
	payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).SetParentSpanID(
		payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SpanID(),
	)
	payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(2).SetParentSpanID(
		payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).SpanID(),
	)
	payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().RemoveIf(func(s ptrace.Span) bool {
		return s.Name() == "span-0-0-1" || s.Name() == "span-0-0-2"
	})

	assert.Equal(t, 2, payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().Len())
	assert.Equal(t, 1, payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().Len())

	store.AddSpan(&payload1)

	// The service root span should be span-1-1-2
	assert.Equal(t, 1, len(store.svcspans))
	assert.Equal(t, "span-0-0-1", store.svcspans[0].Span.Name())

	store.AddSpan(&payload2)

	// Now, The service root span should be span-1-1-1
	assert.Equal(t, 1, len(store.svcspans))
	assert.Equal(t, "span-0-0-0", store.svcspans[0].Span.Name())
}

func TestStoreAddSpanServiceSpanCalculationLimitation(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-3
	// traceid: 1 (the same trace)
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-1
	// traceid: 1 (the same trace)
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-2
	store := NewStore()
	store.maxServiceSpanCount = 1
	payload1, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{3}, [][]int{{3, 0, 0}})
	payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).SetParentSpanID(
		payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SpanID(),
	)
	payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(2).SetParentSpanID(
		payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).SpanID(),
	)
	payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().RemoveIf(func(s ptrace.Span) bool {
		return s.Name() == "span-0-0-0" || s.Name() == "span-0-0-1"
	})
	payload2, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{3}, [][]int{{3, 0, 0}})
	payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).SetParentSpanID(
		payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SpanID(),
	)
	payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(2).SetParentSpanID(
		payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).SpanID(),
	)
	payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().RemoveIf(func(s ptrace.Span) bool {
		return s.Name() == "span-0-0-1" || s.Name() == "span-0-0-2"
	})
	payload3, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{3}, [][]int{{3, 0, 0}})
	payload3.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).SetParentSpanID(
		payload3.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SpanID(),
	)
	payload3.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(2).SetParentSpanID(
		payload3.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1).SpanID(),
	)
	payload3.ResourceSpans().At(0).ScopeSpans().At(0).Spans().RemoveIf(func(s ptrace.Span) bool {
		return s.Name() == "span-0-0-0" || s.Name() == "span-0-0-2"
	})

	assert.Equal(t, 1, payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().Len())
	assert.Equal(t, 1, payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().Len())
	assert.Equal(t, 1, payload3.ResourceSpans().At(0).ScopeSpans().At(0).Spans().Len())

	store.AddSpan(&payload1)

	// The service root span should be span-1-1-3
	assert.Equal(t, 1, len(store.svcspans))
	assert.Equal(t, "span-0-0-2", store.svcspans[0].Span.Name())

	store.AddSpan(&payload2)

	// The service root span should still be span-1-1-3
	assert.Equal(t, 1, len(store.svcspans))
	assert.Equal(t, "span-0-0-2", store.svcspans[0].Span.Name())

	store.AddSpan(&payload3)

	// Finally, The service root span should be span-1-1-2
	assert.Equal(t, 1, len(store.svcspans))
	assert.Equal(t, "span-0-0-1", store.svcspans[0].Span.Name())

	// By RecalculateServiceRootSpanByIdx, we can get span-1-1-1 as the root span
	store.RecalculateServiceRootSpanByIdx(0)
	assert.Equal(t, 1, len(store.svcspans))
	assert.Equal(t, "span-0-0-0", store.svcspans[0].Span.Name())
}

func TestStoreAddMetricWithoutRotation(t *testing.T) {
	// metric: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- metric: metric-1-1-1
	//  | |   └- datapoint: dp-1-1-1-1
	//  | |   └- datapoint: dp-1-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- metric: metric-1-2-1
	//  |     └- datapoint: dp-1-2-1-1
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- metric: metric-2-1-1
	//        └- datapoint: dp-2-1-1-1
	store := NewStore()
	store.maxMetricCount = 3 // no rotation
	before := store.updatedAt
	payload, testdata := test.GenerateOTLPGaugeMetricsPayload(t, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	store.AddMetric(&payload)

	assert.Equal(t, "", store.filterMetric)
	assert.True(t, before.Before(store.updatedAt))

	// assert metrics
	assert.Equal(t, 3, len(store.metrics))
	assert.Equal(t, testdata.Metrics[0], store.metrics[0].Metric)          // metric-1-1-1
	assert.Equal(t, testdata.RMetrics[0], store.metrics[0].ResourceMetric) // test-service-1
	assert.Equal(t, testdata.SMetrics[0], store.metrics[0].ScopeMetric)    // test-scope-1-1
	assert.Equal(t, testdata.Metrics[2], store.metrics[2].Metric)          // metric-2-1-1
	assert.Equal(t, testdata.RMetrics[1], store.metrics[2].ResourceMetric) // test-service-2
	assert.Equal(t, testdata.SMetrics[2], store.metrics[2].ScopeMetric)    // test-scope-2-1

	// assert metricsFiltered
	assert.Equal(t, 3, len(store.metricsFiltered))
	assert.Equal(t, testdata.Metrics[0], store.metricsFiltered[0].Metric) // metric-1-1-1
	assert.Equal(t, testdata.Metrics[2], store.metricsFiltered[2].Metric) // metric-2-1-1

	// assert cache svcmetric2metrics
	assert.Equal(t, 2, len(store.metriccache.svcmetric2metrics))
	assert.Equal(t, 2, len(store.metriccache.svcmetric2metrics["test-service-1"])) // metric-1-1-1, metric-1-2-1
	assert.Equal(t, 1, len(store.metriccache.svcmetric2metrics["test-service-2"])) // metric-2-1-1
	assert.Equal(t, testdata.Metrics[1], store.metriccache.svcmetric2metrics["test-service-1"]["metric 0-1"][0].Metric)
}

func TestStoreAddMetricWithRotation(t *testing.T) {
	// metric: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- metric: metric-1-1-1
	//  | |   └- datapoint: dp-1-1-1-1
	//  | |   └- datapoint: dp-1-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- metric: metric-1-2-1
	//  |     └- datapoint: dp-1-2-1-1
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- metric: metric-2-1-1
	//        └- datapoint: dp-2-1-1-1
	store := NewStore()
	store.maxMetricCount = 1 // no rotation
	before := store.updatedAt
	payload, testdata := test.GenerateOTLPGaugeMetricsPayload(t, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	store.AddMetric(&payload)

	assert.Equal(t, "", store.filterMetric)
	assert.True(t, before.Before(store.updatedAt))

	// assert metrics
	assert.Equal(t, 1, len(store.metrics))
	assert.Equal(t, testdata.Metrics[2], store.metrics[0].Metric)          // metric-2-1-1
	assert.Equal(t, testdata.RMetrics[1], store.metrics[0].ResourceMetric) // test-service-2
	assert.Equal(t, testdata.SMetrics[2], store.metrics[0].ScopeMetric)    // test-scope-2-1

	// assert metricsFiltered
	assert.Equal(t, 1, len(store.metricsFiltered))
	assert.Equal(t, testdata.Metrics[2], store.metricsFiltered[0].Metric) // metric-2-1-1

	// assert cache svcmetric2metrics
	assert.Equal(t, 1, len(store.metriccache.svcmetric2metrics))
	assert.Equal(t, 1, len(store.metriccache.svcmetric2metrics["test-service-2"])) // metric-2-1-1
	assert.Equal(t, testdata.Metrics[2], store.metriccache.svcmetric2metrics["test-service-2"]["metric 1-0"][0].Metric)
}

func TestStoreAddLogWithoutRotation(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1
	//  | | | └- log: log-1-1-1-1
	//  | | | └- log: log-1-1-1-2
	//  | | └- span: span-1-1-2
	//  | |   └- log: log-1-1-2-1
	//  | |   └- log: log-1-1-2-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  |     └- log: log-1-2-3-1
	//  |     └- log: log-1-2-3-2
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- span: span-2-1-1
	//        └- log: log-2-1-1-1
	//        └- log: log-2-1-1-2
	store := NewStore()
	store.maxLogCount = 8 // no rotation
	before := store.updatedAt
	payload, testdata := test.GenerateOTLPLogsPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	store.AddLog(&payload)

	assert.Equal(t, "", store.filterSvc)
	assert.True(t, before.Before(store.updatedAt))

	// assert logs
	assert.Equal(t, 8, len(store.logs))
	assert.Equal(t, testdata.Logs[0], store.logs[0].Log)          // span-1-1-1
	assert.Equal(t, testdata.RLogs[0], store.logs[0].ResourceLog) // test-service-1
	assert.Equal(t, testdata.SLogs[0], store.logs[0].ScopeLog)    // test-scope-1-1
	assert.Equal(t, testdata.Logs[6], store.logs[6].Log)          // span-2-1-1
	assert.Equal(t, testdata.RLogs[1], store.logs[6].ResourceLog) // test-service-2
	assert.Equal(t, testdata.SLogs[2], store.logs[6].ScopeLog)    // test-scope-2-1

	// assert logsFiltered
	assert.Equal(t, 8, len(store.logsFiltered))
	assert.Equal(t, testdata.Logs[0], store.logsFiltered[0].Log) // span-1-1-1
	assert.Equal(t, testdata.Logs[6], store.logsFiltered[6].Log) // span-2-1-1

	// assert cache traceid2logs
	assert.Equal(t, 1, len(store.logcache.traceid2logs))
	traceID := testdata.Logs[0].TraceID().String()
	assert.Equal(t, 8, len(store.logcache.traceid2logs[traceID]))
	for _, got := range store.logcache.traceid2logs[traceID] {
		assert.Contains(t, testdata.Logs, got.Log)
	}
}

func TestStoreAddLogWithRotation(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1
	//  | | | └- log: log-1-1-1-1
	//  | | | └- log: log-1-1-1-2
	//  | | └- span: span-1-1-2
	//  | |   └- log: log-1-1-2-1
	//  | |   └- log: log-1-1-2-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  |     └- log: log-1-2-3-1
	//  |     └- log: log-1-2-3-2
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- span: span-2-1-1
	//        └- log: log-2-1-1-1
	//        └- log: log-2-1-1-2
	store := NewStore()
	store.maxLogCount = 1
	before := store.updatedAt
	payload, testdata := test.GenerateOTLPLogsPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	store.AddLog(&payload)

	assert.Equal(t, "", store.filterSvc)
	assert.True(t, before.Before(store.updatedAt))

	// assert logs
	assert.Equal(t, 1, len(store.logs))
	assert.Equal(t, testdata.Logs[7], store.logs[0].Log)          // span-2-1-1
	assert.Equal(t, testdata.RLogs[1], store.logs[0].ResourceLog) // test-service-2
	assert.Equal(t, testdata.SLogs[2], store.logs[0].ScopeLog)    // test-scope-2-1

	// assert logsFiltered
	assert.Equal(t, 1, len(store.logsFiltered))
	assert.Equal(t, testdata.Logs[7], store.logsFiltered[0].Log) // span-2-1-1

	// assert cache traceid2logs
	assert.Equal(t, 1, len(store.logcache.traceid2logs))
	traceID := testdata.Logs[0].TraceID().String()
	assert.Equal(t, 1, len(store.logcache.traceid2logs[traceID]))
	for _, got := range store.logcache.traceid2logs[traceID] {
		assert.Contains(t, testdata.Logs, got.Log)
	}
}

func TestStoreFlush(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1
	//  | | | └- log: log-1-1-1-1
	//  | | | └- log: log-1-1-1-2
	//  | | └- span: span-1-1-2
	//  | |   └- log: log-1-1-2-1
	//  | |   └- log: log-1-1-2-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  |     └- log: log-1-2-3-1
	//  |     └- log: log-1-2-3-2
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- span: span-2-1-1
	//        └- log: log-2-1-1-1
	//        └- log: log-2-1-1-2
	// traceid: 2
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-1
	//        └- log: log-1-1-1-1
	//        └- log: log-1-1-1-2
	// metric: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- metric: metric-1-1-1
	//  | |   └- datapoint: dp-1-1-1-1
	//  | |   └- datapoint: dp-1-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- metric: metric-1-2-1
	//  |     └- datapoint: dp-1-2-1-1
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- metric: metric-2-1-1
	//        └- datapoint: dp-2-1-1-1
	store := NewStore()
	store.maxServiceSpanCount = 1
	tp1, _ := test.GenerateOTLPTracesPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	tp2, _ := test.GenerateOTLPTracesPayload(t, 2, 1, []int{1}, [][]int{{1}})
	lp1, _ := test.GenerateOTLPLogsPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	lp2, _ := test.GenerateOTLPLogsPayload(t, 2, 1, []int{1}, [][]int{{1}})
	m, _ := test.GenerateOTLPGaugeMetricsPayload(t, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	store.AddSpan(&tp1)
	store.AddSpan(&tp2)
	store.AddLog(&lp1)
	store.AddLog(&lp2)
	store.AddMetric(&m)

	before := store.updatedAt
	store.Flush()

	assert.True(t, before.Before(store.updatedAt))

	// assert traces
	assert.Equal(t, 0, len(store.svcspans))
	assert.Equal(t, 0, len(store.svcspansFiltered))
	assert.Equal(t, 0, len(store.tracecache.spanid2span))
	assert.Equal(t, 0, len(store.tracecache.traceid2spans))
	assert.Equal(t, 0, len(store.tracecache.tracesvc2spans))

	// assert logs
	assert.Equal(t, 0, len(store.logs))
	assert.Equal(t, 0, len(store.logsFiltered))
	assert.Equal(t, 0, len(store.logcache.traceid2logs))

	// assert metrics
	assert.Equal(t, 0, len(store.metrics))
	assert.Equal(t, 0, len(store.metricsFiltered))
	assert.Equal(t, 0, len(store.metriccache.svcmetric2metrics))
}

func TestLogDataGetResolvedBody(t *testing.T) {
	l, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
	lr := l.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	ld := &LogData{
		Log: &lr,
	}
	lr.Body().SetStr("test log. userId={userId}, quantity={quantity}, tags={tags}")
	lr.Attributes().PutStr("userId", "user-12345")
	lr.Attributes().PutInt("quantity", 2000)
	tags := lr.Attributes().PutEmptySlice("tags")
	tags.AppendEmpty().SetStr("tag_A")
	tags.AppendEmpty().SetStr("tag_B")
	want := `test log. userId=user-12345, quantity=2000, tags=["tag_A","tag_B"]`

	assert.Equal(t, want, ld.GetResolvedBody())
}
