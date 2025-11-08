package timeline

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/datetime"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type detail struct {
	commands      *tview.TextView
	view          *tview.Flex
	tree          *tview.TreeView
	showModalFn   layout.ShowModalFunc
	hideModalFn   layout.HideModalFunc
	resizeManager *layout.ResizeManager
}

func newDetail(
	commands *tview.TextView,
	showModalFn layout.ShowModalFunc,
	hideModalFn layout.HideModalFunc,
	resizeManager *layout.ResizeManager,
) *detail {
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Details (d)").SetBorder(true)

	detail := &detail{
		commands:      commands,
		view:          container,
		showModalFn:   showModalFn,
		hideModalFn:   hideModalFn,
		resizeManager: resizeManager,
	}

	detail.update(nil)

	return detail
}

func (d *detail) update(span *telemetry.SpanData) {
	d.view.Clear()
	d.tree = d.getSpanInfoTree(span)
	d.updateCommands()
	d.view.AddItem(d.tree, 0, 1, true)
}

func (d *detail) getSpanInfoTree(span *telemetry.SpanData) *tview.TreeView {
	if span == nil {
		return tview.NewTreeView()
	}
	traceID := span.Span.TraceID().String()
	sname := telemetry.GetServiceNameFromResource(span.ResourceSpan.Resource())
	root := tview.NewTreeNode(fmt.Sprintf("%s (%s)", sname, traceID))
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)

	spanID := span.Span.SpanID().String()
	spanNode := tview.NewTreeNode(fmt.Sprintf("span id: %s", spanID))
	root.AddChild(spanNode)

	parentSpanID := span.Span.ParentSpanID().String()
	parentSpanNode := tview.NewTreeNode(fmt.Sprintf("parent span id: %s", parentSpanID))
	root.AddChild(parentSpanNode)

	state := span.Span.TraceState().AsRaw()
	stateNode := tview.NewTreeNode(fmt.Sprintf("trace state: %s", state))
	root.AddChild(stateNode)

	status := tview.NewTreeNode("Status")
	smessage := span.Span.Status().Message()
	smessageNode := tview.NewTreeNode(fmt.Sprintf("message: %s", smessage))
	status.AddChild(smessageNode)
	scode := span.Span.Status().Code()
	scodeText := ""
	if scode == ptrace.StatusCodeError {
		scodeText = fmt.Sprintf("code: %s ⚠️", scode)
	} else {
		scodeText = fmt.Sprintf("code: %s", scode)
	}
	scodeNode := tview.NewTreeNode(scodeText)
	status.AddChild(scodeNode)
	root.AddChild(status)

	flags := span.Span.Flags()
	flagsNode := tview.NewTreeNode(fmt.Sprintf("flags: %d", flags))
	root.AddChild(flagsNode)

	name := span.Span.Name()
	nameNode := tview.NewTreeNode(fmt.Sprintf("name: %s", name))
	root.AddChild(nameNode)

	kind := span.Span.Kind()
	kindNode := tview.NewTreeNode(fmt.Sprintf("kind: %s", kind))
	root.AddChild(kindNode)

	duration := span.Span.EndTimestamp().AsTime().Sub(span.Span.StartTimestamp().AsTime())
	durationNode := tview.NewTreeNode(fmt.Sprintf("duration: %s", duration.String()))
	root.AddChild(durationNode)

	startTime := datetime.GetFullTime(span.Span.StartTimestamp().AsTime())
	startTimeNode := tview.NewTreeNode(fmt.Sprintf("start time: %s", startTime))
	root.AddChild(startTimeNode)

	endTime := datetime.GetFullTime(span.Span.EndTimestamp().AsTime())
	endTimeNode := tview.NewTreeNode(fmt.Sprintf("end time: %s", endTime))
	root.AddChild(endTimeNode)

	dropped := span.ResourceSpan.Resource().DroppedAttributesCount()
	droppedNode := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", dropped))
	root.AddChild(droppedNode)

	attrs := tview.NewTreeNode("Attributes")
	layout.AppendAttrsSorted(attrs, span.Span.Attributes())
	root.AddChild(attrs)

	// events
	events := tview.NewTreeNode("Events")
	for ei := 0; ei < span.Span.Events().Len(); ei++ {
		event := span.Span.Events().At(ei)
		name := event.Name()
		eventNode := tview.NewTreeNode(name)

		timestamp := datetime.GetFullTime(event.Timestamp().AsTime())
		timestampNode := tview.NewTreeNode(fmt.Sprintf("timestamp: %s", timestamp))
		eventNode.AddChild(timestampNode)

		dropped := event.DroppedAttributesCount()
		droppedNode := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", dropped))
		eventNode.AddChild(droppedNode)

		attrs := tview.NewTreeNode("Attributes")
		layout.AppendAttrsSorted(attrs, event.Attributes())
		eventNode.AddChild(attrs)

		events.AddChild(eventNode)
	}
	root.AddChild(events)

	// links
	links := tview.NewTreeNode("Links")
	for li := 0; li < span.Span.Links().Len(); li++ {
		link := span.Span.Links().At(li)
		linkNode := tview.NewTreeNode(fmt.Sprintf("link %d", li))

		linkTraceID := link.TraceID().String()
		linkTraceIDNode := tview.NewTreeNode(fmt.Sprintf("trace id: %s", linkTraceID))
		linkNode.AddChild(linkTraceIDNode)

		linkSpanID := link.SpanID().String()
		linkSpanIDNode := tview.NewTreeNode(fmt.Sprintf("span id: %s", linkSpanID))
		linkNode.AddChild(linkSpanIDNode)

		linkFlags := link.Flags()
		linkFlagsNode := tview.NewTreeNode(fmt.Sprintf("flags: %d", linkFlags))
		linkNode.AddChild(linkFlagsNode)

		linkState := link.TraceState().AsRaw()
		linkStateNode := tview.NewTreeNode(fmt.Sprintf("trace state: %s", linkState))
		linkNode.AddChild(linkStateNode)

		linkDropped := link.DroppedAttributesCount()
		linkDroppedNode := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", linkDropped))
		linkNode.AddChild(linkDroppedNode)

		attrs := tview.NewTreeNode("Attributes")
		layout.AppendAttrsSorted(attrs, link.Attributes())
		linkNode.AddChild(attrs)

		links.AddChild(linkNode)
	}
	root.AddChild(links)

	// resource info
	rs := span.ResourceSpan
	r := rs.Resource()
	resource := tview.NewTreeNode("Resource")
	root.AddChild(resource)
	rdropped := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", r.DroppedAttributesCount()))
	resource.AddChild(rdropped)
	rschema := tview.NewTreeNode(fmt.Sprintf("schema url: %s", rs.SchemaUrl()))
	resource.AddChild(rschema)

	rattrs := tview.NewTreeNode("Attributes")
	layout.AppendAttrsSorted(rattrs, r.Attributes())
	resource.AddChild(rattrs)

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

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
	})

	layout.AttachModalForTreeAttributes(tree, d.showModalFn, d.hideModalFn)

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
