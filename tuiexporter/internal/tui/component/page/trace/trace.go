package trace

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
)

const (
	defaultTableProportion  = 30
	defaultDetailProportion = 20
)

type TracePage struct {
	view   *tview.Flex
	table  *table
	detail *detail
}

func NewTracePage(
	showModalFn layout.ShowModalFunc,
	hideModalFn layout.HideModalFunc,
	onSelectTableRow func(row, column int),
	store *telemetry.Store,
) *TracePage {
	commands := layout.NewCommandList()
	container := tview.NewFlex().SetDirection(tview.FlexColumn)

	resizeManager := layout.NewResizeManager(layout.ResizeDirectionHorizontal)
	detail := newDetail(commands, showModalFn, hideModalFn, resizeManager)
	table := newTable(commands, onSelectTableRow, store, detail, resizeManager)

	resizeManager.Register(
		container,
		table.view,
		detail.view,
		defaultTableProportion,
		defaultDetailProportion,
		commands,
	)

	container.AddItem(table.view, 0, defaultTableProportion, true).
		AddItem(detail.view, 0, defaultDetailProportion, false)

	trace := &TracePage{
		view:   container,
		table:  table,
		detail: detail,
	}

	trace.view = layout.AttachTab(layout.AttachCommandList(commands, container), layout.PageIDTraces)

	trace.registerCommands()
	store.RegisterOnFlushed(func() {
		trace.flush()
	})

	return trace
}

func (p *TracePage) GetPrimitive() tview.Primitive {
	return p.view
}

func (p *TracePage) flush() {
	p.detail.flush()
}

func (p *TracePage) registerCommands() {
	p.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !p.table.filter.View().HasFocus() {
			switch event.Rune() {
			case 'd':
				navigation.Focus(p.detail.view)
				return nil
			case 't':
				navigation.Focus(p.table.view)
				return nil
			}
		}

		return event
	})
}
