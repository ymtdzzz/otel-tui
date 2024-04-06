package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestSpanDataIsRoot(t *testing.T) {
	_, testdata := test.GenerateOTLPPayload(t, 1, 1, []int{1}, [][]int{{2}})
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
	assert.Equal(t, store.cache, store.GetCache())
	assert.Equal(t, &store.svcspans, store.GetSvcSpans())
	assert.Equal(t, &store.svcspansFiltered, store.GetFilteredSvcSpans())
	assert.Equal(t, store.updatedAt, store.UpdatedAt())
}

func TestStoreFilters(t *testing.T) {
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
	payload, testdata := test.GenerateOTLPPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	traceID := testdata.Spans[0].TraceID().String()
	store.AddSpan(&payload)

	store.ApplyFilterService("service-2")
	assert.Equal(t, 1, len(store.svcspansFiltered))
	assert.Equal(t, traceID, store.GetTraceIDByFilteredIdx(0))
	assert.Equal(t, "", store.GetTraceIDByFilteredIdx(1))

	tests := []struct {
		name string
		idx  int
		want []*SpanData
	}{
		{
			name: "invalid index",
			idx:  1,
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
	payload, testdata := test.GenerateOTLPPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
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
	assert.Equal(t, 4, len(store.cache.spanid2span))
	for _, span := range testdata.Spans {
		got := store.cache.spanid2span[span.SpanID().String()]
		assert.Equal(t, span, got.Span)
	}

	// assert cache traceid2spans
	{
		gotsd := store.cache.traceid2spans[testdata.Spans[0].TraceID().String()]
		assert.Equal(t, 4, len(gotsd))
		gotspans := []*ptrace.Span{}
		for _, sd := range gotsd {
			gotspans = append(gotspans, sd.Span)
		}
		assert.ElementsMatch(t, testdata.Spans, gotspans)
	}

	// assert cache tracesvc2spans
	{
		assert.Equal(t, 1, len(store.cache.tracesvc2spans))
		gotsds := store.cache.tracesvc2spans[testdata.Spans[0].TraceID().String()]
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
	payload1, _ := test.GenerateOTLPPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	payload2, testdata2 := test.GenerateOTLPPayload(t, 2, 1, []int{1}, [][]int{{1}})
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
	assert.Equal(t, 1, len(store.cache.spanid2span))
	{
		want := testdata2.Spans[0]
		got := store.cache.spanid2span[want.SpanID().String()]
		assert.Equal(t, want, got.Span)
	}

	// assert cache traceid2spans
	{
		gotsd := store.cache.traceid2spans[testdata2.Spans[0].TraceID().String()]
		assert.Equal(t, 1, len(gotsd))
		assert.Equal(t, testdata2.Spans[0], gotsd[0].Span)
	}

	// assert cache tracesvc2spans
	{
		assert.Equal(t, 1, len(store.cache.tracesvc2spans))
		gotsds := store.cache.tracesvc2spans[testdata2.Spans[0].TraceID().String()]
		assert.Equal(t, 1, len(gotsds))
		assert.Equal(t, testdata2.Spans[0], gotsds["test-service-1"][0].Span) // trace 2, span-1-1-1
	}
}

func TestStoreFlush(t *testing.T) {
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
	payload1, _ := test.GenerateOTLPPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	payload2, _ := test.GenerateOTLPPayload(t, 2, 1, []int{1}, [][]int{{1}})
	store.AddSpan(&payload1)
	store.AddSpan(&payload2)

	before := store.updatedAt
	store.Flush()

	assert.True(t, before.Before(store.updatedAt))
	assert.Equal(t, 0, len(store.svcspans))
	assert.Equal(t, 0, len(store.svcspansFiltered))
	assert.Equal(t, 0, len(store.cache.spanid2span))
	assert.Equal(t, 0, len(store.cache.traceid2spans))
	assert.Equal(t, 0, len(store.cache.tracesvc2spans))
}
