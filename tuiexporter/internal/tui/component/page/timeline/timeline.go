package timeline

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
)

const (
	defaultGridProportion   = 29
	defaultDetailProportion = 21
)

type TimelinePage struct {
	switchToPageFn func()
	commands       *tview.TextView
	base           *tview.Flex
	container      *tview.Flex
	mainContainer  *tview.Flex
	store          *telemetry.Store
	onEscape       func()
	detail         *detail
	grid           *grid
	logPane        *logPane
	isLogCollapsed bool
	traceID        string
}

func NewTimelinePage(
	showModalFn layout.ShowModalFunc,
	hideModalFn layout.HideModalFunc,
	switchToPageFn func(),
	store *telemetry.Store,
	onEscape func(),
) *TimelinePage {
	commands := layout.NewCommandList()
	base := tview.NewFlex().SetDirection(tview.FlexRow)
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetBorder(false)
	mainContainer := tview.NewFlex().SetDirection(tview.FlexColumn)

	base.AddItem(container, 0, 1, true)

	resizeManager := layout.NewResizeManager(layout.ResizeDirectionHorizontal)
	detail := newDetail(commands, showModalFn, hideModalFn, resizeManager)
	logPane := newLogPane(commands, showModalFn, hideModalFn, store.GetLogCache())
	grid := newGrid(commands, store.GetTraceCache(), resizeManager, detail, logPane)

	resizeManager.Register(
		mainContainer,
		grid.gridView,
		detail.view,
		defaultGridProportion,
		defaultDetailProportion,
		commands,
	)

	timeline := &TimelinePage{
		switchToPageFn: switchToPageFn,
		commands:       commands,
		base:           base,
		container:      container,
		mainContainer:  mainContainer,
		store:          store,
		onEscape:       onEscape,
		detail:         detail,
		grid:           grid,
		logPane:        logPane,
		isLogCollapsed: true,
	}

	timeline.updateContainer()
	timeline.registerCommands()

	timeline.base = layout.AttachCommandList(commands, timeline.base)

	return timeline
}

func (p *TimelinePage) GetPrimitive() tview.Primitive {
	return p.base
}

func (p *TimelinePage) ShowTimelineByRow(row int) {
	p.DrawTimeline(
		p.store.GetTraceIDByFilteredIdx(row),
	)
}

func (p *TimelinePage) DrawTimeline(traceID string) {
	p.traceID = traceID

	p.container.Clear()
	p.mainContainer.Clear()

	span := p.grid.updateGrid(traceID)
	p.detail.update(span)
	p.logPane.updateLog(traceID, span.Span.TraceID().String())

	p.updateContainer()

	p.switchToPageFn()
	navigation.Focus(p.grid.gridView)
}

func (p *TimelinePage) registerCommands() {
	keyMaps := layout.KeyMaps{
		{
			Key: tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				navigation.Focus(p.detail.view)
				return nil
			},
		},
		{
			Key: tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				navigation.Focus(p.grid.gridView)
				return nil
			},
		},
		{
			Key: tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				navigation.Focus(p.logPane.tableView)
				return nil
			},
		},
		{
			Key: tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone),
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				p.isLogCollapsed = !p.isLogCollapsed
				logHeight := 10
				if p.isLogCollapsed {
					logHeight = 2
				}
				p.container.Clear().AddItem(p.mainContainer, 0, 1, p.mainContainer.HasFocus()).
					AddItem(p.logPane.tableView, logHeight, 1, p.logPane.tableView.HasFocus())

				return nil
			},
		},
		{
			Key: tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone),
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				p.logPane.toggleAllLogs(p.traceID, p.grid.getCurrentSpan())
				return nil
			},
		},
		{
			Key: tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone),
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				p.onEscape()
				return nil
			},
		},
	}
	layout.RegisterCommandList(p.commands, p.container, nil, keyMaps)
}

func (p *TimelinePage) updateContainer() {
	p.mainContainer.AddItem(p.grid.gridView, 0, defaultGridProportion, true).
		AddItem(p.detail.view, 0, defaultDetailProportion, false)
	p.container.AddItem(p.mainContainer, 0, 1, true).
		AddItem(p.logPane.tableView, 2, 1, false)
}
