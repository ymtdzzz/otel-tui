package log

import (
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
)

type body struct {
	view *tview.TextView
}

func newBody(
	commands *tview.TextView,
	resizeManager *layout.ResizeManager,
) *body {
	container := tview.NewTextView()
	container.SetTitle("Body (b)").SetBorder(true)

	b := &body{
		view: container,
	}

	b.registerCommands(commands, resizeManager)

	return b
}

func (b *body) flush() {
	b.view.Clear()
}

func (b *body) update(body string) {
	b.view.SetText(body)
}

func (b *body) registerCommands(commands *tview.TextView, resizeManager *layout.ResizeManager) {
	keyMaps := layout.KeyMaps{}
	keyMaps.Merge(resizeManager.KeyMaps())
	layout.RegisterCommandList2(commands, b.view, nil, keyMaps)
}
