package metric

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jonboulle/clockwork"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"gotest.tools/v3/assert"
)

func setupMetricPage(t *testing.T) (*MetricPage, tcell.SimulationScreen, *telemetry.Store) {
	t.Helper()

	mockClock := clockwork.NewFakeClockAt(time.Date(2025, 11, 9, 12, 15, 0, 0, time.UTC))
	store := telemetry.NewStore(mockClock)

	sw, sh := 220, 50
	screen := tcell.NewSimulationScreen("")
	screen.Init()
	screen.SetSize(sw, sh)

	page := NewMetricPage(store)
	page.table.table.Focus(nil)

	page.view.SetRect(0, 0, sw, sh)
	page.view.Draw(screen)
	screen.Sync()

	return page, screen, store
}

func TestMetricPage(t *testing.T) {
	t.Run("initial rendering and received first metric", func(t *testing.T) {
		page, screen, store := setupMetricPage(t)

		got := test.GetScreenContent(t, screen)
		want := test.LoadTestdata(t, "tui/component/page/metric/metric_initial.txt")

		assert.Equal(t, want, got.String())

		// Related issue: https://github.com/ymtdzzz/otel-tui/issues/214
		payload, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
		store.AddMetric(&payload)

		page.view.Draw(screen)
		screen.Sync()

		got = test.GetScreenContent(t, screen)
		want = test.LoadTestdata(t, "tui/component/page/metric/metric_first_metric_received.txt")

		assert.Equal(t, want, got.String())
	})

	// Related issue: https://github.com/ymtdzzz/otel-tui/issues/354
	t.Run("receive new span when the details pane is focused", func(t *testing.T) {
		page, screen, store := setupMetricPage(t)

		payload, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
		store.AddMetric(&payload)

		page.table.table.Blur()
		page.detail.view.Focus(func(p tview.Primitive) {
			p.Focus(nil)
		})

		page.view.Draw(screen)
		screen.Sync()

		newPayload, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
		store.AddMetric(&newPayload)

		page.view.Draw(screen)
		screen.Sync()

		assert.Equal(t, true, page.detail.view.HasFocus())
	})

	t.Run("key event handling", func(t *testing.T) {
		t.Run("table", func(t *testing.T) {
			t.Run("filter metrics", func(t *testing.T) {
				page, screen, store := setupMetricPage(t)

				payload1, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
				payload1.ResourceMetrics().At(0).Resource().Attributes().PutStr("service.name", "service-1")
				payload1.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).SetName("trace-1")
				payload2, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
				payload2.ResourceMetrics().At(0).Resource().Attributes().PutStr("service.name", "service-2")
				payload2.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).SetName("trace-2")
				payload3, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
				payload3.ResourceMetrics().At(0).Resource().Attributes().PutStr("service.name", "service-3")
				payload3.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).SetName("trace-3")
				store.AddMetric(&payload1)
				store.AddMetric(&payload2)
				store.AddMetric(&payload3)

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
				want := test.LoadTestdata(t, "tui/component/page/metric/metric_table_filter_metrics.txt")

				assert.Equal(t, want, got.String())
			})

			t.Run("change selection", func(t *testing.T) {
				page, screen, store := setupMetricPage(t)

				payload1, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
				payload1.ResourceMetrics().At(0).Resource().Attributes().PutStr("service.name", "service-1")
				payload1.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).SetName("trace-1")
				payload2, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
				payload2.ResourceMetrics().At(0).Resource().Attributes().PutStr("service.name", "service-2")
				payload2.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).SetName("trace-2")
				payload3, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
				payload3.ResourceMetrics().At(0).Resource().Attributes().PutStr("service.name", "service-3")
				payload3.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).SetName("trace-3")
				store.AddMetric(&payload1)
				store.AddMetric(&payload2)
				store.AddMetric(&payload3)

				handler := page.table.view.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone), nil)

				page.view.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/metric/metric_table_change_selection.txt")

				assert.Equal(t, want, got.String())
			})

			t.Run("flush", func(t *testing.T) {
				page, screen, store := setupMetricPage(t)

				payload, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
				store.AddMetric(&payload)

				handler := page.table.view.InputHandler()
				handler(tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone), nil)

				page.view.Draw(screen)
				screen.Sync()

				got := test.GetScreenContent(t, screen)
				want := test.LoadTestdata(t, "tui/component/page/metric/metric_table_flush.txt")

				assert.Equal(t, want, got.String())

				// After flush, when the next span is received, the detail pane renders its content
				newPayload, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
				store.AddMetric(&newPayload)

				page.view.Draw(screen)
				screen.Sync()

				got = test.GetScreenContent(t, screen)
				want = test.LoadTestdata(t, "tui/component/page/metric/metric_table_flush_metric_received.txt")

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
					wantContentPath: "tui/component/page/metric/metric_table_key_handling_divider_left.txt",
				},
				{
					name:            "right",
					key:             tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/metric/metric_table_key_handling_divider_right.txt",
				},
			}

			for _, tt := range tests {
				t.Run("move divider "+tt.name, func(t *testing.T) {
					page, screen, store := setupMetricPage(t)

					payload, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
					store.AddMetric(&payload)

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
			tests := []struct {
				name            string
				key             *tcell.EventKey
				wantContentPath string
			}{
				{
					name:            "left",
					key:             tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/metric/metric_detail_key_handling_divider_left.txt",
				},
				{
					name:            "right",
					key:             tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/metric/metric_detail_key_handling_divider_right.txt",
				},
				{
					name:            "up",
					key:             tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/metric/metric_detail_key_handling_divider_up.txt",
				},
				{
					name:            "down",
					key:             tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/metric/metric_detail_key_handling_divider_down.txt",
				},
			}

			for _, tt := range tests {
				t.Run("move divider "+tt.name, func(t *testing.T) {
					page, screen, store := setupMetricPage(t)

					payload, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
					store.AddMetric(&payload)

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

		t.Run("chart", func(t *testing.T) {
			tests := []struct {
				name            string
				key             *tcell.EventKey
				wantContentPath string
			}{
				{
					name:            "left",
					key:             tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/metric/metric_chart_key_handling_divider_left.txt",
				},
				{
					name:            "right",
					key:             tcell.NewEventKey(tcell.KeyCtrlL, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/metric/metric_chart_key_handling_divider_right.txt",
				},
				{
					name:            "up",
					key:             tcell.NewEventKey(tcell.KeyCtrlK, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/metric/metric_chart_key_handling_divider_up.txt",
				},
				{
					name:            "down",
					key:             tcell.NewEventKey(tcell.KeyCtrlJ, ' ', tcell.ModNone),
					wantContentPath: "tui/component/page/metric/metric_chart_key_handling_divider_down.txt",
				},
			}

			for _, tt := range tests {
				t.Run("move divider "+tt.name, func(t *testing.T) {
					page, screen, store := setupMetricPage(t)

					payload, _ := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
					store.AddMetric(&payload)

					page.table.table.Blur()
					page.chart.view.Focus(func(p tview.Primitive) {
						p.Focus(func(p tview.Primitive) {
							p.Focus(nil)
						})
					})

					handler := page.chart.view.InputHandler()
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
