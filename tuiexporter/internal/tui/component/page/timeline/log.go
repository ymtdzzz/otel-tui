package timeline

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/table"
)

type logPane struct {
	commands    *tview.TextView
	tableView   *tview.Table
	showModalFn layout.ShowModalFunc
	hideModalFn layout.HideModalFunc
	lcache      *telemetry.LogCache
	logData     *table.LogDataForTable
	allLogs     bool
}

func newLogPane(
	commands *tview.TextView,
	showModalFn layout.ShowModalFunc,
	hideModalFn layout.HideModalFunc,
	lcache *telemetry.LogCache,
) *logPane {
	container := tview.NewTable().SetBorders(false).SetSelectable(true, false)

	return &logPane{
		commands:    commands,
		tableView:   container,
		showModalFn: showModalFn,
		hideModalFn: hideModalFn,
		lcache:      lcache,
		logData:     nil,
		allLogs:     false,
	}
}

func (l *logPane) updateLog(traceID, spanID string) {
	logCount := 0
	if lds, ok := l.lcache.GetLogsByTraceID(traceID); ok {
		if !l.allLogs && spanID != "" {
			flds := []*telemetry.LogData{}
			for _, ld := range lds {
				if ld.Log.SpanID().String() == spanID {
					flds = append(flds, ld)
				}
			}
			lds = flds
		}
		logCount = len(lds)
		log.Printf("Log count(%s): %d", traceID, logCount)
		logData := table.NewLogDataForTableForTimeline(&lds)
		if l.logData != nil {
			logData.SetFullDatetime(l.logData.IsFullDatetime())
		}
		l.logData = &logData
		l.tableView.SetContent(&logData)
		layout.AttachModalForTableRows(l.tableView, &logData, l.showModalFn, l.hideModalFn)
	}
	l.tableView.SetBorder(true).SetTitle(fmt.Sprintf("Logs (l) -- %d logs found (L: toggle collapse, A: toggle filter by span)", logCount))
	l.updateCommands()
}

func (l *logPane) toggleAllLogs(traceID string, currentSpan *telemetry.SpanData) {
	l.allLogs = !l.allLogs
	if currentSpan != nil {
		l.updateLog(traceID, currentSpan.Span.SpanID().String())
	}
}

func (l *logPane) updateCommands() {
	keyMaps := layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyCtrlF, ' ', tcell.ModNone),
			Description: "Toggle full datetime",
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				if l.logData != nil {
					l.logData.SetFullDatetime(!l.logData.IsFullDatetime())
				}
				return nil
			},
		},
	}
	layout.RegisterCommandList(l.commands, l.tableView, nil, keyMaps)
}
