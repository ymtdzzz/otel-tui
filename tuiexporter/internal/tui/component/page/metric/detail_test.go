package metric

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/test"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestInputCaptureAfterModalClosed(t *testing.T) {
	_, testdata := test.GenerateOTLPGaugeMetricsPayload(t, 1, []int{1}, [][]int{{1}})
	metrics := make([]*telemetry.MetricData, 0, 1)
	metrics = append(metrics, &telemetry.MetricData{
		Metric:         testdata.Metrics[0],
		ResourceMetric: testdata.RMetrics[0],
		ScopeMetric:    testdata.SMetrics[0],
	})

	detail := newDetail(layout.NewCommandList(), []*layout.ResizeManager{
		layout.NewResizeManager(layout.ResizeDirectionHorizontal),
	})
	detail.update(metrics[0])

	handler := detail.tree.InputHandler()

	// open modal
	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), nil)

	// close modal
	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil)

	got := detail.tree.GetInputCapture()(tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone))

	// resize key should be captured
	assert.Nil(t, got)
}

func newTestMetricData(m pmetric.Metric) *telemetry.MetricData {
	rm := pmetric.NewResourceMetrics()
	sm := rm.ScopeMetrics().AppendEmpty()
	return &telemetry.MetricData{
		Metric:         &m,
		ResourceMetric: &rm,
		ScopeMetric:    &sm,
	}
}

// datapointTexts returns the texts of the nodes directly under the first
// datapoint, so a test can check the fields shown for it.
func datapointTexts(t *testing.T, tree *tview.TreeView) []string {
	t.Helper()
	var dps *tview.TreeNode
	tree.GetRoot().Walk(func(node, _ *tview.TreeNode) bool {
		if node.GetText() == "Datapoints" {
			dps = node
			return false
		}
		return true
	})
	if dps == nil || len(dps.GetChildren()) == 0 {
		t.Fatal("no datapoint in the tree")
	}
	children := dps.GetChildren()[0].GetChildren()
	texts := make([]string, 0, len(children))
	for _, c := range children {
		texts = append(texts, c.GetText())
	}
	return texts
}

func rootTexts(tree *tview.TreeView) []string {
	children := tree.GetRoot().GetChildren()
	texts := make([]string, 0, len(children))
	for _, c := range children {
		texts = append(texts, c.GetText())
	}
	return texts
}

func newDetailForTest() *detail {
	return newDetail(layout.NewCommandList(), []*layout.ResizeManager{
		layout.NewResizeManager(layout.ResizeDirectionHorizontal),
	})
}

func TestDetailSumShowsTemporalityAndMonotonicity(t *testing.T) {
	m := pmetric.NewMetric()
	m.SetName("requests")
	sum := m.SetEmptySum()
	sum.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
	sum.SetIsMonotonic(true)
	sum.DataPoints().AppendEmpty().SetIntValue(3)

	tree := newDetailForTest().getMetricInfoTree(newTestMetricData(m))

	got := rootTexts(tree)
	assert.Contains(t, got, "aggregation temporality: Delta")
	assert.Contains(t, got, "is monotonic: true")
}

func TestDetailHistogramOptionalFields(t *testing.T) {
	t.Run("unset min, max and sum are shown as not set", func(t *testing.T) {
		m := pmetric.NewMetric()
		h := m.SetEmptyHistogram()
		h.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		h.DataPoints().AppendEmpty().SetCount(2)

		tree := newDetailForTest().getMetricInfoTree(newTestMetricData(m))

		assert.Contains(t, rootTexts(tree), "aggregation temporality: Cumulative")
		got := datapointTexts(t, tree)
		assert.Contains(t, got, "max: (not set)")
		assert.Contains(t, got, "min: (not set)")
		assert.Contains(t, got, "sum: (not set)")
	})

	t.Run("a min, max or sum of zero is shown as a value", func(t *testing.T) {
		m := pmetric.NewMetric()
		dp := m.SetEmptyHistogram().DataPoints().AppendEmpty()
		dp.SetMin(0)
		dp.SetMax(1.5)
		dp.SetSum(0)

		got := datapointTexts(t, newDetailForTest().getMetricInfoTree(newTestMetricData(m)))

		assert.Contains(t, got, "max: 1.500000")
		assert.Contains(t, got, "min: 0.000000")
		assert.Contains(t, got, "sum: 0.000000")
	})
}

func TestDetailExponentialHistogramFields(t *testing.T) {
	t.Run("unset min, max and sum are shown as not set", func(t *testing.T) {
		m := pmetric.NewMetric()
		eh := m.SetEmptyExponentialHistogram()
		eh.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
		dp := eh.DataPoints().AppendEmpty()
		dp.SetZeroCount(4)
		dp.SetZeroThreshold(0.001)

		tree := newDetailForTest().getMetricInfoTree(newTestMetricData(m))

		assert.Contains(t, rootTexts(tree), "aggregation temporality: Delta")
		got := datapointTexts(t, tree)
		assert.Contains(t, got, "zero count: 4")
		assert.Contains(t, got, "zero threshold: 0.001000")
		assert.Contains(t, got, "max: (not set)")
		assert.Contains(t, got, "min: (not set)")
		assert.Contains(t, got, "sum: (not set)")
	})

	t.Run("set min, max and sum are shown as values", func(t *testing.T) {
		m := pmetric.NewMetric()
		dp := m.SetEmptyExponentialHistogram().DataPoints().AppendEmpty()
		dp.SetMin(-2)
		dp.SetMax(0)
		dp.SetSum(7.25)

		got := datapointTexts(t, newDetailForTest().getMetricInfoTree(newTestMetricData(m)))

		assert.Contains(t, got, "max: 0.000000")
		assert.Contains(t, got, "min: -2.000000")
		assert.Contains(t, got, "sum: 7.250000")
	})
}

func TestDetailSummaryWithoutQuantiles(t *testing.T) {
	m := pmetric.NewMetric()
	m.SetEmptySummary().DataPoints().AppendEmpty().SetCount(1)

	d := newDetailForTest()

	assert.NotPanics(t, func() {
		d.update(newTestMetricData(m))
	})
	assert.Contains(t, datapointTexts(t, d.tree), "Quantile Values")
}
