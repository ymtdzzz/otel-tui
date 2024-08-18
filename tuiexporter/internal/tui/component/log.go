package component

import (
	"fmt"
	"log"

	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

// LogDataForTable is a wrapper for logs to be displayed in a table
type LogDataForTable struct {
	tview.TableContentReadOnly
	logs *[]*telemetry.LogData
}

// NewLogDataForTable creates a new LogDataForTable.
func NewLogDataForTable(logs *[]*telemetry.LogData) LogDataForTable {
	return LogDataForTable{
		logs: logs,
	}
}

// implementations for tview Virtual Table
// see: https://github.com/rivo/tview/wiki/VirtualTable
func (l LogDataForTable) GetCell(row, column int) *tview.TableCell {
	if row >= 0 && row < len(*l.logs) {
		return getCellFromLog((*l.logs)[row], column)
	}
	return tview.NewTableCell("N/A")
}

func (l LogDataForTable) GetRowCount() int {
	log.Printf("len(*l.logs): %d", len(*l.logs))
	return len(*l.logs)
}

func (l LogDataForTable) GetColumnCount() int {
	// 0: TraceID
	// 1: ServiceName
	// 2: Timestamp
	// 3: Severity
	// 4: EventName
	// 5: RawData
	return 6
}

// getCellFromLog returns a table cell for the given log and column.
func getCellFromLog(log *telemetry.LogData, column int) *tview.TableCell {
	text := "N/A"

	switch column {
	case 0:
		text = log.Log.TraceID().String()
	case 1:
		sname, _ := log.ResourceLog.Resource().Attributes().Get("service.name")
		text = sname.AsString()
	case 2:
		text = log.Log.Timestamp().AsTime().Format("2006/01/02 15:04:05")
	case 3:
		text = log.Log.SeverityText()
	case 4:
		// see: https://github.com/open-telemetry/semantic-conventions/blob/a4fc971e0c7ffa4b9572654f075d3cb8560db770/docs/general/events.md#event-definition
		if ename, ok := log.Log.Attributes().Get("event.name"); ok {
			text = ename.AsString()
		} else {
			text = "N/A<Event Name>"
		}
	case 5:
		text = log.Log.Body().AsString()
	}

	if text == "" {
		text = "N/A"
	}

	return tview.NewTableCell(text)
}

func getLogInfoTree(l *telemetry.LogData, tcache *telemetry.TraceCache, drawTimelineFn func(traceID string)) *tview.TreeView {
	if l == nil {
		return nil
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

	timestamp := l.Log.Timestamp().AsTime().Format("2006/01/02 15:04:05.000000")
	record.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", timestamp)))

	otimestamp := l.Log.ObservedTimestamp().AsTime().Format("2006/01/02 15:04:05.000000")
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

	return tree
}
