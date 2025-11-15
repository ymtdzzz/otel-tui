package layout

import "github.com/rivo/tview"

const (
	PageIDTraces        = "Traces"
	PageIDMetrics       = "Metrics"
	PageIDLogs          = "Logs"
	PageIDTraceTopology = "TraceTopology"
	PageIDTimeline      = "Timeline"
	PageIDModal         = "Modal"
)

func AttachTab(p tview.Primitive, name string) *tview.Flex {
	var text string
	switch name {
	case PageIDTraces:
		text = "< [yellow]Traces[white] | Metrics | Logs | Topology (beta) > (Tab to switch)"
	case PageIDMetrics:
		text = "< Traces | [yellow]Metrics[white] | Logs | Topology (beta) > (Tab to switch)"
	case PageIDLogs:
		text = "< Traces | Metrics | [yellow]Logs[white] | Topology (beta) > (Tab to switch)"
	case PageIDTraceTopology:
		text = "< Traces | Metrics | Logs | [yellow]Topology (beta)[white] > (Tab to switch)"
	}

	tabs := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(text)

	base := tview.NewFlex().SetDirection(tview.FlexRow)
	base.AddItem(tabs, 1, 1, false).
		AddItem(p, 0, 1, true)

	return base
}
