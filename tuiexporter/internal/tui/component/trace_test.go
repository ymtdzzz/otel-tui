package component

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
)

func TestSpanDataForTable(t *testing.T) {
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
	_, testdata1 := test.GenerateOTLPTracesPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	_, testdata2 := test.GenerateOTLPTracesPayload(t, 2, 1, []int{1}, [][]int{{1}})
	receivedAt := time.Date(2024, 3, 30, 12, 30, 15, 0, time.UTC)
	svcspans := &telemetry.SvcSpans{
		&telemetry.SpanData{
			Span:         testdata1.Spans[0],
			ResourceSpan: testdata1.RSpans[0],
			ReceivedAt:   receivedAt,
		}, // trace 1, span-1-1-1
		&telemetry.SpanData{
			Span:         testdata1.Spans[3],
			ResourceSpan: testdata1.RSpans[1],
			ReceivedAt:   receivedAt,
		}, // trace 1, span-2-1-1
		&telemetry.SpanData{
			Span:         testdata2.Spans[0],
			ResourceSpan: testdata2.RSpans[0],
			ReceivedAt:   receivedAt,
		}, // trace 2, span-1-1-1
	}
	sdftable := NewSpanDataForTable(svcspans)

	t.Run("GetRowCount", func(t *testing.T) {
		assert.Equal(t, 4, sdftable.GetRowCount()) // including header row
	})

	t.Run("GetColumnCount", func(t *testing.T) {
		assert.Equal(t, 4, sdftable.GetColumnCount())
	})

	t.Run("GetCell", func(t *testing.T) {
		tests := []struct {
			name   string
			row    int
			column int
			want   string
		}{
			{
				name:   "invalid row",
				row:    3,
				column: 0,
				want:   "N/A",
			},
			{
				name:   "invalid column",
				row:    0,
				column: 4,
				want:   "N/A",
			},
			{
				name:   "trace ID trace 1 span-1-1-1",
				row:    0,
				column: 0,
				want:   "01000000000000000000000000000000",
			},
			{
				name:   "service name trace 1 span-2-1-1",
				row:    1,
				column: 1,
				want:   "test-service-2",
			},
			{
				name:   "received at trace 2 span-1-1-1",
				row:    2,
				column: 2,
				want:   receivedAt.Local().Format("2006-01-02 15:04:05"),
			},
			{
				name:   "span name trace 2 span-1-1-1",
				row:    2,
				column: 3,
				want:   "span-0-0-0",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.want, sdftable.GetCell(tt.row+1, tt.column).Text)
			})
		}
	})
}

func TestGetTraceInfoTree(t *testing.T) {
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
	sw, sh := 55, 24
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	gottree := getTraceInfoTree(context.Background(), nil, spans, nil)
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
├──Root Span                                           
│  ├──[ Searching... ]                                 
│  ├──[ Searching... ]                                 
│  └──[ Searching... ]                                 
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
         └──version: v0.0.1                            
`
	assert.Equal(t, want, got.String())
}

func TestGetTraceInfoTreeNoSpans(t *testing.T) {
	assert.Nil(t, getTraceInfoTree(context.Background(), nil, nil, nil).GetRoot())
}
