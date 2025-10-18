package log

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

const (
	defaultTableProportion  = 30
	defaultDetailProportion = 20
	defaultMainProportion   = 15
	defaultBodyProportion   = 3
)

type LogPage struct {
	setFocusFn func(primitive tview.Primitive)
	view       *tview.Flex
	table      *table
	detail     *detail
	body       *body
}

func NewLogPage(
	setFocusFn func(primitive tview.Primitive),
	showModalFn layout.ShowModalFunc,
	hideModalFn layout.HideModalFunc,
	drawTimelineFn func(traceID string),
	store *telemetry.Store,
) *LogPage {
	commands := layout.NewCommandList()
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	mainContainer := tview.NewFlex().SetDirection(tview.FlexColumn)

	resizeManager := layout.NewResizeManager(layout.ResizeDirectionVertical)
	mainResizeManager := layout.NewResizeManager(layout.ResizeDirectionHorizontal)
	detail := newDetail(commands, showModalFn, hideModalFn, drawTimelineFn, []*layout.ResizeManager{
		mainResizeManager,
		resizeManager,
	}, store.GetTraceCache())
	body := newBody(commands, resizeManager)
	table := newTable(commands, setFocusFn, store, detail, body, []*layout.ResizeManager{
		mainResizeManager,
		resizeManager,
	})

	resizeManager.Register(
		container,
		mainContainer,
		body.view,
		defaultMainProportion,
		defaultBodyProportion,
		commands,
	)
	mainResizeManager.Register(
		mainContainer,
		table.view,
		detail.view,
		defaultTableProportion,
		defaultDetailProportion,
		commands,
	)

	mainContainer.AddItem(table.view, 0, defaultTableProportion, true).
		AddItem(detail.view, 0, defaultDetailProportion, false)
	container.AddItem(mainContainer, 0, defaultMainProportion, true).
		AddItem(body.view, 0, defaultBodyProportion, false)

	logPage := &LogPage{
		setFocusFn: setFocusFn,
		view:       container,
		table:      table,
		detail:     detail,
		body:       body,
	}

	logPage.view = layout.AttachTab(layout.AttachCommandList(commands, container), layout.PAGE_LOGS)

	logPage.registerCommands()
	store.RegisterOnFlushed(func() {
		logPage.flush()
	})

	return logPage
}

func (p *LogPage) GetPrimitive() tview.Primitive {
	return p.view
}

func (p *LogPage) flush() {
	p.detail.flush()
	p.body.flush()
}

func (p *LogPage) registerCommands() {
	p.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !p.table.filter.View().HasFocus() {
			switch event.Rune() {
			case 'd':
				p.setFocusFn(p.detail.view)
				return nil
			case 'o':
				p.setFocusFn(p.table.view)
				return nil
			case 'b':
				p.setFocusFn(p.body.view)
				return nil
			}
		}

		return event
	})
}
