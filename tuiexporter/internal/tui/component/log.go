package component

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/datetime"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

func getLogInfoTree(commands *tview.TextView, showModalFn layout.ShowModalFunc, hideModalFn layout.HideModalFunc, l *telemetry.LogData, tcache *telemetry.TraceCache, drawTimelineFn func(traceID string)) *tview.TreeView {
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
	layout.AppendAttrsSorted(attrs, r.Attributes())
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
	layout.AppendAttrsSorted(sattrs, s.Attributes())
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
	layout.AppendAttrsSorted(lattrs, l.Log.Attributes())
	record.AddChild(lattrs)

	resource.AddChild(record)

	root.AddChild(resource)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
	})

	layout.AttachModalForTreeAttributes(tree, showModalFn, hideModalFn)

	layout.RegisterCommandList(commands, tree, nil, layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModCtrl),
			Description: "Reduce the width",
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModCtrl),
			Description: "Expand the width",
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			Description: "Toggle folding the child nodes",
		},
	})

	return tree
}
