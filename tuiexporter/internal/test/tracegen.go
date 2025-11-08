package test

import (
	"fmt"
	"testing"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// This is written referencing following code: https://github.com/CtrlSpice/otel-desktop-viewer/blob/af38ec47a37564e5f03b6d9cefa20b2422033e03/desktopexporter/testdata/trace.go
var (
	spanStartTimestamp = pcommon.NewTimestampFromTime(time.Date(2022, 10, 21, 7, 10, 2, 100000000, time.UTC))
	spanEventTimestamp = pcommon.NewTimestampFromTime(time.Date(2022, 10, 21, 7, 10, 2, 150000000, time.UTC))
	spanEndTimestamp   = pcommon.NewTimestampFromTime(time.Date(2022, 10, 21, 7, 10, 2, 300000000, time.UTC))
)

type GeneratedSpans struct {
	Spans  []*ptrace.Span
	RSpans []*ptrace.ResourceSpans
	SSpans []*ptrace.ScopeSpans
}

// This is written referencing following code: https://github.com/CtrlSpice/otel-desktop-viewer/blob/af38ec47a37564e5f03b6d9cefa20b2422033e03/desktopexporter/testdata/trace.go
func GenerateOTLPTracesPayload(t *testing.T, traceID, resourceCount int, scopeCount []int, spanCount [][]int) (ptrace.Traces, *GeneratedSpans) {
	t.Helper()

	generatedSpans := &GeneratedSpans{
		Spans:  []*ptrace.Span{},
		RSpans: []*ptrace.ResourceSpans{},
		SSpans: []*ptrace.ScopeSpans{},
	}
	traceData := ptrace.NewTraces()
	uniqueSpanIndex := 0

	// Create and populate resource data
	traceData.ResourceSpans().EnsureCapacity(resourceCount)
	for resourceIndex := range resourceCount {
		scopeCount := scopeCount[resourceIndex]
		resourceSpan := traceData.ResourceSpans().AppendEmpty()
		fillResource(t, resourceSpan.Resource(), resourceIndex)
		generatedSpans.RSpans = append(generatedSpans.RSpans, &resourceSpan)

		// Create and populate instrumentation scope data
		resourceSpan.ScopeSpans().EnsureCapacity(scopeCount)
		for scopeIndex := range scopeCount {
			spanCount := spanCount[resourceIndex][scopeIndex]
			scopeSpan := resourceSpan.ScopeSpans().AppendEmpty()
			fillScope(t, scopeSpan.Scope(), resourceIndex, scopeIndex)
			generatedSpans.SSpans = append(generatedSpans.SSpans, &scopeSpan)

			//Create and populate spans
			scopeSpan.Spans().EnsureCapacity(spanCount)
			for spanIndex := range spanCount {
				span := scopeSpan.Spans().AppendEmpty()
				fillSpan(t, span, traceID, resourceIndex, scopeIndex, spanIndex, uniqueSpanIndex)
				generatedSpans.Spans = append(generatedSpans.Spans, &span)
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

// GenerateSpanWithDuration returns a span with specified span name and duration.
func GenerateSpanWithDuration(t *testing.T, spanName string, duration time.Duration) *ptrace.Span {
	t.Helper()

	span := ptrace.NewSpan()
	span.SetName(spanName)
	endTimeStamp := pcommon.NewTimestampFromTime(spanStartTimestamp.AsTime().Add(duration))
	span.SetStartTimestamp(spanStartTimestamp)
	span.SetEndTimestamp(endTimeStamp)

	return &span
}
