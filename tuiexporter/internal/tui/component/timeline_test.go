package component

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
)

func TestNewSpanTreeWithServiceName(t *testing.T) {
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
	payload, testdata := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{6}})
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

	st, d := newSpanTree(testdata.Spans[0].TraceID().String(), store.GetTraceCache())

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

func TestNewSpanTreeWithoutServiceName(t *testing.T) {
	// traceid: 1
	//  └- resource: [Empty]
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-1 [root]
	//        └- span: span-1-1-2
	//          └- span: span-1-1-3
	//        └- span: span-1-1-4
	//      └- span: span-1-1-5 [root] multiple root span is allowed
	//        └- span: span-1-1-6
	store := telemetry.NewStore()
	payload, testdata := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{6}})
	testdata.RSpans[0].Resource().Attributes().Clear()
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

	st, d := newSpanTree(testdata.Spans[0].TraceID().String(), store.GetTraceCache())

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

func TestRoundDownDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "0",
			duration: 0,
			want:     "0s", // do nothing
		},
		{
			name:     "10 nano second",
			duration: 10 * time.Nanosecond,
			want:     "10ns", // do nothing
		},
		{
			name:     "10.25 micro second",
			duration: 10250 * time.Nanosecond,
			want:     "10µs",
		},
		{
			name:     "10.25 milli second",
			duration: 10250 * time.Microsecond,
			want:     "10ms",
		},
		{
			name:     "10.25 second",
			duration: 10250 * time.Millisecond,
			want:     "10s",
		},
		{
			name:     "10.25 minute",
			duration: 615 * time.Second,
			want:     "10m0s",
		},
		{
			name:     "10.25 hour",
			duration: 615 * time.Minute,
			want:     "10h15m0s", // do nothing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := roundDownDuration(tt.duration)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestNarrowInLimit(t *testing.T) {
	tests := []struct {
		name  string
		step  int
		curr  int
		limit int
		want  int
	}{
		{
			name:  "Modified",
			step:  5,
			curr:  35,
			limit: 30,
			want:  30,
		},
		{
			name:  "No_Effect_Limit_Over",
			step:  5,
			curr:  34,
			limit: 30,
			want:  34,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := narrowInLimit(tt.step, tt.curr, tt.limit)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWidenInLimit(t *testing.T) {
	tests := []struct {
		name  string
		step  int
		curr  int
		limit int
		want  int
	}{
		{
			name:  "Modified",
			step:  5,
			curr:  35,
			limit: 40,
			want:  40,
		},
		{
			name:  "No_Effect_Limit_Over",
			step:  5,
			curr:  30,
			limit: 34,
			want:  30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := widenInLimit(tt.step, tt.curr, tt.limit)
			assert.Equal(t, tt.want, got)
		})
	}
}
