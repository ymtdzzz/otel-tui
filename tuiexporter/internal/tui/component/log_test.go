package component

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

var noopShowModalFn layout.ShowModalFunc = func(p tview.Primitive, s string) *tview.TextView {
	return tview.NewTextView()
}

var noopHideModalFn layout.HideModalFunc = func(p tview.Primitive) {}

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
