package component

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
)

func TestNewSpanTree(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-1 [root]
	//        └- span: span-1-1-2
	//          └- span: span-1-1-3
	//        └- span: span-1-1-4
	//      └- span: span-1-1-5 [root] multiple root span is allowed
	//        └- span: span-1-1-6
	store := telemetry.NewStore()
	payload, testdata := test.GenerateOTLPPayload(t, 1, 1, []int{1}, [][]int{{6}})
	sds := []*telemetry.SpanData{}
	for _, span := range testdata.Spans {
		sds = append(sds, &telemetry.SpanData{
			Span:         span,
			ResourceSpan: testdata.RSpans[0],
			ScopeSpans:   testdata.SSpans[0],
		})
	}
	sds[0].Span.SetParentSpanID([8]byte{byte(0)}) // root span
	sds[1].Span.SetParentSpanID(sds[0].Span.SpanID())
	sds[2].Span.SetParentSpanID(sds[1].Span.SpanID())
	sds[3].Span.SetParentSpanID(sds[0].Span.SpanID())
	sds[4].Span.SetParentSpanID([8]byte{byte(0)}) // root span
	sds[5].Span.SetParentSpanID(sds[4].Span.SpanID())

	store.AddSpan(&payload)

	st, d := newSpanTree(testdata.Spans[0].TraceID().String(), store.GetCache())

	// duration assertion
	assert.Equal(t, 200*time.Millisecond, d)

	// node assertion
	assert.Equal(t, 2, len(st))
	assert.Equal(t, sds[0].Span, st[0].span.Span)
	assert.Equal(t, sds[1].Span, st[0].children[0].span.Span)
	assert.Equal(t, sds[2].Span, st[0].children[0].children[0].span.Span)
	assert.Equal(t, sds[3].Span, st[0].children[1].span.Span)
	assert.Equal(t, sds[4].Span, st[1].span.Span)
	assert.Equal(t, sds[5].Span, st[1].children[0].span.Span)
}
