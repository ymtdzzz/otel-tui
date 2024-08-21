package component

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

var spanTableHeader = [4]string{
	"Trace ID",
	"Service Name",
	"Received At",
	"Span Name",
}

// SpanDataForTable is a wrapper for spans to be displayed in a table.
type SpanDataForTable struct {
	tview.TableContentReadOnly
	spans *telemetry.SvcSpans
}

// NewSpanDataForTable creates a new SpanDataForTable.
func NewSpanDataForTable(spans *telemetry.SvcSpans) SpanDataForTable {
	return SpanDataForTable{
		spans: spans,
	}
}

// implementations for tview Virtual Table
// see: https://github.com/rivo/tview/wiki/VirtualTable
func (s SpanDataForTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return getHeaderCell(spanTableHeader[:], column)
	}
	if row > 0 && row <= len(*s.spans) {
		return getCellFromSpan((*s.spans)[row-1], column)
	}
	return tview.NewTableCell("N/A")
}

func (s SpanDataForTable) GetRowCount() int {
	return len(*s.spans) + 1
}

func (s SpanDataForTable) GetColumnCount() int {
	return len(spanTableHeader)
}

func getHeaderCell(header []string, column int) *tview.TableCell {
	cell := tview.NewTableCell("N/A").
		SetSelectable(false).
		SetTextColor(tcell.ColorYellow)
	if column >= len(header) {
		return cell
	}
	cell.SetText(header[column])

	return cell
}

// getCellFromSpan returns a table cell for the given span and column.
func getCellFromSpan(span *telemetry.SpanData, column int) *tview.TableCell {
	text := "N/A"

	switch column {
	case 0:
		text = span.Span.TraceID().String()
	case 1:
		if serviceName, ok := span.ResourceSpan.Resource().Attributes().Get("service.name"); ok {
			text = serviceName.AsString()
		}
	case 2:
		text = span.ReceivedAt.Local().Format("2006-01-02 15:04:05")
	case 3:
		text = span.Span.Name()
	}

	return tview.NewTableCell(text)
}

func getTraceInfoTree(commands *tview.TextView, spans []*telemetry.SpanData) *tview.TreeView {
	if len(spans) == 0 {
		return tview.NewTreeView()
	}
	traceID := spans[0].Span.TraceID().String()
	sname, _ := spans[0].ResourceSpan.Resource().Attributes().Get("service.name")
	root := tview.NewTreeNode(fmt.Sprintf("%s (%s)", sname.AsString(), traceID))
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)

	// statistics
	statistics := tview.NewTreeNode("Statistics")
	spanCount := tview.NewTreeNode(fmt.Sprintf("span count: %d", len(spans)))
	statistics.AddChild(spanCount)

	root.AddChild(statistics)

	// resource info
	rs := spans[0].ResourceSpan
	r := rs.Resource()
	resource := tview.NewTreeNode("Resource")
	rdropped := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", r.DroppedAttributesCount()))
	resource.AddChild(rdropped)
	rschema := tview.NewTreeNode(fmt.Sprintf("schema url: %s", rs.SchemaUrl()))
	resource.AddChild(rschema)

	attrs := tview.NewTreeNode("Attributes")
	appendAttrsSorted(attrs, r.Attributes())
	resource.AddChild(attrs)

	// scope info
	scopes := tview.NewTreeNode("Scopes")
	for si := 0; si < rs.ScopeSpans().Len(); si++ {
		ss := rs.ScopeSpans().At(si)
		scope := tview.NewTreeNode(ss.Scope().Name())
		sschema := tview.NewTreeNode(fmt.Sprintf("schema url: %s", ss.SchemaUrl()))
		scope.AddChild(sschema)

		scope.AddChild(tview.NewTreeNode(fmt.Sprintf("version: %s", ss.Scope().Version())))
		scope.AddChild(tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", ss.Scope().DroppedAttributesCount())))

		attrs := tview.NewTreeNode("Attributes")
		appendAttrsSorted(attrs, ss.Scope().Attributes())
		scope.AddChild(attrs)

		scopes.AddChild(scope)
	}
	resource.AddChild(scopes)

	root.AddChild(resource)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
	})

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

func appendAttrsSorted(parent *tview.TreeNode, attrs pcommon.Map) {
	keys := make([]string, 0, attrs.Len())
	attrs.Range(func(k string, _ pcommon.Value) bool {
		keys = append(keys, k)
		return true
	})
	sort.Strings(keys)

	for _, k := range keys {
		v, _ := attrs.Get(k)
		attr := tview.NewTreeNode(fmt.Sprintf("%s: %s", k, v.AsString()))
		parent.AddChild(attr)
	}
}
