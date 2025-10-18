package trace

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
	spanData   *ctable.SpanDataForTable
	filter     *filter.Filter
	detail     *detail
}

func newTable(
	commands *tview.TextView,
	setFocusFn func(primitive tview.Primitive),
	onSelectTableRow func(row, column int),
	store *telemetry.Store,
	detail *detail,
	resizeManager *layout.ResizeManager,
) *table {
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Traces (t)").SetBorder(true)

	t := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetSelectedFunc(onSelectTableRow).
		SetFixed(1, 0)

	filter := filter.NewFilter(
		commands,
		"Filter by service or span name (/): ",
		func(inputConfirmed string, sortType telemetry.SortType) {
			store.ApplyFilterTraces(inputConfirmed, sortType)
		},
		func() {
			setFocusFn(t)
		},
		nil,
		func(inputConfirmed string, sortType telemetry.SortType) {
			store.ApplyFilterTraces(inputConfirmed, sortType)
		},
	)

	spanData := ctable.NewSpanDataForTable(store.GetTraceCache(), store.GetFilteredSvcSpans(), filter.SortType())
	t.SetContent(&spanData)
	store.SetOnSpanAdded(func() {
		t.Select(t.GetSelection())
	})

	stable := &table{
		setFocusFn: setFocusFn,
		store:      store,
		view:       container,
		table:      t,
		spanData:   &spanData,
		filter:     filter,
		detail:     detail,
	}

	t.SetSelectionChangedFunc(stable.onSelectionChangedFunc())

	container.
		AddItem(filter.View(), 1, 0, false).
		AddItem(t, 0, 1, true)

	stable.registerCommands(commands, resizeManager)

	return stable
}

func (t *table) registerCommands(commands *tview.TextView, resizeManager *layout.ResizeManager) {
	keyMaps := layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
			Description: "Search traces",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				if !t.filter.View().HasFocus() {
					t.setFocusFn(t.filter.View())
				}
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyCtrlS, ' ', tcell.ModNone),
			Description: "Toggle sort (Latency)",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				t.filter.RotateSortType()
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyCtrlF, ' ', tcell.ModNone),
			Description: "Toggle full datetime",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				t.spanData.SetFullDatetime(!t.spanData.IsFullDatetime())
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone),
			Description: "Recalculate service root span",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				log.Println("Recalculate service root span")
				row, _ := t.table.GetSelection()
				if row == 0 {
					return nil
				}
				t.store.RecalculateServiceRootSpanByIdx(row - 1)

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
	keyMaps.Merge(resizeManager.KeyMaps())
	layout.RegisterCommandList2(commands, t.table, nil, keyMaps)
}

func (t *table) onSelectionChangedFunc() func(row, col int) {
	return func(row, _ int) {
		if row == 0 {
			return
		}
		spans := t.store.GetFilteredServiceSpansByIdx(row - 1)
		if spans == nil {
			return
		}
		t.detail.update(spans)
		log.Printf("selected row(original): %d", row)
	}
}
