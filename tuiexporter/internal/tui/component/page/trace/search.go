package trace

import (
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

type search struct {
	setFocusFn            func(primitive tview.Primitive)
	table                 tview.Primitive
	store                 *telemetry.Store
	sortType              telemetry.SortType
	view                  *tview.InputField
	input, inputConfirmed string
}

func newSearch(
	commands *tview.TextView,
	setFocusFn func(primitive tview.Primitive),
	table tview.Primitive,
	store *telemetry.Store,
) *search {
	field := tview.NewInputField().
		SetLabel("Filter by service or span name (/): ").
		SetFieldWidth(20)

	search := &search{
		setFocusFn: setFocusFn,
		table:      table,
		store:      store,
		sortType:   telemetry.SORT_TYPE_NONE,
		view:       field,
	}

	field.SetDoneFunc(search.onDoneFunc())
	field.SetChangedFunc(search.onChangedFunc())

	search.registerCommands(commands)

	return search
}

func (s *search) onChangedFunc() func(text string) {
	return func(text string) {
		// remove the suffix '/' from input because it is passed from SetInputCapture()
		if strings.HasSuffix(text, "/") {
			text = strings.TrimSuffix(text, "/")
			s.view.SetText(text)
		}
		s.input = text
	}
}

func (s *search) onDoneFunc() func(key tcell.Key) {
	return func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			s.inputConfirmed = s.input
			log.Println("search service name: ", s.inputConfirmed)
			s.store.ApplyFilterTraces(s.inputConfirmed, s.sortType)
		case tcell.KeyEsc:
			s.view.SetText(s.inputConfirmed)
		}
		s.setFocusFn(s.table)
	}
}

func (s *search) changeSortType(sortType telemetry.SortType) {
	s.sortType = sortType
}

func (s *search) registerCommands(commands *tview.TextView) {
	layout.RegisterCommandList2(commands, s.view, nil, layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone),
			Description: "Cancel",
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			Description: "Confirm",
		},
	})
}
