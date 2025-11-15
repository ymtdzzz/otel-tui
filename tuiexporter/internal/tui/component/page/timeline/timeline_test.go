package timeline

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jonboulle/clockwork"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/mock"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"gotest.tools/v3/assert"
)

type mockTimelineHandler struct {
	mock.Mock
}

func (m *mockTimelineHandler) switchToPageHandler() {
	m.Called()
}

func (m *mockTimelineHandler) onEscapeHandler() {
	m.Called()
}

func setupTimelinePage(t *testing.T) (*mockTimelineHandler, *TimelinePage, tcell.SimulationScreen, *telemetry.Store) {
	t.Helper()

	mockHandler := new(mockTimelineHandler)
	mockClock := clockwork.NewFakeClockAt(time.Date(2025, 11, 9, 12, 15, 0, 0, time.UTC))
	store := telemetry.NewStore(mockClock)

	sw, sh := 220, 50
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	page := NewTimelinePage(mockHandler.switchToPageHandler, store, mockHandler.onEscapeHandler)
	page.base.Focus(func(p tview.Primitive) {
		page.container.Focus(func(p tview.Primitive) {
			page.mainContainer.Focus(func(p tview.Primitive) {
				page.grid.gridView.Focus(nil)
			})
		})
	})

	page.base.SetRect(0, 0, sw, sh)
	page.base.Draw(screen)
	screen.Sync()

	return mockHandler, page, screen, store
}

func TestTimelinePage(t *testing.T) {
	t.Run("initial rendering", func(t *testing.T) {
		mockHandler, page, screen, store := setupTimelinePage(t)

		payload, spans := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{3}})
		store.AddSpan(&payload)

		mockHandler.On("switchToPageHandler").Return().Once()

		page.DrawTimeline(spans.Spans[0].TraceID().String())
		page.grid.gridView.Focus(nil)
		page.base.Draw(screen)
		screen.Sync()

		got := test.GetScreenContent(t, screen)
		want := test.LoadTestdata(t, "tui/component/page/timeline/timeline_initial.txt")

		assert.Equal(t, want, got.String())
		mockHandler.AssertExpectations(t)
	})

	t.Run("key event handling", func(t *testing.T) {
		t.Run("grid", func(t *testing.T) {
			t.Run("escape", func(t *testing.T) {
				mockHandler, page, screen, store := setupTimelinePage(t)

				payload, spans := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{3}})
				store.AddSpan(&payload)

				mockHandler.On("switchToPageHandler").Return().Once()
				mockHandler.On("onEscapeHandler").Return().Once()

				page.DrawTimeline(spans.Spans[0].TraceID().String())
				page.grid.gridView.Focus(nil)
				page.base.Draw(screen)
				screen.Sync()

				handler := page.base.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyEscape, ' ', tcell.ModNone), nil)

				mockHandler.AssertExpectations(t)
			})

			t.Run("change selection", func(t *testing.T) {
				mockHandler, page, screen, store := setupTimelinePage(t)

				payload, spans := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{3}})
				store.AddSpan(&payload)

				mockHandler.On("switchToPageHandler").Return().Once()

				page.DrawTimeline(spans.Spans[0].TraceID().String())
				page.grid.gridView.Focus(nil)

				handler := page.base.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone), nil)

				page.base.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/timeline/timeline_grid_change_selection.txt")

				assert.Equal(t, want, got.String())
			})

			t.Run("collapse log pane", func(t *testing.T) {
				mockHandler, page, screen, store := setupTimelinePage(t)

				payload, spans := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{3}})
				store.AddSpan(&payload)

				mockHandler.On("switchToPageHandler").Return().Once()

				page.DrawTimeline(spans.Spans[0].TraceID().String())
				page.grid.gridView.Focus(nil)

				handler := page.base.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone), nil)

				page.base.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/timeline/timeline_grid_collapse_log_pane.txt")

				assert.Equal(t, want, got.String())
			})

			tests := []struct {
				name            string
				key             *tcell.EventKey
				wantContentPath string
			}{
				{
					name:            "left",
					key:             tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/timeline/timeline_grid_key_handling_divider_left.txt",
				},
				{
					name:            "right",
					key:             tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/timeline/timeline_grid_key_handling_divider_right.txt",
				},
			}

			for _, tt := range tests {
				t.Run("move divider "+tt.name, func(t *testing.T) {
					mockHandler, page, screen, store := setupTimelinePage(t)

					payload, spans := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
					store.AddSpan(&payload)

					mockHandler.On("switchToPageHandler").Return().Once()

					page.DrawTimeline(spans.Spans[0].TraceID().String())
					page.grid.gridView.Focus(nil)

					handler := page.base.InputHandler()
					for range 5 {
						handler(tt.key, nil)
					}

					page.base.Draw(screen)
					screen.Sync()

					got := test.GetScreenContent(t, screen)
					want := test.LoadTestdata(t, tt.wantContentPath)

					assert.Equal(t, want, got.String())
				})
			}
		})

		t.Run("detail", func(t *testing.T) {
			tests := []struct {
				name            string
				key             *tcell.EventKey
				wantContentPath string
			}{
				{
					name:            "left",
					key:             tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/timeline/timeline_detail_key_handling_divider_left.txt",
				},
				{
					name:            "right",
					key:             tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/timeline/timeline_detail_key_handling_divider_right.txt",
				},
			}

			for _, tt := range tests {
				t.Run("move divider "+tt.name, func(t *testing.T) {
					mockHandler, page, screen, store := setupTimelinePage(t)

					payload, spans := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
					store.AddSpan(&payload)

					mockHandler.On("switchToPageHandler").Return().Once()

					page.DrawTimeline(spans.Spans[0].TraceID().String())
					page.grid.gridView.Blur()
					page.detail.view.Focus(func(p tview.Primitive) {
						p.Focus(nil)
					})

					handler := page.base.InputHandler()
					for range 5 {
						handler(tt.key, nil)
					}

					page.base.Draw(screen)
					screen.Sync()

					got := test.GetScreenContent(t, screen)
					want := test.LoadTestdata(t, tt.wantContentPath)

					assert.Equal(t, want, got.String())
				})
			}
		})

		t.Run("log", func(t *testing.T) {
			t.Run("current span and collapsed", func(t *testing.T) {
				mockHandler, page, screen, store := setupTimelinePage(t)

				payload, spans := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{3}})
				store.AddSpan(&payload)

				lpayload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
				lpayload.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).SetSpanID(spans.Spans[0].SpanID())
				lpayload.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(1).SetSpanID(spans.Spans[1].SpanID())
				store.AddLog(&lpayload)

				mockHandler.On("switchToPageHandler").Return().Once()

				page.DrawTimeline(spans.Spans[0].TraceID().String())
				page.grid.gridView.Focus(nil)

				page.base.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/timeline/timeline_log_current_span.txt")

				assert.Equal(t, want, got.String())
			})

			t.Run("all span and not collapsed", func(t *testing.T) {
				mockHandler, page, screen, store := setupTimelinePage(t)

				payload, spans := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{3}})
				store.AddSpan(&payload)

				lpayload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
				lpayload.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).SetSpanID(spans.Spans[0].SpanID())
				lpayload.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(1).SetSpanID(spans.Spans[1].SpanID())
				store.AddLog(&lpayload)

				mockHandler.On("switchToPageHandler").Return().Once()

				page.DrawTimeline(spans.Spans[0].TraceID().String())
				page.grid.gridView.Focus(nil)

				handler := page.base.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone), nil)
				handler(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone), nil)

				page.base.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/timeline/timeline_log_all_spans.txt")

				assert.Equal(t, want, got.String())
			})
		})
	})
}
