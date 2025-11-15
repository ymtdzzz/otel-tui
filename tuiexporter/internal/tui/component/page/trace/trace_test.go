package trace

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jonboulle/clockwork"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
)

type mockSelectTableRowHandler struct {
	mock.Mock
}

func (m *mockSelectTableRowHandler) Handle(row, column int) {
	m.Called(row, column)
}

func setupTracePage(t *testing.T) (*mockSelectTableRowHandler, *TracePage, tcell.SimulationScreen, *telemetry.Store) {
	t.Helper()

	mockHandler := new(mockSelectTableRowHandler)
	mockClock := clockwork.NewFakeClockAt(time.Date(2025, 11, 9, 12, 15, 0, 0, time.UTC))
	store := telemetry.NewStore(mockClock)

	sw, sh := 220, 50
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	page := NewTracePage(mockHandler.Handle, store)
	page.table.table.Focus(nil)

	page.view.SetRect(0, 0, sw, sh)
	page.view.Draw(screen)
	screen.Sync()

	return mockHandler, page, screen, store
}

func TestTracePage(t *testing.T) {
	t.Run("initial rendering and receive first span", func(t *testing.T) {
		_, page, screen, store := setupTracePage(t)

		got := test.GetScreenContent(t, screen)
		want := test.LoadTestdata(t, "tui/component/page/trace/trace_initial.txt")

		assert.Equal(t, want, got.String())

		// Related issue: https://github.com/ymtdzzz/otel-tui/issues/214
		payload, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
		store.AddSpan(&payload)

		page.view.Draw(screen)
		screen.Sync()

		got = test.GetScreenContent(t, screen)
		want = test.LoadTestdata(t, "tui/component/page/trace/trace_first_span_received.txt")

		assert.Equal(t, want, got.String())
	})

	// Related issue: https://github.com/ymtdzzz/otel-tui/issues/354
	t.Run("receive new span when the details pane is focused", func(t *testing.T) {
		_, page, screen, store := setupTracePage(t)

		payload, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
		store.AddSpan(&payload)

		page.table.table.Blur()
		page.detail.view.Focus(func(p tview.Primitive) {
			p.Focus(nil)
		})

		page.view.Draw(screen)
		screen.Sync()

		newPayload, _ := test.GenerateOTLPTracesPayload(t, 2, 1, []int{1}, [][]int{{1}})
		store.AddSpan(&newPayload)

		page.view.Draw(screen)
		screen.Sync()

		assert.Equal(t, true, page.detail.view.HasFocus())
	})

	t.Run("key event handling", func(t *testing.T) {
		t.Run("table", func(t *testing.T) {
			t.Run("filter spans", func(t *testing.T) {
				_, page, screen, store := setupTracePage(t)

				payload1, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
				payload1.ResourceSpans().At(0).Resource().Attributes().PutStr("service.name", "service-1")
				payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SetName("trace-1")
				payload2, _ := test.GenerateOTLPTracesPayload(t, 2, 1, []int{1}, [][]int{{1}})
				payload2.ResourceSpans().At(0).Resource().Attributes().PutStr("service.name", "service-2")
				payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SetName("trace-2")
				payload3, _ := test.GenerateOTLPTracesPayload(t, 3, 1, []int{1}, [][]int{{1}})
				payload3.ResourceSpans().At(0).Resource().Attributes().PutStr("service.name", "service-3")
				payload3.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SetName("trace-3")
				store.AddSpan(&payload1)
				store.AddSpan(&payload2)
				store.AddSpan(&payload3)

				page.table.table.Blur()
				page.table.filter.View().Focus(nil)
				handler := page.table.filter.View().InputHandler()
				handler(tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone), nil)
				handler(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), nil)
				page.table.filter.View().Blur()
				page.table.table.Focus(nil)

				page.view.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/trace/trace_table_filter_spans.txt")

				assert.Equal(t, want, got.String())
			})

			t.Run("change selection", func(t *testing.T) {
				_, page, screen, store := setupTracePage(t)

				payload1, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
				payload1.ResourceSpans().At(0).Resource().Attributes().PutStr("service.name", "service-1")
				payload1.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SetName("trace-1")
				payload2, _ := test.GenerateOTLPTracesPayload(t, 2, 1, []int{1}, [][]int{{1}})
				payload2.ResourceSpans().At(0).Resource().Attributes().PutStr("service.name", "service-2")
				payload2.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SetName("trace-2")
				payload3, _ := test.GenerateOTLPTracesPayload(t, 3, 1, []int{1}, [][]int{{1}})
				payload3.ResourceSpans().At(0).Resource().Attributes().PutStr("service.name", "service-3")
				payload3.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SetName("trace-3")
				store.AddSpan(&payload1)
				store.AddSpan(&payload2)
				store.AddSpan(&payload3)

				handler := page.table.view.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone), nil)

				page.view.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/trace/trace_table_change_selection.txt")

				assert.Equal(t, want, got.String())
			})

			t.Run("select table row", func(t *testing.T) {
				mockHandler, page, screen, store := setupTracePage(t)

				payload, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
				store.AddSpan(&payload)

				mockHandler.On("Handle", 1, 0).Once()

				handler := page.table.view.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), nil)

				page.view.Draw(screen)
				screen.Sync()

				mockHandler.AssertExpectations(t)
			})

			t.Run("flush", func(t *testing.T) {
				_, page, screen, store := setupTracePage(t)

				payload, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
				store.AddSpan(&payload)

				handler := page.table.view.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyCtrlX, ' ', tcell.ModNone), nil)

				page.view.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/trace/trace_table_flush.txt")

				assert.Equal(t, want, got.String())

				// After flush, when the next span is received, the detail pane renders its content
				newPayload, _ := test.GenerateOTLPTracesPayload(t, 2, 1, []int{1}, [][]int{{1}})
				store.AddSpan(&newPayload)

				page.view.Draw(screen)
				screen.Sync()

				got = test.GetScreenContent(t, screen)
				want = test.LoadTestdata(t, "tui/component/page/trace/trace_table_flush_span_received.txt")

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
					wantContentPath: "tui/component/page/trace/trace_table_key_handling_divider_left.txt",
				},
				{
					name:            "right",
					key:             tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/trace/trace_table_key_handling_divider_right.txt",
				},
			}

			for _, tt := range tests {
				t.Run("move divider "+tt.name, func(t *testing.T) {
					_, page, screen, store := setupTracePage(t)

					payload, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
					store.AddSpan(&payload)

					handler := page.table.view.InputHandler()
					for range 5 {
						handler(tt.key, nil)
					}

					page.view.Draw(screen)
					screen.Sync()

					got := test.GetScreenContent(t, screen)
					want := test.LoadTestdata(t, tt.wantContentPath)

					assert.Equal(t, want, got.String())
				})
			}
		})

		t.Run("detail", func(t *testing.T) {
			t.Run("flush", func(t *testing.T) {
				_, page, _, store := setupTracePage(t)

				payload, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
				store.AddSpan(&payload)

				page.detail.flush()
				page.detail.commands.SetText("")

				// Assert that SetFocusFunc is called even after flushing
				page.detail.view.Focus(func(p tview.Primitive) {
					p.Focus(nil)
				})

				assert.True(t, len(page.detail.commands.GetText(true)) > 0)
			})

			tests := []struct {
				name            string
				key             *tcell.EventKey
				wantContentPath string
			}{
				{
					name:            "left",
					key:             tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/trace/trace_detail_key_handling_divider_left.txt",
				},
				{
					name:            "right",
					key:             tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/trace/trace_detail_key_handling_divider_right.txt",
				},
			}

			for _, tt := range tests {
				t.Run("move divider "+tt.name, func(t *testing.T) {
					_, page, screen, store := setupTracePage(t)

					payload, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
					store.AddSpan(&payload)

					page.table.table.Blur()
					page.detail.view.Focus(func(p tview.Primitive) {
						p.Focus(nil)
					})

					handler := page.detail.view.InputHandler()
					for range 5 {
						handler(tt.key, nil)
					}

					page.view.Draw(screen)
					screen.Sync()

					got := test.GetScreenContent(t, screen)
					want := test.LoadTestdata(t, tt.wantContentPath)

					assert.Equal(t, want, got.String())
				})
			}
		})
	})
}
