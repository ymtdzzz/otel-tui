package log

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/datetime"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
)

type detail struct {
	commands       *tview.TextView
	view           *tview.Flex
	tree           *tview.TreeView
	drawTimelineFn func(traceID string)
	resizeManagers []*layout.ResizeManager
	tcache         *telemetry.TraceCache
}

func newDetail(
	commands *tview.TextView,
	drawTimelineFn func(traceID string),
	resizeManagers []*layout.ResizeManager,
	tcache *telemetry.TraceCache,
) *detail {
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Details (d)").SetBorder(true)

	detail := &detail{
		commands:       commands,
		view:           container,
		drawTimelineFn: drawTimelineFn,
		resizeManagers: resizeManagers,
		tcache:         tcache,
	}

	detail.update(nil)

	return detail
}

func (d *detail) flush() {
	d.view.Clear()
	d.tree.SetRoot(nil)
}

func (d *detail) update(l *telemetry.LogData) {
	hasFocus := d.view.HasFocus()
	d.view.Clear()
	d.tree = d.getLogInfoTree(l)
	d.updateCommands()
	d.view.AddItem(d.tree, 0, 1, true)
	if hasFocus {
		navigation.Focus(d.view)
	}
}

func (d *detail) getLogInfoTree(l *telemetry.LogData) *tview.TreeView {
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
	if d.tcache != nil {
		if _, ok := d.tcache.GetSpansByTraceID(traceID); ok {
			traceNode.SetText("(🔗)" + traceNode.GetText())
			traceNode.SetSelectable(true)
			traceNode.SetSelectedFunc(func() {
				d.drawTimelineFn(traceID)
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

	layout.AttachModalForTreeAttributes(tree)

	return tree
}

func (d *detail) updateCommands() {
	keyMaps := layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			Description: "Toggle folding (parent), Show full text (child)",
		},
	}
	for _, rm := range d.resizeManagers {
		keyMaps.Merge(rm.KeyMaps())
	}
	layout.RegisterCommandList(d.commands, d.tree, nil, keyMaps)
}
