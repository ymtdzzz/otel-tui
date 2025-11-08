package trace

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
)

type detail struct {
	commands      *tview.TextView
	view          *tview.Flex
	tree          *tview.TreeView
	resizeManager *layout.ResizeManager
}

func newDetail(
	commands *tview.TextView,
	resizeManager *layout.ResizeManager,
) *detail {
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Details (d)").SetBorder(true)

	detail := &detail{
		commands:      commands,
		view:          container,
		resizeManager: resizeManager,
	}

	detail.update(nil)

	return detail
}

func (d *detail) flush() {
	d.view.Clear()
	d.tree.SetRoot(nil)
}

func (d *detail) update(spans []*telemetry.SpanData) {
	hasFocus := d.view.HasFocus()
	d.view.Clear()
	d.tree = d.getTraceInfoTree(spans)
	d.updateCommands()
	d.view.AddItem(d.tree, 0, 1, true)
	if hasFocus {
		navigation.Focus(d.view)
	}
}

func (d *detail) getTraceInfoTree(spans []*telemetry.SpanData) *tview.TreeView {
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
	layout.AppendAttrsSorted(attrs, r.Attributes())
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
		layout.AppendAttrsSorted(attrs, ss.Scope().Attributes())
		scope.AddChild(attrs)

		scopes.AddChild(scope)
	}
	resource.AddChild(scopes)

	root.AddChild(resource)

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
	keyMaps.Merge(d.resizeManager.KeyMaps())
	layout.RegisterCommandList(d.commands, d.tree, nil, keyMaps)
}
