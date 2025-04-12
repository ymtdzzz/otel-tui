package component

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
)

func TestLogDataForTable(t *testing.T) {
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
	_, testdata1 := test.GenerateOTLPLogsPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	_, testdata2 := test.GenerateOTLPLogsPayload(t, 2, 1, []int{1}, [][]int{{1}})
	testdata1.Logs[0].Attributes().PutStr("event.name", "device.app.lifecycle")
	logs := &[]*telemetry.LogData{
		{
			Log:         testdata1.Logs[0],
			ResourceLog: testdata1.RLogs[0],
		},
		{
			Log:         testdata1.Logs[1],
			ResourceLog: testdata1.RLogs[0],
		},
		{
			Log:         testdata1.Logs[2],
			ResourceLog: testdata1.RLogs[0],
		},
		{
			Log:         testdata1.Logs[3],
			ResourceLog: testdata1.RLogs[0],
		},
		{
			Log:         testdata1.Logs[4],
			ResourceLog: testdata1.RLogs[0],
		},
		{
			Log:         testdata1.Logs[5],
			ResourceLog: testdata1.RLogs[0],
		},
		{
			Log:         testdata1.Logs[6],
			ResourceLog: testdata1.RLogs[1],
		},
		{
			Log:         testdata1.Logs[7],
			ResourceLog: testdata1.RLogs[1],
		},
		{
			Log:         testdata2.Logs[0],
			ResourceLog: testdata2.RLogs[0],
		},
		{
			Log:         testdata2.Logs[1],
			ResourceLog: testdata2.RLogs[0],
		},
	}
	ldftable := NewLogDataForTable(logs)
	ldftableForTL := NewLogDataForTableForTimeline(logs)

	t.Run("GetRowCount", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			assert.Equal(t, 11, ldftable.GetRowCount()) // including header row
		})
		t.Run("for timeline", func(t *testing.T) {
			assert.Equal(t, 11, ldftableForTL.GetRowCount()) // including header row
		})
	})

	t.Run("GetColumnCount", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			assert.Equal(t, 6, ldftable.GetColumnCount())
		})
		t.Run("for timeline", func(t *testing.T) {
			assert.Equal(t, 5, ldftableForTL.GetColumnCount())
		})
	})

	t.Run("GetCell", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			tests := []struct {
				name   string
				row    int
				column int
				want   string
			}{
				{
					name:   "invalid row",
					row:    10,
					column: 0,
					want:   "N/A",
				},
				{
					name:   "invalid column",
					row:    0,
					column: 6,
					want:   "N/A",
				},
				{
					name:   "trace ID trace 1 span-1-1-1",
					row:    0,
					column: 0,
					want:   "01000000000000000000000000000000",
				},
				{
					name:   "event name trace 1 span-1-1-1",
					row:    0,
					column: 4,
					want:   "device.app.lifecycle",
				},
				{
					name:   "service name trace 1 span-2-1-1",
					row:    6,
					column: 1,
					want:   "test-service-2",
				},
				{
					name:   "timestamp trace 1 span-2-1-1",
					row:    6,
					column: 2,
					want:   "2022-10-21 07:10:02",
				},
				{
					name:   "serverity trace 1 span-2-1-1",
					row:    6,
					column: 3,
					want:   "INFO",
				},
				{
					name:   "event name trace 2 span-1-1-1",
					row:    8,
					column: 4,
					want:   "N/A",
				},
				{
					name:   "raw data trace 2 span-1-1-1",
					row:    8,
					column: 5,
					want:   "log body 0-0-0-0",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					assert.Equal(t, tt.want, ldftable.GetCell(tt.row+1, tt.column).Text)
				})
			}

			t.Run("full datetime", func(t *testing.T) {
				ldftable.SetFullDatetime(true)
				defer ldftable.SetFullDatetime(false)
				assert.Equal(t, "2022-10-21 07:10:02.100000Z", ldftable.GetCell(1, 2).Text)
			})
		})
		t.Run("for header", func(t *testing.T) {
			tests := []struct {
				name   string
				row    int
				column int
				want   string
			}{
				{
					name:   "trace ID",
					row:    -1,
					column: 0,
					want:   "Trace ID",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					assert.Equal(t, tt.want, ldftable.GetCell(tt.row+1, tt.column).Text)
				})
			}
		})
		t.Run("for timeline", func(t *testing.T) {
			tests := []struct {
				name   string
				row    int
				column int
				want   string
			}{
				{
					name:   "invalid row",
					row:    10,
					column: 0,
					want:   "N/A",
				},
				{
					name:   "invalid column",
					row:    0,
					column: 5,
					want:   "N/A",
				},
				{
					name:   "event name trace 1 span-1-1-1",
					row:    0,
					column: 3,
					want:   "device.app.lifecycle",
				},
				{
					name:   "service name trace 1 span-2-1-1",
					row:    6,
					column: 0,
					want:   "test-service-2",
				},
				{
					name:   "timestamp trace 1 span-2-1-1",
					row:    6,
					column: 1,
					want:   "2022-10-21 07:10:02",
				},
				{
					name:   "serverity trace 1 span-2-1-1",
					row:    6,
					column: 2,
					want:   "INFO",
				},
				{
					name:   "event name trace 2 span-1-1-1",
					row:    8,
					column: 3,
					want:   "N/A",
				},
				{
					name:   "raw data trace 2 span-1-1-1",
					row:    8,
					column: 4,
					want:   "log body 0-0-0-0",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					assert.Equal(t, tt.want, ldftableForTL.GetCell(tt.row+1, tt.column).Text)
				})
			}

			t.Run("full datetime", func(t *testing.T) {
				ldftableForTL.SetFullDatetime(true)
				defer ldftableForTL.SetFullDatetime(false)
				assert.Equal(t, "2022-10-21 07:10:02.100000Z", ldftableForTL.GetCell(1, 1).Text)
			})
		})
	})

	t.Run("tableModalMapper GetColumnIdx", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			assert.Equal(t, "RawData", ldftable.mapper[ldftable.GetColumnIdx()].header)
		})
		t.Run("for timeline", func(t *testing.T) {
			assert.Equal(t, "RawData", ldftableForTL.mapper[ldftableForTL.GetColumnIdx()].header)
		})
	})
}

func TestGetLogInfoTree(t *testing.T) {
	// traceid: 1
	//  └- resource: test-service-1
	//    └- scope: test-scope-1-1
	//      └- span: span-1-1-1
	//        └- log: log-1-1-1-1
	//        └- log: log-1-1-1-2
	_, testdata := test.GenerateOTLPLogsPayload(t, 1, 2, []int{2, 1}, [][]int{{2, 1}, {1}})
	logs := []*telemetry.LogData{
		{
			Log:         testdata.Logs[0],
			ResourceLog: testdata.RLogs[0],
			ScopeLog:    testdata.SLogs[0],
		},
	}
	sw, sh := 55, 26
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	gottree := getLogInfoTree(nil, noopShowModalFn, noopHideModalFn, logs[0], nil, nil)
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

	want := `Log
└──Resource
   ├──dropped attributes count: 1
   ├──schema url:
   ├──Attributes
   │  ├──resource attribute: resource attribute value
   │  ├──resource index: 0
   │  └──service.name: test-service-1
   ├──Scopes
   │  └──test-scope-1-1
   │     ├──schema url:
   │     ├──version: v0.0.1
   │     ├──dropped attributes count: 2
   │     └──Attributes
   │        └──scope index: 0
   └──LogRecord
      ├──trace id: 01000000000000000000000000000000
      ├──span id: 0100000000000000
      ├──timestamp: 2022-10-21 07:10:02.100000Z
      ├──observed timestamp: 2022-10-21 07:10:02.200000
      ├──body: log body 0-0-0-0
      ├──severity: INFO (9)
      ├──flags: 0
      ├──dropped attributes count: 3
      └──Attributes
         └──span index: 0
`
	gotLines := strings.Split(got.String(), "\n")
	wantLines := strings.Split(want, "\n")

	assert.Equal(t, len(wantLines), len(gotLines))

	for i := 0; i < len(wantLines); i++ {
		assert.Equal(t, strings.TrimRight(wantLines[i], " \t\r"), strings.TrimRight(gotLines[i], " \t\r"))
	}
}
