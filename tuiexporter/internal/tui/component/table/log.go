package table

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

var defaultLogCellMappers = cellMappers[telemetry.LogData]{
	0: {
		header: "Trace ID",
		getTextRowFn: func(log *telemetry.LogData) string {
			return log.GetTraceID()
		},
	},
	1: {
		header: "Service Name",
		getTextRowFn: func(log *telemetry.LogData) string {
			return log.GetServiceName()
		},
	},
	2: {
		header: "Timestamp",
		getTextRowFn: func(log *telemetry.LogData) string {
			panic("Timestamp column should be overridden")
		},
	},
	3: {
		header: "Severity",
		getTextRowFn: func(log *telemetry.LogData) string {
			return log.GetSeverity()
		},
	},
	4: {
		header: "Event Name",
		getTextRowFn: func(log *telemetry.LogData) string {
			return log.GetEventName()
		},
	},
	5: {
		header: "RawData",
		getTextRowFn: func(log *telemetry.LogData) string {
			return log.GetRawData()
		},
	},
}

var logCellMappersForTimeline = cellMappers[telemetry.LogData]{
	0: {
		header: "Service Name",
		getTextRowFn: func(log *telemetry.LogData) string {
			return log.GetServiceName()
		},
	},
	1: {
		header: "Timestamp",
		getTextRowFn: func(log *telemetry.LogData) string {
			panic("Timestamp column should be overridden")
		},
	},
	2: {
		header: "Severity",
		getTextRowFn: func(log *telemetry.LogData) string {
			return log.GetSeverity()
		},
	},
	3: {
		header: "Event Name",
		getTextRowFn: func(log *telemetry.LogData) string {
			return log.GetEventName()
		},
	},
	4: {
		header: "RawData",
		getTextRowFn: func(log *telemetry.LogData) string {
			return log.GetRawData()
		},
	},
}

// LogDataForTable is a wrapper for logs to be displayed in a table
type LogDataForTable struct {
	tview.TableContentReadOnly
	logs           *[]*telemetry.LogData
	mapper         cellMappers[telemetry.LogData]
	isFullDatetime bool
}

// NewLogDataForTable creates a new LogDataForTable.
func NewLogDataForTable(logs *[]*telemetry.LogData) LogDataForTable {
	l := LogDataForTable{
		logs:   logs,
		mapper: defaultLogCellMappers,
	}
	l.updateTimestampMapper()

	return l
}

// NewLogDataForTableForTimeline creates a new LogDataForTable for timeline page.
func NewLogDataForTableForTimeline(logs *[]*telemetry.LogData) LogDataForTable {
	l := LogDataForTable{
		logs:   logs,
		mapper: logCellMappersForTimeline,
	}
	l.updateTimestampMapper()

	return l
}

// SetFullDatetime sets whether to display full datetime or not
func (l *LogDataForTable) SetFullDatetime(full bool) {
	l.isFullDatetime = full
	l.updateTimestampMapper()
}

// IsFullDatetime returns whether to display full datetime or not
func (l LogDataForTable) IsFullDatetime() bool {
	return l.isFullDatetime
}

func (l *LogDataForTable) updateTimestampMapper() {
	for k, m := range l.mapper {
		if m.header == "Timestamp" {
			m.getTextRowFn = func(data *telemetry.LogData) string {
				return data.GetTimestampText(l.isFullDatetime)
			}
			l.mapper[k] = m
			break
		}
	}
}

// implementation for tableModalMapper interface
func (l *LogDataForTable) GetColumnIdx() int {
	return len(l.mapper) - 1
}

// implementations for tview Virtual Table
// see: https://github.com/rivo/tview/wiki/VirtualTable
func (l LogDataForTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return l.getHeaderCell(column)
	}
	if row > 0 && row <= len(*l.logs) {
		return getCellFromData(l.mapper, (*l.logs)[row-1], column)
	}
	return tview.NewTableCell("N/A")
}

func (l LogDataForTable) GetRowCount() int {
	return len(*l.logs) + 1
}

func (l LogDataForTable) GetColumnCount() int {
	return len(l.mapper)
}

func (l LogDataForTable) getHeaderCell(column int) *tview.TableCell {
	cell := tview.NewTableCell("N/A").
		SetSelectable(false).
		SetTextColor(tcell.ColorYellow)
	h, ok := l.mapper[column]
	if !ok {
		return cell
	}
	cell.SetText(h.header)

	return cell
}
