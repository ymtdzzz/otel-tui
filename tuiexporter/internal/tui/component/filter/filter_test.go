package filter

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/mock"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"gotest.tools/v3/assert"
)

type filterCallbackMock struct {
	mock.Mock
}

func (m *filterCallbackMock) OnInputEnter(inputConfirmed string, sortType telemetry.SortType) {
	m.Called(inputConfirmed, sortType)
}
func (m *filterCallbackMock) OnInputDone() {
	m.Called()
}
func (m *filterCallbackMock) OnInputChanged(text string) {
	m.Called(text)
}
func (m *filterCallbackMock) OnSortTypeChanged(inputConfirmed string, sortType telemetry.SortType) {
	m.Called(inputConfirmed, sortType)
}

func TestDrawFilter(t *testing.T) {
	sw, sh := 50, 1
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	setup := func() *Filter {
		return NewFilter(layout.NewCommandList(), "test input: ", nil, nil, nil, nil)
	}

	t.Run("initial drawing", func(t *testing.T) {
		filter := setup()
		filter.view.Draw(screen)
		screen.Sync()

		got := test.GetScreenContent(t, screen)
		want := test.LoadTestdata(t, "tui/component/filter/filter_initial.txt")

		assert.Equal(t, want, got.String())
	})

	t.Run("input change", func(t *testing.T) {
		filter := setup()
		handler := filter.view.InputHandler()

		mockcb := &filterCallbackMock{}
		mockcb.On("OnInputChanged", "a").Once()
		mockcb.On("OnInputChanged", "a-").Once()

		filter.onInputChangedFn = mockcb.OnInputChanged
		filter.view.Focus(nil)

		handler(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone), nil)
		handler(tcell.NewEventKey(tcell.KeyRune, '-', tcell.ModNone), nil)

		filter.view.Draw(screen)
		screen.Sync()

		got := test.GetScreenContent(t, screen)
		want := test.LoadTestdata(t, "tui/component/filter/filter_input_change.txt")

		assert.Equal(t, want, got.String())
		assert.Equal(t, "a-", filter.input)
		assert.Equal(t, "", filter.inputConfirmed)

		mockcb.AssertExpectations(t)
	})

	t.Run("done", func(t *testing.T) {
		setup := func() *Filter {
			filter := NewFilter(layout.NewCommandList(), "test input: ", nil, nil, nil, nil)
			filter.input = "a-"
			filter.inputConfirmed = ""
			filter.view.SetText("a-")
			return filter
		}

		t.Run("enter", func(t *testing.T) {
			filter := setup()
			handler := filter.view.InputHandler()

			mockcb := &filterCallbackMock{}
			mockcb.On("OnInputEnter", "a-", telemetry.SORT_TYPE_NONE).Once()
			mockcb.On("OnInputDone").Once()

			filter.onInputEnterFn = mockcb.OnInputEnter
			filter.onInputDoneFn = mockcb.OnInputDone

			filter.view.Focus(nil)

			handler(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), nil)

			assert.Equal(t, "a-", filter.input)
			assert.Equal(t, "a-", filter.inputConfirmed)
			assert.Equal(t, "a-", filter.view.GetText())

			mockcb.AssertExpectations(t)
		})

		t.Run("escape", func(t *testing.T) {
			filter := setup()
			handler := filter.view.InputHandler()

			mockcb := &filterCallbackMock{}
			mockcb.On("OnInputDone").Once()

			filter.onInputEnterFn = mockcb.OnInputEnter
			filter.onInputDoneFn = mockcb.OnInputDone

			filter.view.Focus(nil)

			handler(tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone), nil)

			assert.Equal(t, "", filter.input)
			assert.Equal(t, "", filter.inputConfirmed)
			assert.Equal(t, "", filter.view.GetText())

			mockcb.AssertExpectations(t)
			mockcb.AssertNotCalled(t, "OnInputEnter")
		})
	})

	t.Run("rotate sort type", func(t *testing.T) {
		tests := []struct {
			name  string
			input telemetry.SortType
			want  telemetry.SortType
		}{
			{
				name:  "None to Latency Desc",
				input: telemetry.SORT_TYPE_NONE,
				want:  telemetry.SORT_TYPE_LATENCY_DESC,
			},
			{
				name:  "Latency Desc to Latency Asc",
				input: telemetry.SORT_TYPE_LATENCY_DESC,
				want:  telemetry.SORT_TYPE_LATENCY_ASC,
			},
			{
				name:  "Latency Asc to None",
				input: telemetry.SORT_TYPE_LATENCY_ASC,
				want:  telemetry.SORT_TYPE_NONE,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				filter := setup()
				filter.sortType = tt.input

				mockcb := &filterCallbackMock{}
				mockcb.On("OnSortTypeChanged", "", tt.want).Once()

				filter.onSortTypeChangedFn = mockcb.OnSortTypeChanged
				filter.RotateSortType()

				assert.Equal(t, tt.want, filter.sortType)
				mockcb.AssertExpectations(t)
			})
		}
	})
}
