package log

import (
	"log"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/json"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/filter"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	ctable "github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/table"
)

type table struct {
	setFocusFn      func(primitive tview.Primitive)
	store           *telemetry.Store
	view            *tview.Flex
	table           *tview.Table
	logData         *ctable.LogDataForTable
	filter          *filter.Filter
	detail          *detail
	body            *body
	resolvedLogBody string
}

func newTable(
	commands *tview.TextView,
	setFocusFn func(primitive tview.Primitive),
	store *telemetry.Store,
	detail *detail,
	body *body,
	resizeManagers []*layout.ResizeManager,
) *table {
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Logs (o)").SetBorder(true)

	t := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	filter := filter.NewFilter(
		commands,
		"Filter by service or body (/): ",
		func(inputConfirmed string, _ telemetry.SortType) {
			store.ApplyFilterLogs(inputConfirmed)
		},
		func() {
			setFocusFn(t)
		},
		nil,
		func(inputConfirmed string, sortType telemetry.SortType) {
			store.ApplyFilterTraces(inputConfirmed, sortType)
		},
	)

	logData := ctable.NewLogDataForTable(store.GetFilteredLogs())
	t.SetContent(&logData)
	store.SetOnSpanAdded(func() {
		t.Select(t.GetSelection())
	})

	stable := &table{
		setFocusFn: setFocusFn,
		store:      store,
		view:       container,
		table:      t,
		logData:    &logData,
		filter:     filter,
		detail:     detail,
		body:       body,
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
			Description: "Search traces",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				if !t.filter.View().HasFocus() {
					t.setFocusFn(t.filter.View())
				}
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyCtrlF, ' ', tcell.ModNone),
			Description: "Toggle full datetime",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				t.logData.SetFullDatetime(!t.logData.IsFullDatetime())
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),
			Description: "Copy log to clipboard",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				if err := clipboard.WriteAll(t.resolvedLogBody); err != nil {
					log.Printf("Failed to copy log body to clipboard: %v", err)
				} else {
					log.Println("Selected log body has been copied to your clipboard")
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
		selected := t.store.GetFilteredLogByIdx(row - 1)
		if selected == nil {
			return
		}
		t.detail.update(selected)
		log.Printf("selected row(original): %d", row)

		t.resolvedLogBody = json.PrettyJSON(selected.GetResolvedBody())
		t.body.update(t.resolvedLogBody)
	}
}
