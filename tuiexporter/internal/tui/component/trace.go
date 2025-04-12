package component

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"go.opentelemetry.io/collector/pdata/pcommon"
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

func getTraceInfoTree(commands *tview.TextView, showModalFn showModalFunc, hideModalFn hideModalFunc, spans []*telemetry.SpanData) *tview.TreeView {
	if len(spans) == 0 {
		return tview.NewTreeView()
	}
	traceID := spans[0].Span.TraceID().String()
	sname := telemetry.GetServiceNameFromResource(spans[0].ResourceSpan.Resource())
	root := tview.NewTreeNode(fmt.Sprintf("%s (%s)", sname, traceID))
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
			description: "Toggle folding (parent), Show full text (child)",
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
