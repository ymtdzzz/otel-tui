package metric

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/filter"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	ctable "github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/table"
)

type table struct {
	setFocusFn func(primitive tview.Primitive)
	store      *telemetry.Store
	view       *tview.Flex
	table      *tview.Table
	metricData *ctable.MetricDataForTable
	filter     *filter.Filter
	detail     *detail
	chart      *chart
}

func newTable(
	commands *tview.TextView,
	setFocusFn func(primitive tview.Primitive),
	store *telemetry.Store,
	detail *detail,
	chart *chart,
	resizeManagers []*layout.ResizeManager,
) *table {
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Metrics (m)").SetBorder(true)

	t := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	filter := filter.NewFilter(
		commands,
		"Filter by service or metric name (/): ",
		func(inputConfirmed string, _ telemetry.SortType) {
			store.ApplyFilterMetrics(inputConfirmed)
		},
		func() {
			setFocusFn(t)
		},
		nil,
		func(inputConfirmed string, _ telemetry.SortType) {
			store.ApplyFilterMetrics(inputConfirmed)
		},
	)

	metricData := ctable.NewMetricDataForTable(store.GetFilteredMetrics())
	t.SetContent(&metricData)
	store.SetOnMetricAdded(func() {
		t.Select(t.GetSelection())
	})

	stable := &table{
		setFocusFn: setFocusFn,
		store:      store,
		view:       container,
		table:      t,
		metricData: &metricData,
		filter:     filter,
		detail:     detail,
		chart:      chart,
	}

	t.SetSelectionChangedFunc(stable.onSelectionChangedFunc())

	container.
		AddItem(filter.View(), 1, 0, false).
		AddItem(t, 0, 1, true)

	stable.registerCommands(commands, resizeManagers)

	return stable
}

func (t *table) registerCommands(commands *tview.TextView, resizeManagers []*layout.ResizeManager) {
	keyMaps := layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
			Description: "Search metrics",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				if !t.filter.View().HasFocus() {
					t.setFocusFn(t.filter.View())
				}
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone),
			Description: "Clear all data",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				t.store.Flush()
				t.table.Select(0, 0)
				return nil
			},
		},
	}
	for _, rm := range resizeManagers {
		keyMaps.Merge(rm.KeyMaps())
	}
	layout.RegisterCommandList2(commands, t.table, nil, keyMaps)
}

func (t *table) onSelectionChangedFunc() func(row, col int) {
	return func(row, _ int) {
		if row == 0 {
			return
		}
		selected := t.store.GetFilteredMetricByIdx(row - 1)
		if selected == nil {
			return
		}
		t.detail.update(selected)
		t.chart.update(selected)
		log.Printf("selected row(original): %d", row)
	}
}
