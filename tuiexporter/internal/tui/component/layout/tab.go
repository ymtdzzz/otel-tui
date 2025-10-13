package layout

import "github.com/rivo/tview"

const (
	PAGE_TRACES         = "Traces"
	PAGE_METRICS        = "Metrics"
	PAGE_LOGS           = "Logs"
	PAGE_TRACE_TOPOLOGY = "TraceTopology"
)

func AttachTab(p tview.Primitive, name string) *tview.Flex {
	var text string
	switch name {
	case PAGE_TRACES:
		text = "< [yellow]Traces[white] | Metrics | Logs | Topology (beta) > (Tab to switch)"
	case PAGE_METRICS:
		text = "< Traces | [yellow]Metrics[white] | Logs | Topology (beta) > (Tab to switch)"
	case PAGE_LOGS:
		text = "< Traces | Metrics | [yellow]Logs[white] | Topology (beta) > (Tab to switch)"
	case PAGE_TRACE_TOPOLOGY:
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
