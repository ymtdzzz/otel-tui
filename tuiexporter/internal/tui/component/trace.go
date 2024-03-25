package component

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

func GetTraceInfoTree(spans []*telemetry.SpanData) *tview.TreeView {
	if len(spans) == 0 {
		return nil
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
	for k, v := range r.Attributes().AsRaw() {
		attr := tview.NewTreeNode(fmt.Sprintf("%s: %s", k, v))
		attrs.AddChild(attr)
	}
	resource.AddChild(attrs)

	// scope info
	scopes := tview.NewTreeNode("Scopes")
	for si := 0; si < rs.ScopeSpans().Len(); si++ {
		ss := rs.ScopeSpans().At(si)
		scope := tview.NewTreeNode(fmt.Sprintf("Scope %d", si))
		sschema := tview.NewTreeNode(fmt.Sprintf("schema url: %s", ss.SchemaUrl()))
		scope.AddChild(sschema)

		isc := tview.NewTreeNode("Instrumentation Scope")
		isc.AddChild(tview.NewTreeNode(fmt.Sprintf("name: %s", ss.Scope().Name())))
		isc.AddChild(tview.NewTreeNode(fmt.Sprintf("version: %s", ss.Scope().Version())))
		isc.AddChild(tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", ss.Scope().DroppedAttributesCount())))

		aatrs := tview.NewTreeNode("Attributes")
		for k, v := range ss.Scope().Attributes().AsRaw() {
			attr := tview.NewTreeNode(fmt.Sprintf("%s: %s", k, v))
			aatrs.AddChild(attr)
		}
		isc.AddChild(aatrs)

		scopes.AddChild(scope)
	}
	resource.AddChild(scopes)

	root.AddChild(resource)

	return tree
}
