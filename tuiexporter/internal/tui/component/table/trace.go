package table

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

var defaultSpanCellMappers = cellMappers[telemetry.SpanData]{
	1: {
		header: "Service Name",
		getTextRowFn: func(data *telemetry.SpanData) string {
			return data.GetServiceName()
		},
	},
	2: {
		header: "Latency",
		getTextRowFn: func(data *telemetry.SpanData) string {
			return data.GetDurationText()
		},
	},
	3: {
		header: "Received At",
		getTextRowFn: func(data *telemetry.SpanData) string {
			panic("Received At column should be overridden")
		},
	},
	4: {
		header: "Span Name",
		getTextRowFn: func(data *telemetry.SpanData) string {
			return data.GetSpanName()
		},
	},
}

// SpanDataForTable is a wrapper for spans to be displayed in a table.
type SpanDataForTable struct {
	tview.TableContentReadOnly
	tcache         *telemetry.TraceCache
	spans          *telemetry.SvcSpans
	sortType       *telemetry.SortType
	mapper         cellMappers[telemetry.SpanData]
	isFullDatetime bool
}

// NewSpanDataForTable creates a new SpanDataForTable.
func NewSpanDataForTable(tcache *telemetry.TraceCache, spans *telemetry.SvcSpans, sortType *telemetry.SortType) SpanDataForTable {
	t := SpanDataForTable{
		tcache:   tcache,
		spans:    spans,
		sortType: sortType,
		mapper:   defaultSpanCellMappers,
	}
	t.updateReceivedAtMapper()

	return t
}

// SetFullDatetime sets the full datetime flag for the table.
func (s *SpanDataForTable) SetFullDatetime(full bool) {
	s.isFullDatetime = full
	s.updateReceivedAtMapper()
}

// IsFullDatetime returns the full datetime flag for the table.
func (s SpanDataForTable) IsFullDatetime() bool {
	return s.isFullDatetime
}

func (s *SpanDataForTable) updateReceivedAtMapper() {
	for k, m := range s.mapper {
		if m.header == "Received At" {
			m.getTextRowFn = func(data *telemetry.SpanData) string {
				return data.GetReceivedAtText(s.isFullDatetime)
			}
			s.mapper[k] = m
			break
		}
	}
}

// implementations for tview Virtual Table
// see: https://github.com/rivo/tview/wiki/VirtualTable
func (s SpanDataForTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return s.getHeaderCell(column, s.sortType)
	}
	if row > 0 && row <= len(*s.spans) {
		sd := (*s.spans)[row-1]
		if column == 0 {
			return s.getErrorIndicator(sd)
		}
		return getCellFromData(s.mapper, sd, column)
	}
	return tview.NewTableCell("N/A")
}

func (s SpanDataForTable) GetRowCount() int {
	return len(*s.spans) + 1
}

func (s SpanDataForTable) GetColumnCount() int {
	return len(s.mapper) + 1 // including error indicator
}

func (s SpanDataForTable) getErrorIndicator(span *telemetry.SpanData) *tview.TableCell {
	if s.tcache == nil {
		return tview.NewTableCell("")
	}
	text := ""
	if sname, ok := span.ResourceSpan.Resource().Attributes().Get("service.name"); ok {
		if haserr, ok := s.tcache.HasErrorByTraceIDAndSvc(span.Span.TraceID().String(), sname.AsString()); ok && haserr {
			text = "[!]"
		}
	}
	return tview.NewTableCell(text)
}

func (s SpanDataForTable) getHeaderCell(column int, sortType *telemetry.SortType) *tview.TableCell {
	cell := tview.NewTableCell("N/A").
		SetSelectable(false).
		SetTextColor(tcell.ColorYellow)
	h, ok := s.mapper[column]
	if !ok {
		if column == 0 {
			cell.SetText(" ") // Error indicator
		}
		return cell
	}
	if !sortType.IsNone() && sortType.GetHeaderLabel() == h.header {
		if sortType.IsDesc() {
			cell.SetText(h.header + " ▼")
		} else {
			cell.SetText(h.header + " ▲")
		}
		return cell
	}
	cell.SetText(h.header)

	return cell
}
