package filter

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

type onInputEnterFn func(inputConfirmed string, sortType telemetry.SortType)
type onInputDoneFn func()
type onInputChangedFn func(text string)
type onSortTypeChangedFn func(inputConfirmed string, sortType telemetry.SortType)

type Filter struct {
	view                  *tview.InputField
	sortType              telemetry.SortType
	input, inputConfirmed string
	onInputEnterFn        onInputEnterFn
	onInputDoneFn         onInputDoneFn
	onInputChangedFn      onInputChangedFn
	onSortTypeChangedFn   onSortTypeChangedFn
}

func NewFilter(
	commands *tview.TextView,
	label string,
	onInputEnterFn onInputEnterFn,
	onInputDoneFn onInputDoneFn,
	onInputChangedFn onInputChangedFn,
	onSortTypeChangedFn onSortTypeChangedFn,
) *Filter {
	field := tview.NewInputField().
		SetLabel(label).
		SetFieldWidth(20)

	filter := &Filter{
		view:                field,
		sortType:            telemetry.SORT_TYPE_NONE,
		onInputEnterFn:      onInputEnterFn,
		onInputDoneFn:       onInputDoneFn,
		onInputChangedFn:    onInputChangedFn,
		onSortTypeChangedFn: onSortTypeChangedFn,
	}

	field.SetDoneFunc(filter.onDoneFunc())
	field.SetChangedFunc(filter.onChangedFunc())

	filter.registerCommands(commands)

	return filter
}

func (f *Filter) onChangedFunc() func(text string) {
	return func(text string) {
		f.input = text
		if f.onInputChangedFn != nil {
			f.onInputChangedFn(text)
		}
	}
}

func (f *Filter) onDoneFunc() func(key tcell.Key) {
	return func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			f.inputConfirmed = f.input
			if f.onInputEnterFn != nil {
				f.onInputEnterFn(f.inputConfirmed, f.sortType)
			}
		case tcell.KeyEsc:
			f.view.SetText(f.inputConfirmed)
		}
		if f.onInputDoneFn != nil {
			f.onInputDoneFn()
		}
	}
}

func (f *Filter) registerCommands(commands *tview.TextView) {
	layout.RegisterCommandList(commands, f.view, nil, layout.KeyMaps{
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

func (f *Filter) RotateSortType() {
	switch f.sortType {
	case telemetry.SORT_TYPE_NONE:
		f.sortType = telemetry.SORT_TYPE_LATENCY_DESC
	case telemetry.SORT_TYPE_LATENCY_DESC:
		f.sortType = telemetry.SORT_TYPE_LATENCY_ASC
	default:
		f.sortType = telemetry.SORT_TYPE_NONE
	}
	if f.onSortTypeChangedFn != nil {
		f.onSortTypeChangedFn(f.inputConfirmed, f.sortType)
	}
}

func (f *Filter) InputConfirmed() string {
	return f.inputConfirmed
}

func (f *Filter) SortType() *telemetry.SortType {
	return &f.sortType
}

func (f *Filter) View() *tview.InputField {
	return f.view
}
