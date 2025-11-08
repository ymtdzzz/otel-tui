package log

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
	defaultMainProportion   = 15
	defaultBodyProportion   = 3
)

type LogPage struct {
	view   *tview.Flex
	table  *table
	detail *detail
	body   *body
}

func NewLogPage(
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
	table := newTable(commands, store, detail, body, []*layout.ResizeManager{
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
		view:   container,
		table:  table,
		detail: detail,
		body:   body,
	}

	logPage.view = layout.AttachTab(layout.AttachCommandList(commands, container), layout.PageIDLogs)

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
				navigation.Focus(p.detail.view)
				return nil
			case 'o':
				navigation.Focus(p.table.view)
				return nil
			case 'b':
				navigation.Focus(p.body.view)
				return nil
			}
		}

		return event
	})
}
