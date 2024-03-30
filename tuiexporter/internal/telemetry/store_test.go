package telemetry

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// This is written referencing following code: https://github.com/CtrlSpice/otel-desktop-viewer/blob/af38ec47a37564e5f03b6d9cefa20b2422033e03/desktopexporter/testdata/trace.go
var (
	spanStartTimestamp = pcommon.NewTimestampFromTime(time.Date(2022, 10, 21, 7, 10, 2, 100, time.UTC))
	spanEventTimestamp = pcommon.NewTimestampFromTime(time.Date(2020, 10, 21, 7, 10, 2, 150, time.UTC))
	spanEndTimestamp   = pcommon.NewTimestampFromTime(time.Date(2020, 10, 21, 7, 10, 2, 300, time.UTC))
)

type generatedSpans struct {
	spans  []*ptrace.Span
	rspans []*ptrace.ResourceSpans
	sspans []*ptrace.ScopeSpans
}

func TestSpanDataIsRoot(t *testing.T) {
	_, testdata := generateOTLPPayload(t, 1, 1, []int{1}, [][]int{{2}})
	parentSpan := testdata.spans[0]
	childSpan := testdata.spans[1]
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
	payload, testdata := generateOTLPPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	traceID := testdata.spans[0].TraceID().String()
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
					Span:         testdata.spans[3],  // span-2-1-1
					ResourceSpan: testdata.rspans[1], // test-service-2
					ScopeSpans:   testdata.sspans[2], // test-scope-2-1
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
	payload, testdata := generateOTLPPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	store.AddSpan(&payload)

	assert.Equal(t, "", store.filterSvc)
	assert.True(t, before.Before(store.updatedAt))

	// assert svcspans
	assert.Equal(t, 2, len(store.svcspans))
	assert.Equal(t, testdata.spans[0], store.svcspans[0].Span)          // span-1-1-1
	assert.Equal(t, testdata.rspans[0], store.svcspans[0].ResourceSpan) // test-service-1
	assert.Equal(t, testdata.sspans[0], store.svcspans[0].ScopeSpans)   // test-scope-1-1
	assert.Equal(t, testdata.spans[3], store.svcspans[1].Span)          // span-2-1-1
	assert.Equal(t, testdata.rspans[1], store.svcspans[1].ResourceSpan) // test-service-2
	assert.Equal(t, testdata.sspans[2], store.svcspans[1].ScopeSpans)   // test-scope-2-1

	// assert svcspansFiltered
	assert.Equal(t, 2, len(store.svcspansFiltered))
	assert.Equal(t, testdata.spans[0], store.svcspansFiltered[0].Span) // span-1-1-1
	assert.Equal(t, testdata.spans[3], store.svcspansFiltered[1].Span) // span-2-1-1

	// assert cache spanid2span
	assert.Equal(t, 4, len(store.cache.spanid2span))
	for _, span := range testdata.spans {
		got, _ := store.cache.spanid2span[span.SpanID().String()]
		assert.Equal(t, span, got.Span)
	}

	// assert cache traceid2spans
	{
		gotsd := store.cache.traceid2spans[testdata.spans[0].TraceID().String()]
		assert.Equal(t, 4, len(gotsd))
		gotspans := []*ptrace.Span{}
		for _, sd := range gotsd {
			gotspans = append(gotspans, sd.Span)
		}
		assert.ElementsMatch(t, testdata.spans, gotspans)
	}

	// assert cache tracesvc2spans
	{
		assert.Equal(t, 1, len(store.cache.tracesvc2spans))
		gotsds := store.cache.tracesvc2spans[testdata.spans[0].TraceID().String()]
		assert.Equal(t, 2, len(gotsds))
		assert.Equal(t, testdata.spans[0], gotsds["test-service-1"][0].Span) // span-1-1-1
		assert.Equal(t, testdata.spans[1], gotsds["test-service-1"][1].Span) // span-1-1-2
		assert.Equal(t, testdata.spans[2], gotsds["test-service-1"][2].Span) // span-1-1-3
		assert.Equal(t, testdata.spans[3], gotsds["test-service-2"][0].Span) // span-2-1-1
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
	payload1, _ := generateOTLPPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	payload2, testdata2 := generateOTLPPayload(t, 2, 1, []int{1}, [][]int{{1}})
	store.AddSpan(&payload1)
	store.AddSpan(&payload2)

	// assert rotation
	// assert svcspans
	assert.Equal(t, 1, len(store.svcspans))
	assert.Equal(t, testdata2.spans[0], store.svcspans[0].Span)          // trace 2, span-1-1-1
	assert.Equal(t, testdata2.rspans[0], store.svcspans[0].ResourceSpan) // trace 2, test-service-1
	assert.Equal(t, testdata2.sspans[0], store.svcspans[0].ScopeSpans)   // trace 2, test-scope-1-1

	// assert svcspansFiltered
	assert.Equal(t, 1, len(store.svcspansFiltered))
	assert.Equal(t, testdata2.spans[0], store.svcspansFiltered[0].Span) // trace 2, span-1-1-1

	// assert cache spanid2span
	assert.Equal(t, 1, len(store.cache.spanid2span))
	{
		want := testdata2.spans[0]
		got, _ := store.cache.spanid2span[want.SpanID().String()]
		assert.Equal(t, want, got.Span)
	}

	// assert cache traceid2spans
	{
		gotsd := store.cache.traceid2spans[testdata2.spans[0].TraceID().String()]
		assert.Equal(t, 1, len(gotsd))
		assert.Equal(t, testdata2.spans[0], gotsd[0].Span)
	}

	// assert cache tracesvc2spans
	{
		assert.Equal(t, 1, len(store.cache.tracesvc2spans))
		gotsds := store.cache.tracesvc2spans[testdata2.spans[0].TraceID().String()]
		assert.Equal(t, 1, len(gotsds))
		assert.Equal(t, testdata2.spans[0], gotsds["test-service-1"][0].Span) // trace 2, span-1-1-1
	}
}

// This is written referencing following code: https://github.com/CtrlSpice/otel-desktop-viewer/blob/af38ec47a37564e5f03b6d9cefa20b2422033e03/desktopexporter/testdata/trace.go
func generateOTLPPayload(t *testing.T, traceID, resourceCount int, scopeCount []int, spanCount [][]int) (ptrace.Traces, *generatedSpans) {
	t.Helper()

	generatedSpans := &generatedSpans{
		spans:  []*ptrace.Span{},
		rspans: []*ptrace.ResourceSpans{},
		sspans: []*ptrace.ScopeSpans{},
	}
	traceData := ptrace.NewTraces()
	uniqueSpanIndex := 0

	// Create and populate resource data
	traceData.ResourceSpans().EnsureCapacity(resourceCount)
	for resourceIndex := 0; resourceIndex < resourceCount; resourceIndex++ {
		scopeCount := scopeCount[resourceIndex]
		resourceSpan := traceData.ResourceSpans().AppendEmpty()
		fillResource(t, resourceSpan.Resource(), resourceIndex)
		generatedSpans.rspans = append(generatedSpans.rspans, &resourceSpan)

		// Create and populate instrumentation scope data
		resourceSpan.ScopeSpans().EnsureCapacity(scopeCount)
		for scopeIndex := 0; scopeIndex < scopeCount; scopeIndex++ {
			spanCount := spanCount[resourceIndex][scopeIndex]
			scopeSpan := resourceSpan.ScopeSpans().AppendEmpty()
			fillScope(t, scopeSpan.Scope(), resourceIndex, scopeIndex)
			generatedSpans.sspans = append(generatedSpans.sspans, &scopeSpan)

			//Create and populate spans
			scopeSpan.Spans().EnsureCapacity(spanCount)
			for spanIndex := 0; spanIndex < spanCount; spanIndex++ {
				span := scopeSpan.Spans().AppendEmpty()
				fillSpan(t, span, traceID, resourceIndex, scopeIndex, spanIndex, uniqueSpanIndex)
				generatedSpans.spans = append(generatedSpans.spans, &span)
				uniqueSpanIndex++
			}
		}
	}

	return traceData, generatedSpans
}

// This is written referencing following code: https://github.com/CtrlSpice/otel-desktop-viewer/blob/af38ec47a37564e5f03b6d9cefa20b2422033e03/desktopexporter/testdata/trace.go
func fillResource(t *testing.T, resource pcommon.Resource, resourceIndex int) {
	t.Helper()
	resource.SetDroppedAttributesCount(1)
	resource.Attributes().PutStr("service.name", fmt.Sprintf("test-service-%d", resourceIndex+1))
	resource.Attributes().PutStr("resource attribute", "resource attribute value")
	resource.Attributes().PutInt("resource index", int64(resourceIndex))
}

// This is written referencing following code: https://github.com/CtrlSpice/otel-desktop-viewer/blob/af38ec47a37564e5f03b6d9cefa20b2422033e03/desktopexporter/testdata/trace.go
func fillScope(t *testing.T, scope pcommon.InstrumentationScope, resourceIndex, scopeIndex int) {
	t.Helper()
	scope.SetDroppedAttributesCount(2)
	scope.SetName(fmt.Sprintf("test-scope-%d-%d", resourceIndex+1, scopeIndex+1))
	scope.SetVersion("v0.0.1")
	scope.Attributes().PutInt("scope index", int64(scopeIndex))
}

// This is written referencing following code: https://github.com/CtrlSpice/otel-desktop-viewer/blob/af38ec47a37564e5f03b6d9cefa20b2422033e03/desktopexporter/testdata/trace.go
func fillSpan(t *testing.T, span ptrace.Span, traceID, resourceIndex, scopeIndex, spanIndex, uniqueSpanIndex int) {
	t.Helper()
	spanID := [8]byte{byte(uniqueSpanIndex + 1)}

	span.SetName(fmt.Sprintf("span-%d-%d-%d", resourceIndex, scopeIndex, spanIndex))
	span.SetKind(ptrace.SpanKindInternal)
	span.SetStartTimestamp(spanStartTimestamp)
	span.SetEndTimestamp(spanEndTimestamp)
	span.SetDroppedAttributesCount(3)
	span.SetTraceID([16]byte{byte(traceID)})
	span.SetSpanID(spanID)
	span.SetParentSpanID([8]byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28})
	span.Attributes().PutInt("span index", int64(spanIndex))
	span.SetDroppedAttributesCount(3)
	span.SetDroppedEventsCount(4)
	span.SetDroppedLinksCount(5)

	event := span.Events().AppendEmpty()
	event.SetTimestamp(spanEventTimestamp)
	event.SetName("span event")
	event.Attributes().PutStr("span event attribute", "span event attribute value")
	event.SetDroppedAttributesCount(6)

	link := span.Links().AppendEmpty()
	link.SetTraceID([16]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10})
	link.Attributes().PutStr("span link attribute", "span link attribute value")
	link.SetDroppedAttributesCount(7)

	status := span.Status()
	status.SetCode(ptrace.StatusCodeOk)
	status.SetMessage("status ok")
}
