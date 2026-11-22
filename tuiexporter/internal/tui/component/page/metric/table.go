package metric

import (
	"log"
	"slices"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/filter"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
	ctable "github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/table"
)

type table struct {
	store      *telemetry.Store
	view       *tview.Flex
	table      *tview.Table
	metricData *ctable.MetricDataForTable
	filter     *filter.Filter
	detail     *detail
	chart      *chart
	selected   *telemetry.MetricData
}

func newTable(
	commands *tview.TextView,
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
			navigation.Focus(t)
		},
		nil,
		func(inputConfirmed string, _ telemetry.SortType) {
			store.ApplyFilterMetrics(inputConfirmed)
		},
	)

	metricData := ctable.NewMetricDataForTable(store.GetFilteredMetrics())
	t.SetContent(&metricData)
	stable := &table{
		store:      store,
		view:       container,
		table:      t,
		metricData: &metricData,
		filter:     filter,
		detail:     detail,
		chart:      chart,
	}

	t.SetSelectionChangedFunc(stable.onSelectionChangedFunc())
	store.SetOnMetricAdded(stable.onMetricAdded)

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
					navigation.Focus(t.filter.View())
				}
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyCtrlX, ' ', tcell.ModNone),
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
	layout.RegisterCommandList(commands, t.table, nil, keyMaps)
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
		t.selected = selected
		t.detail.update(selected)
		t.chart.update(selected)
		log.Printf("selected row(original): %d", row)
	}
}

// onMetricAdded keeps the selected metric selected when the new metrics sort in above it.
func (t *table) onMetricAdded() {
	if t.detail.tree.GetRoot() == nil {
		// Select the first data row (row 1), not the header (row 0)
		t.table.Select(1, 0)
		return
	}
	idx := slices.Index(*t.store.GetFilteredMetrics(), t.selected)
	if row, _ := t.table.GetSelection(); idx >= 0 && row != idx+1 {
		t.table.Select(idx+1, 0)
	}
}
