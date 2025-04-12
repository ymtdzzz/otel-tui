package component

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/datetime"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

var noopShowModalFn showModalFunc = func(p tview.Primitive, s string) *tview.TextView {
	return tview.NewTextView()
}

var noopHideModalFn hideModalFunc = func(p tview.Primitive) {}

func TestSpanDataForTable(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1 (code: Error)
	//  | | └- span: span-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- span: span-2-1-1 (code: OK)
	// traceid: 2
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-1 (code: Unset)
	_, testdata1 := test.GenerateOTLPTracesPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	_, testdata2 := test.GenerateOTLPTracesPayload(t, 2, 1, []int{1}, [][]int{{1}})
	receivedAt := time.Date(2024, 3, 30, 12, 30, 15, 0, time.UTC)
	testdata1.Spans[0].Status().SetCode(ptrace.StatusCodeError)
	testdata1.Spans[3].Status().SetCode(ptrace.StatusCodeOk)
	testdata2.Spans[0].Status().SetCode(ptrace.StatusCodeUnset)
	svc1sds := []*telemetry.SpanData{
		{
			Span:         testdata1.Spans[0],
			ResourceSpan: testdata1.RSpans[0],
			ReceivedAt:   receivedAt,
		}, // trace 1, span-1-1-1
		{
			Span:         testdata1.Spans[3],
			ResourceSpan: testdata1.RSpans[1],
			ReceivedAt:   receivedAt,
		}, // trace 1, span-2-1-1
	}
	svc2sds := []*telemetry.SpanData{
		{
			Span:         testdata2.Spans[0],
			ResourceSpan: testdata2.RSpans[0],
			ReceivedAt:   receivedAt,
		}, // trace 2, span-1-1-1
	}
	svcspans := &telemetry.SvcSpans{
		svc1sds[0],
		svc1sds[1],
		svc2sds[0],
	}
	tcache := telemetry.NewTraceCache()
	for _, sd := range svc1sds {
		tcache.UpdateCache("test-service-1", sd)
	}
	for _, sd := range svc2sds {
		tcache.UpdateCache("test-service-2", sd)
	}
	sortType := telemetry.SORT_TYPE_NONE
	sdftable := NewSpanDataForTable(tcache, svcspans, &sortType)

	t.Run("GetRowCount", func(t *testing.T) {
		assert.Equal(t, 4, sdftable.GetRowCount()) // including header row
	})

	t.Run("GetColumnCount", func(t *testing.T) {
		assert.Equal(t, 5, sdftable.GetColumnCount())
	})

	t.Run("GetCell_Header", func(t *testing.T) {
		tests := []struct {
			name     string
			sortType telemetry.SortType
			column   int
			want     string
		}{
			{
				name:     "N/A",
				sortType: telemetry.SORT_TYPE_NONE,
				column:   5,
				want:     "N/A",
			},
			{
				name:     "Latency None",
				sortType: telemetry.SORT_TYPE_NONE,
				column:   2,
				want:     "Latency",
			},
			{
				name:     "Latency Desc",
				sortType: telemetry.SORT_TYPE_LATENCY_DESC,
				column:   2,
				want:     "Latency ▼",
			},
			{
				name:     "Latency Asc",
				sortType: telemetry.SORT_TYPE_LATENCY_ASC,
				column:   2,
				want:     "Latency ▲",
			},
			{
				name:     "Service Name no effect",
				sortType: telemetry.SORT_TYPE_LATENCY_DESC,
				column:   1,
				want:     "Service Name",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				sortType = tt.sortType
				assert.Equal(t, tt.want, sdftable.GetCell(0, tt.column).Text)
			})
		}
	})

	t.Run("GetCell_Body", func(t *testing.T) {
		tests := []struct {
			name   string
			row    int
			column int
			want   string
		}{
			{
				name:   "invalid row",
				row:    3,
				column: 1,
				want:   "N/A",
			},
			{
				name:   "invalid column",
				row:    0,
				column: 5,
				want:   "N/A",
			},
			{
				name:   "has error trace 1 span-1-1-1",
				row:    0,
				column: 0,
				want:   "[!]",
			},
			{
				name:   "has no errors (OK) trace 1 span-2-1-1",
				row:    1,
				column: 0,
				want:   "",
			},
			{
				name:   "has no errors (Unset) trace 2 span-1-1-1",
				row:    2,
				column: 0,
				want:   "",
			},
			{
				name:   "service name trace 1 span-2-1-1",
				row:    1,
				column: 1,
				want:   "test-service-2",
			},
			{
				name:   "latency span-1-1-1",
				row:    0,
				column: 2,
				want:   "200ms",
			},
			{
				name:   "received at trace 2 span-1-1-1",
				row:    2,
				column: 3,
				want:   datetime.GetSimpleTime(receivedAt.Local()),
			},
			{
				name:   "span name trace 2 span-1-1-1",
				row:    2,
				column: 4,
				want:   "span-0-0-0",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.want, sdftable.GetCell(tt.row+1, tt.column).Text)
			})
		}

		t.Run("full datetime", func(t *testing.T) {
			sdftable.SetFullDatetime(true)
			defer sdftable.SetFullDatetime(false)
			assert.Equal(t, datetime.GetFullTime(receivedAt.Local()), sdftable.GetCell(3, 3).Text)
		})
	})
}

func TestGetTraceInfoTreeWithServiceName(t *testing.T) {
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
	_, testdata := test.GenerateOTLPTracesPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	spans := []*telemetry.SpanData{}
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[0],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[0],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[1],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[0],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[2],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[1],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[3],
		ResourceSpan: testdata.RSpans[1],
		ScopeSpans:   testdata.SSpans[2],
	})
	sw, sh := 55, 23
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	gottree := getTraceInfoTree(nil, noopShowModalFn, noopHideModalFn, spans)
	gottree.SetRect(0, 0, sw, sh)
	gottree.Draw(screen)
	screen.Sync()

	contents, w, _ := screen.GetContents()
	var got bytes.Buffer
	for n, v := range contents {
		var err error
		if n%w == w-1 {
			_, err = fmt.Fprintf(&got, "%c\n", v.Runes[0])
		} else {
			_, err = fmt.Fprintf(&got, "%c", v.Runes[0])
		}
		if err != nil {
			t.Error(err)
		}
	}

	want := `test-service-1 (01000000000000000000000000000000)
├──Statistics
│  └──span count: 4
└──Resource
   ├──dropped attributes count: 1
   ├──schema url:
   ├──Attributes
   │  ├──resource attribute: resource attribute value
   │  ├──resource index: 0
   │  └──service.name: test-service-1
   └──Scopes
      ├──test-scope-1-1
      │  ├──schema url:
      │  ├──version: v0.0.1
      │  ├──dropped attributes count: 2
      │  └──Attributes
      │     └──scope index: 0
      └──test-scope-1-2
         ├──schema url:
         ├──version: v0.0.1
         ├──dropped attributes count: 2
         └──Attributes
            └──scope index: 1
`
	gotLines := strings.Split(got.String(), "\n")
	wantLines := strings.Split(want, "\n")

	assert.Equal(t, len(wantLines), len(gotLines))

	for i := 0; i < len(wantLines); i++ {
		assert.Equal(t, strings.TrimRight(wantLines[i], " \t\r"), strings.TrimRight(gotLines[i], " \t\r"))
	}
}

func TestGetTraceInfoTreeWithoutServiceName(t *testing.T) {
	// traceid: 1
	//  └- resource: [Empty]
	//  | └- scope: test-scope-1-1
	//  | | └- span: span-1-1-1
	//  | | └- span: span-1-1-2
	//  | └- scope: test-scope-1-2
	//  |   └- span: span-1-2-3
	//  └- resource: test-service-2
	//    └- scope: test-scope-2-1
	//      └- span: span-2-1-1
	_, testdata := test.GenerateOTLPTracesPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	testdata.RSpans[0].Resource().Attributes().Remove("service.name")
	spans := []*telemetry.SpanData{}
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[0],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[0],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[1],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[0],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[2],
		ResourceSpan: testdata.RSpans[0],
		ScopeSpans:   testdata.SSpans[1],
	})
	spans = append(spans, &telemetry.SpanData{
		Span:         testdata.Spans[3],
		ResourceSpan: testdata.RSpans[1],
		ScopeSpans:   testdata.SSpans[2],
	})
	sw, sh := 55, 22
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	gottree := getTraceInfoTree(nil, noopShowModalFn, noopHideModalFn, spans)
	gottree.SetRect(0, 0, sw, sh)
	gottree.Draw(screen)
	screen.Sync()

	contents, w, _ := screen.GetContents()
	var got bytes.Buffer
	for n, v := range contents {
		var err error
		if n%w == w-1 {
			_, err = fmt.Fprintf(&got, "%c\n", v.Runes[0])
		} else {
			_, err = fmt.Fprintf(&got, "%c", v.Runes[0])
		}
		if err != nil {
			t.Error(err)
		}
	}

	want := `unknown (01000000000000000000000000000000)
├──Statistics
│  └──span count: 4
└──Resource
   ├──dropped attributes count: 1
   ├──schema url:
   ├──Attributes
   │  ├──resource attribute: resource attribute value
   │  └──resource index: 0
   └──Scopes
      ├──test-scope-1-1
      │  ├──schema url:
      │  ├──version: v0.0.1
      │  ├──dropped attributes count: 2
      │  └──Attributes
      │     └──scope index: 0
      └──test-scope-1-2
         ├──schema url:
         ├──version: v0.0.1
         ├──dropped attributes count: 2
         └──Attributes
            └──scope index: 1
`
	gotLines := strings.Split(got.String(), "\n")
	wantLines := strings.Split(want, "\n")

	assert.Equal(t, len(wantLines), len(gotLines))

	for i := 0; i < len(wantLines); i++ {
		assert.Equal(t, strings.TrimRight(wantLines[i], " \t\r"), strings.TrimRight(gotLines[i], " \t\r"))
	}
}

func TestGetTraceInfoTreeNoSpans(t *testing.T) {
	assert.Nil(t, getTraceInfoTree(nil, noopShowModalFn, noopHideModalFn, nil).GetRoot())
}
