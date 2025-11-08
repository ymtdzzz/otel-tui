package metric

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
)

const (
	defaultTableProportion = 25
	defaultSideProportion  = 25
	defaultDetaiProportion = 25
	defaultChartProportion = 25
)

type MetricPage struct {
	view   *tview.Flex
	table  *table
	detail *detail
	chart  *chart
}

func NewMetricPage(
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
	table := newTable(commands, store, detail, chart, []*layout.ResizeManager{resizeManager})

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
		view:   container,
		table:  table,
		detail: detail,
		chart:  chart,
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
				navigation.Focus(p.detail.view)
				return nil
			case 'm':
				navigation.Focus(p.table.view)
				return nil
			case 'c':
				navigation.Focus(p.chart.view)
				return nil
			}
		}

		return event
	})
}
