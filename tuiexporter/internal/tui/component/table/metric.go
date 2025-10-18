package table

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

var defaultMetricCellMappers = cellMappers[telemetry.MetricData]{
	0: {
		header: "Service Name",
		getTextRowFn: func(data *telemetry.MetricData) string {
			return data.GetServiceName()
		},
	},
	1: {
		header: "Metric Name",
		getTextRowFn: func(data *telemetry.MetricData) string {
			return data.GetMetricName()
		},
	},
	2: {
		header: "Metric Type",
		getTextRowFn: func(data *telemetry.MetricData) string {
			return data.GetMetricTypeText()
		},
	},
	3: {
		header: "Data Point Count",
		getTextRowFn: func(data *telemetry.MetricData) string {
			return data.GetDataPointNum()
		},
	},
}

type MetricDataForTable struct {
	tview.TableContentReadOnly
	metrics *[]*telemetry.MetricData
	mapper  cellMappers[telemetry.MetricData]
}

func NewMetricDataForTable(metrics *[]*telemetry.MetricData) MetricDataForTable {
	return MetricDataForTable{
		metrics: metrics,
		mapper:  defaultMetricCellMappers,
	}
}

// implementations for tview Virtual Table
// see: https://github.com/rivo/tview/wiki/VirtualTable
func (m MetricDataForTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return m.getHeaderCell(column)
	}
	if row > 0 && row <= len(*m.metrics) {
		return getCellFromData(m.mapper, (*m.metrics)[row-1], column)
	}
	return tview.NewTableCell("N/A")
}

func (m MetricDataForTable) GetRowCount() int {
	return len(*m.metrics) + 1
}

func (m MetricDataForTable) GetColumnCount() int {
	return len(m.mapper)
}

func (m MetricDataForTable) getHeaderCell(column int) *tview.TableCell {
	cell := tview.NewTableCell("N/A").
		SetSelectable(false).
		SetTextColor(tcell.ColorYellow)
	h, ok := m.mapper[column]
	if !ok {
		return cell
	}
	cell.SetText(h.header)

	return cell
}
