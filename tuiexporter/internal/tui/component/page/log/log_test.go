package log

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

type mockDrawTimelineHandler struct {
	mock.Mock
}

func (m *mockDrawTimelineHandler) DrawTimeline(traceID string) {
	m.Called(traceID)
}

func setupLogPage(t *testing.T) (*mockDrawTimelineHandler, *LogPage, tcell.SimulationScreen, *telemetry.Store) {
	t.Helper()

	mockHandler := new(mockDrawTimelineHandler)
	mockClock := clockwork.NewFakeClockAt(time.Date(2025, 11, 9, 12, 15, 0, 0, time.UTC))
	store := telemetry.NewStore(mockClock)

	sw, sh := 220, 50
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to initialize screen: %v", err)
	}
	screen.SetSize(sw, sh)

	page := NewLogPage(mockHandler.DrawTimeline, store)
	page.table.table.Focus(nil)

	page.view.SetRect(0, 0, sw, sh)
	page.view.Draw(screen)
	screen.Sync()

	return mockHandler, page, screen, store
}

func TestLogPage(t *testing.T) {
	t.Run("initial rendering and receive first log", func(t *testing.T) {
		_, page, screen, store := setupLogPage(t)

		got := test.GetScreenContent(t, screen)
		want := test.LoadTestdata(t, "tui/component/page/log/log_initial.txt")

		assert.Equal(t, want, got.String())

		// Related issue: https://github.com/ymtdzzz/otel-tui/issues/214
		payload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
		store.AddLog(&payload)

		page.view.Draw(screen)
		screen.Sync()

		got = test.GetScreenContent(t, screen)
		want = test.LoadTestdata(t, "tui/component/page/log/log_first_log_received.txt")

		assert.Equal(t, want, got.String())
	})

	// Related issue: https://github.com/ymtdzzz/otel-tui/issues/354
	t.Run("receive new span when the details pane is focused", func(t *testing.T) {
		_, page, screen, store := setupLogPage(t)

		payload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
		store.AddLog(&payload)

		page.table.table.Blur()
		page.detail.view.Focus(func(p tview.Primitive) {
			p.Focus(nil)
		})

		page.view.Draw(screen)
		screen.Sync()

		newPayload, _ := test.GenerateOTLPLogsPayload(t, 2, 1, []int{1}, [][]int{{1}})
		store.AddLog(&newPayload)

		page.view.Draw(screen)
		screen.Sync()

		assert.Equal(t, true, page.detail.view.HasFocus())
	})

	t.Run("key event handling", func(t *testing.T) {
		t.Run("table", func(t *testing.T) {
			t.Run("filter logs", func(t *testing.T) {
				_, page, screen, store := setupLogPage(t)

				payload1, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
				payload1.ResourceLogs().At(0).Resource().Attributes().PutStr("service.name", "service-1")
				payload2, _ := test.GenerateOTLPLogsPayload(t, 2, 1, []int{1}, [][]int{{1}})
				payload2.ResourceLogs().At(0).Resource().Attributes().PutStr("service.name", "service-2")
				payload3, _ := test.GenerateOTLPLogsPayload(t, 3, 1, []int{1}, [][]int{{1}})
				payload3.ResourceLogs().At(0).Resource().Attributes().PutStr("service.name", "service-3")
				store.AddLog(&payload1)
				store.AddLog(&payload2)
				store.AddLog(&payload3)

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
				want := test.LoadTestdata(t, "tui/component/page/log/log_table_filter_logs.txt")

				assert.Equal(t, want, got.String())
			})

			t.Run("change selection", func(t *testing.T) {
				_, page, screen, store := setupLogPage(t)

				payload1, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
				payload1.ResourceLogs().At(0).Resource().Attributes().PutStr("service.name", "service-1")
				payload2, _ := test.GenerateOTLPLogsPayload(t, 2, 1, []int{1}, [][]int{{1}})
				payload2.ResourceLogs().At(0).Resource().Attributes().PutStr("service.name", "service-2")
				payload3, _ := test.GenerateOTLPLogsPayload(t, 3, 1, []int{1}, [][]int{{1}})
				payload3.ResourceLogs().At(0).Resource().Attributes().PutStr("service.name", "service-3")
				store.AddLog(&payload1)
				store.AddLog(&payload2)
				store.AddLog(&payload3)

				handler := page.table.view.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone), nil)
				handler(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone), nil)

				page.view.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/log/log_table_change_selection.txt")

				assert.Equal(t, want, got.String())
			})

			t.Run("flush", func(t *testing.T) {
				_, page, screen, store := setupLogPage(t)

				payload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
				store.AddLog(&payload)

				handler := page.table.view.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyCtrlX, ' ', tcell.ModNone), nil)

				page.view.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/log/log_table_flush.txt")

				assert.Equal(t, want, got.String())

				// After flush, when the next span is received, the detail pane renders its content
				newPayload, _ := test.GenerateOTLPLogsPayload(t, 2, 1, []int{1}, [][]int{{1}})
				store.AddLog(&newPayload)

				page.view.Draw(screen)
				screen.Sync()

				got = test.GetScreenContent(t, screen)
				want = test.LoadTestdata(t, "tui/component/page/log/log_table_flush_log_received.txt")

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
					wantContentPath: "tui/component/page/log/log_table_key_handling_divider_left.txt",
				},
				{
					name:            "right",
					key:             tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/log/log_table_key_handling_divider_right.txt",
				},
				{
					name:            "up",
					key:             tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/log/log_table_key_handling_divider_up.txt",
				},
				{
					name:            "down",
					key:             tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/log/log_table_key_handling_divider_down.txt",
				},
			}

			for _, tt := range tests {
				t.Run("move divider "+tt.name, func(t *testing.T) {
					_, page, screen, store := setupLogPage(t)

					payload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
					store.AddLog(&payload)

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
				_, page, _, store := setupLogPage(t)

				payload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
				store.AddLog(&payload)

				page.detail.flush()
				page.detail.commands.SetText("")

				// Assert that SetFocusFunc is called even after flushing
				page.detail.view.Focus(func(p tview.Primitive) {
					p.Focus(nil)
				})

				assert.True(t, len(page.detail.commands.GetText(true)) > 0)
			})

			t.Run("jump to trace", func(t *testing.T) {
				mockHandler, page, _, store := setupLogPage(t)

				tpayload, _ := test.GenerateOTLPTracesPayload(t, 1, 1, []int{1}, [][]int{{1}})
				store.AddSpan(&tpayload)

				payload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
				store.AddLog(&payload)

				page.table.table.Blur()
				page.detail.view.Focus(func(p tview.Primitive) {
					p.Focus(nil)
				})

				mockHandler.On("DrawTimeline", "01000000000000000000000000000000").Once()

				handler := page.detail.view.InputHandler()
				for range 16 {
					handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil)
				}
				handler(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), nil)

				mockHandler.AssertExpectations(t)
			})

			tests := []struct {
				name            string
				key             *tcell.EventKey
				wantContentPath string
			}{
				{
					name:            "left",
					key:             tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/log/log_detail_key_handling_divider_left.txt",
				},
				{
					name:            "right",
					key:             tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/log/log_detail_key_handling_divider_right.txt",
				},
				{
					name:            "up",
					key:             tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/log/log_detail_key_handling_divider_up.txt",
				},
				{
					name:            "down",
					key:             tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/log/log_detail_key_handling_divider_down.txt",
				},
			}

			for _, tt := range tests {
				t.Run("move divider "+tt.name, func(t *testing.T) {
					_, page, screen, store := setupLogPage(t)

					payload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
					store.AddLog(&payload)

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

		t.Run("body", func(t *testing.T) {
			tests := []struct {
				name            string
				key             *tcell.EventKey
				wantContentPath string
			}{
				{
					name:            "up",
					key:             tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/log/log_body_key_handling_divider_up.txt",
				},
				{
					name:            "down",
					key:             tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/log/log_body_key_handling_divider_down.txt",
				},
			}

			for _, tt := range tests {
				t.Run("move divider "+tt.name, func(t *testing.T) {
					_, page, screen, store := setupLogPage(t)

					payload, _ := test.GenerateOTLPLogsPayload(t, 1, 1, []int{1}, [][]int{{1}})
					store.AddLog(&payload)

					page.table.table.Blur()
					page.body.view.Focus(nil)

					handler := page.body.view.InputHandler()
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
