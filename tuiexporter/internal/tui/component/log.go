package component

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/datetime"
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

func getLogInfoTree(commands *tview.TextView, showModalFn showModalFunc, hideModalFn hideModalFunc, l *telemetry.LogData, tcache *telemetry.TraceCache, drawTimelineFn func(traceID string)) *tview.TreeView {
	if l == nil {
		return tview.NewTreeView()
	}
	root := tview.NewTreeNode("Log")
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)

	// resource info
	rl := l.ResourceLog
	r := rl.Resource()
	resource := tview.NewTreeNode("Resource")
	rdropped := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", r.DroppedAttributesCount()))
	resource.AddChild(rdropped)
	rschema := tview.NewTreeNode(fmt.Sprintf("schema url: %s", rl.SchemaUrl()))
	resource.AddChild(rschema)

	attrs := tview.NewTreeNode("Attributes")
	appendAttrsSorted(attrs, r.Attributes())
	resource.AddChild(attrs)

	// scope info
	scopes := tview.NewTreeNode("Scopes")
	sl := l.ScopeLog
	s := sl.Scope()
	scope := tview.NewTreeNode(s.Name())
	sschema := tview.NewTreeNode(fmt.Sprintf("schema url: %s", sl.SchemaUrl()))
	scope.AddChild(sschema)

	scope.AddChild(tview.NewTreeNode(fmt.Sprintf("version: %s", s.Version())))
	scope.AddChild(tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", s.DroppedAttributesCount())))

	sattrs := tview.NewTreeNode("Attributes")
	appendAttrsSorted(sattrs, s.Attributes())
	scope.AddChild(sattrs)

	scopes.AddChild(scope)
	resource.AddChild(scopes)

	// log body
	record := tview.NewTreeNode("LogRecord")

	traceID := l.Log.TraceID().String()
	traceNode := tview.NewTreeNode(fmt.Sprintf("trace id: %s", traceID))
	if tcache != nil {
		if _, ok := tcache.GetSpansByTraceID(traceID); ok {
			traceNode.SetText("(🔗)" + traceNode.GetText())
			traceNode.SetSelectable(true)
			traceNode.SetSelectedFunc(func() {
				drawTimelineFn(traceID)
			})
		}
	}
	record.AddChild(traceNode)

	spanID := l.Log.SpanID().String()
	spanNode := tview.NewTreeNode(fmt.Sprintf("span id: %s", spanID))
	record.AddChild(spanNode)

	timestamp := datetime.GetFullTime(l.Log.Timestamp().AsTime())
	record.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", timestamp)))

	otimestamp := datetime.GetFullTime(l.Log.ObservedTimestamp().AsTime())
	record.AddChild(tview.NewTreeNode(fmt.Sprintf("observed timestamp: %s", otimestamp)))

	body := tview.NewTreeNode(fmt.Sprintf("body: %s", l.Log.Body().AsString()))
	record.AddChild(body)

	severity := tview.NewTreeNode(fmt.Sprintf("severity: %s (%d)", l.Log.SeverityText(), l.Log.SeverityNumber()))
	record.AddChild(severity)

	flags := tview.NewTreeNode(fmt.Sprintf("flags: %d", l.Log.Flags()))
	record.AddChild(flags)

	ldropped := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", l.Log.DroppedAttributesCount()))
	record.AddChild(ldropped)

	lattrs := tview.NewTreeNode("Attributes")
	appendAttrsSorted(lattrs, l.Log.Attributes())
	record.AddChild(lattrs)

	resource.AddChild(record)

	root.AddChild(resource)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
	})

	attachModalForTreeAttributes(tree, showModalFn, hideModalFn)

	registerCommandList(commands, tree, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModCtrl),
			description: "Reduce the width",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModCtrl),
			description: "Expand the width",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			description: "Toggle folding the child nodes",
		},
	})

	return tree
}
