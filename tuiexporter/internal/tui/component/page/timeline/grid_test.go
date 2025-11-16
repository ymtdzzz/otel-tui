package timeline

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"gotest.tools/v3/assert"
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
	store := telemetry.NewStore(clockwork.NewRealClock())
	payload, testdata := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{6}})
	sds := make([]*telemetry.SpanData, 0, len(testdata.Spans))
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

	grid := newGrid(nil, store.GetTraceCache(), nil, nil, nil)
	st, d := grid.newSpanTree(testdata.Spans[0].TraceID().String())

	// duration assertion
	assert.Equal(t, 200*time.Millisecond, d)

	// node assertion
	assert.Equal(t, 2, len(st))
	assert.Equal(t, *sds[0].Span, *st[0].span.Span)
	assert.Equal(t, *sds[1].Span, *st[0].children[0].span.Span)
	assert.Equal(t, *sds[2].Span, *st[0].children[0].children[0].span.Span)
	assert.Equal(t, *sds[3].Span, *st[0].children[1].span.Span)
	assert.Equal(t, *sds[4].Span, *st[1].span.Span)
	assert.Equal(t, *sds[5].Span, *st[1].children[0].span.Span)
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
	store := telemetry.NewStore(clockwork.NewRealClock())
	payload, testdata := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{6}})
	testdata.RSpans[0].Resource().Attributes().Clear()
	sds := make([]*telemetry.SpanData, 0, len(testdata.Spans))
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

	grid := newGrid(nil, store.GetTraceCache(), nil, nil, nil)
	st, d := grid.newSpanTree(testdata.Spans[0].TraceID().String())

	// duration assertion
	assert.Equal(t, 200*time.Millisecond, d)

	// node assertion
	assert.Equal(t, 2, len(st))
	assert.Equal(t, *sds[0].Span, *st[0].span.Span)
	assert.Equal(t, *sds[1].Span, *st[0].children[0].span.Span)
	assert.Equal(t, *sds[2].Span, *st[0].children[0].children[0].span.Span)
	assert.Equal(t, *sds[3].Span, *st[0].children[1].span.Span)
	assert.Equal(t, *sds[4].Span, *st[1].span.Span)
	assert.Equal(t, *sds[5].Span, *st[1].children[0].span.Span)
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

func TestStepBy(t *testing.T) {
	store := telemetry.NewStore(clockwork.NewRealClock())
	payload, testdata := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{10}})
	store.AddSpan(&payload)

	tests := []struct {
		name           string
		initialRow     int
		totalRow       int
		step           int
		wantCurrentRow int
		wantOffsetRow  int
		wantOffsetCol  int
	}{
		{
			name:           "Forward_In_Range",
			initialRow:     2,
			totalRow:       10,
			step:           1,
			wantCurrentRow: 3,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
		{
			name:           "Backward_In_Range",
			initialRow:     5,
			totalRow:       10,
			step:           -1,
			wantCurrentRow: 4,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
		{
			name:           "Forward_To_Last",
			initialRow:     8,
			totalRow:       10,
			step:           1,
			wantCurrentRow: 9,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
		{
			name:           "Backward_To_First",
			initialRow:     1,
			totalRow:       10,
			step:           -1,
			wantCurrentRow: 0,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
		{
			name:           "Beyond_Last",
			initialRow:     9,
			totalRow:       10,
			step:           1,
			wantCurrentRow: 9,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
		{
			name:           "Before_First",
			initialRow:     0,
			totalRow:       10,
			step:           -1,
			wantCurrentRow: 0,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newGrid(nil, store.GetTraceCache(), nil, nil, nil)
			g.currentRow = tt.initialRow
			g.totalRow = tt.totalRow
			g.items = make([]*tview.TextView, tt.totalRow)
			g.nodes = make([]*spanTreeNode, tt.totalRow)
			for i := 0; i < tt.totalRow; i++ {
				g.items[i] = tview.NewTextView()
				g.nodes[i] = &spanTreeNode{span: &telemetry.SpanData{
					Span:         testdata.Spans[0],
					ResourceSpan: testdata.RSpans[0],
					ScopeSpans:   testdata.SSpans[0],
				}}
			}

			handler := g.stepBy(tt.step)
			handler(nil)

			assert.Equal(t, tt.wantCurrentRow, g.currentRow)
			row, col := g.gridView.GetOffset()
			assert.Equal(t, tt.wantOffsetRow, row)
			assert.Equal(t, tt.wantOffsetCol, col)
		})
	}
}

func TestGoToFirst(t *testing.T) {
	store := telemetry.NewStore(clockwork.NewRealClock())
	payload, testdata := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{10}})
	store.AddSpan(&payload)

	tests := []struct {
		name           string
		initialRow     int
		totalRow       int
		wantCurrentRow int
		wantOffsetRow  int
		wantOffsetCol  int
	}{
		{
			name:           "From_Middle",
			initialRow:     5,
			totalRow:       10,
			wantCurrentRow: 0,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
		{
			name:           "From_Last",
			initialRow:     9,
			totalRow:       10,
			wantCurrentRow: 0,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
		{
			name:           "Already_First",
			initialRow:     0,
			totalRow:       10,
			wantCurrentRow: 0,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newGrid(nil, store.GetTraceCache(), nil, nil, nil)
			g.currentRow = tt.initialRow
			g.totalRow = tt.totalRow
			g.items = make([]*tview.TextView, tt.totalRow)
			g.nodes = make([]*spanTreeNode, tt.totalRow)
			for i := 0; i < tt.totalRow; i++ {
				g.items[i] = tview.NewTextView()
				g.nodes[i] = &spanTreeNode{span: &telemetry.SpanData{
					Span:         testdata.Spans[0],
					ResourceSpan: testdata.RSpans[0],
					ScopeSpans:   testdata.SSpans[0],
				}}
			}

			g.goToFirst(nil)

			assert.Equal(t, tt.wantCurrentRow, g.currentRow)
			row, col := g.gridView.GetOffset()
			assert.Equal(t, tt.wantOffsetRow, row)
			assert.Equal(t, tt.wantOffsetCol, col)
		})
	}
}

func TestGoToLast(t *testing.T) {
	store := telemetry.NewStore(clockwork.NewRealClock())
	payload, testdata := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{10}})
	store.AddSpan(&payload)

	tests := []struct {
		name           string
		initialRow     int
		totalRow       int
		wantCurrentRow int
		wantOffsetRow  int
		wantOffsetCol  int
	}{
		{
			name:           "From_First",
			initialRow:     0,
			totalRow:       10,
			wantCurrentRow: 9,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
		{
			name:           "From_Middle",
			initialRow:     5,
			totalRow:       10,
			wantCurrentRow: 9,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
		{
			name:           "Already_Last",
			initialRow:     9,
			totalRow:       10,
			wantCurrentRow: 9,
			wantOffsetRow:  0,
			wantOffsetCol:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newGrid(nil, store.GetTraceCache(), nil, nil, nil)
			g.currentRow = tt.initialRow
			g.totalRow = tt.totalRow
			g.items = make([]*tview.TextView, tt.totalRow)
			g.nodes = make([]*spanTreeNode, tt.totalRow)
			for i := 0; i < tt.totalRow; i++ {
				g.items[i] = tview.NewTextView()
				g.nodes[i] = &spanTreeNode{span: &telemetry.SpanData{
					Span:         testdata.Spans[0],
					ResourceSpan: testdata.RSpans[0],
					ScopeSpans:   testdata.SSpans[0],
				}}
			}

			g.goToLast(nil)

			assert.Equal(t, tt.wantCurrentRow, g.currentRow)
			row, col := g.gridView.GetOffset()
			assert.Equal(t, tt.wantOffsetRow, row)
			assert.Equal(t, tt.wantOffsetCol, col)
		})
	}
}
