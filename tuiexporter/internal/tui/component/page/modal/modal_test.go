package modal

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/mock"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
	"gotest.tools/v3/assert"
)

type mockModalPageHandler struct {
	mock.Mock
}

type firstColumn struct{}

func (firstColumn) GetColumnIdx() int { return 0 }

func TestModalFocusTransitionDoesNotDeadlock(t *testing.T) {
	done := make(chan bool, 1)
	go func() {
		app := tview.NewApplication()
		table := tview.NewTable().SetSelectable(true, false)
		table.SetCell(0, 0, tview.NewTableCell("attribute"))
		modal := NewModalPage()
		pages := tview.NewPages().AddPage("table", table, true, true).
			AddPage("modal", modal.GetPrimitive(), true, false)
		app.SetRoot(pages, true)
		navigation.Init(func(p tview.Primitive) { app.SetFocus(p) },
			modal.ShowModalFunc(func() { pages.ShowPage("modal").SendToFront("modal") }),
			modal.HideModalFunc(func() { pages.SendToBack("modal").HidePage("modal") }))
		var queued []func()
		navigation.SetQueueUpdateFunc(func(update func()) { queued = append(queued, update) })
		layout.AttachModalForTableRows(table, firstColumn{}, nil)
		table.SetSelectionChangedFunc(nil)
		table.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
		name, _ := pages.GetFrontPage()
		if name != "modal" {
			done <- false
			return
		}
		destination := tview.NewBox()
		app.SetFocus(destination)
		for _, update := range queued {
			update()
		}
		name, _ = pages.GetFrontPage()
		done <- name == "table" && app.GetFocus() == destination
	}()
	select {
	case ok := <-done:
		assert.Assert(t, ok)
	case <-time.After(2 * time.Second):
		t.Fatal("opening or blurring a modal deadlocked the actual tview focus path")
	}
}

func (m *mockModalPageHandler) showModal() {
	m.Called()
}

func (m *mockModalPageHandler) hideModal() {
	m.Called()
}

func TestModalPage(t *testing.T) {
	t.Run("show and hide modal", func(t *testing.T) {
		mockHandler := new(mockModalPageHandler)
		modalPage := NewModalPage()

		showModalFn := modalPage.ShowModalFunc(mockHandler.showModal)
		hideModalFn := modalPage.HideModalFunc(mockHandler.hideModal)

		want := "This is a test modal text."

		mockHandler.On("showModal").Once()
		showModalFn(nil, want)
		mockHandler.AssertCalled(t, "showModal")
		assert.Equal(t, want, modalPage.textView.GetText(true))

		mockHandler.On("hideModal").Once()
		hideModalFn(nil)
		mockHandler.AssertCalled(t, "hideModal")
	})
}

func TestHideModalWithoutCurrentFocus(t *testing.T) {
	focusCalls := 0
	navigation.Init(func(tview.Primitive) { focusCalls++ }, nil, nil)

	modalPage := NewModalPage()
	hideModalFn := modalPage.HideModalFunc(func() {})
	hideModalFn(nil)

	assert.Equal(t, 0, focusCalls)

	current := tview.NewBox()
	hideModalFn(current)
	assert.Equal(t, 1, focusCalls)
}
