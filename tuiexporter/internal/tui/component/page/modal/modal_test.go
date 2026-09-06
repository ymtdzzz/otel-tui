package modal

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/mock"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
	"gotest.tools/v3/assert"
)

type mockModalPageHandler struct {
	mock.Mock
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
}
