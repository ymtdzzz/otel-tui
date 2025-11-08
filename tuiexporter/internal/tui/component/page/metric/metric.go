package metric

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

const (
	defaultTableProportion = 25
	defaultSideProportion  = 25
	defaultDetaiProportion = 25
	defaultChartProportion = 25
)

type MetricPage struct {
	setFocusFn func(primitive tview.Primitive)
	view       *tview.Flex
	table      *table
	detail     *detail
	chart      *chart
}

func NewMetricPage(
	setFocusFn func(primitive tview.Primitive),
	showModalFn layout.ShowModalFunc,
	hideModalFn layout.HideModalFunc,
	store *telemetry.Store,
) *MetricPage {
	commands := layout.NewCommandList()
	container := tview.NewFlex().SetDirection(tview.FlexColumn)
	sideContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	resizeManager := layout.NewResizeManager(layout.ResizeDirectionHorizontal)
	sideResizeManager := layout.NewResizeManager(layout.ResizeDirectionVertical)
	detail := newDetail(commands, showModalFn, hideModalFn, []*layout.ResizeManager{
		sideResizeManager,
		resizeManager,
	})
	chart := newChart(commands, store, []*layout.ResizeManager{
		sideResizeManager,
		resizeManager,
	})
	table := newTable(commands, setFocusFn, store, detail, chart, []*layout.ResizeManager{resizeManager})

	resizeManager.Register(
		container,
		table.view,
		sideContainer,
		defaultTableProportion,
		defaultSideProportion,
		commands,
	)
	sideResizeManager.Register(
		sideContainer,
		detail.view,
		chart.view,
		defaultDetaiProportion,
		defaultChartProportion,
		commands,
	)

	sideContainer.AddItem(detail.view, 0, defaultDetaiProportion, false).
		AddItem(chart.view, 0, defaultChartProportion, false)
	container.AddItem(table.view, 0, defaultTableProportion, true).
		AddItem(sideContainer, 0, defaultSideProportion, false)

	metric := &MetricPage{
		setFocusFn: setFocusFn,
		view:       container,
		table:      table,
		detail:     detail,
		chart:      chart,
	}

	metric.view = layout.AttachTab(layout.AttachCommandList(commands, container), layout.PageIDMetrics)

	metric.registerCommands()
	store.RegisterOnFlushed(func() {
		metric.flush()
	})

	return metric
}

func (p *MetricPage) GetPrimitive() tview.Primitive {
	return p.view
}

func (p *MetricPage) flush() {
	p.detail.flush()
	p.chart.flush()
}

func (p *MetricPage) registerCommands() {
	p.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !p.table.filter.View().HasFocus() {
			switch event.Rune() {
			case 'd':
				p.setFocusFn(p.detail.view)
				return nil
			case 'm':
				p.setFocusFn(p.table.view)
				return nil
			case 'c':
				p.setFocusFn(p.chart.view)
				return nil
			}
		}

		return event
	})
}
