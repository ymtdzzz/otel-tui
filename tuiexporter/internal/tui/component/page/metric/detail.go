package metric

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type detail struct {
	commands       *tview.TextView
	view           *tview.Flex
	tree           *tview.TreeView
	resizeManagers []*layout.ResizeManager
}

func newDetail(
	commands *tview.TextView,
	resizeManagers []*layout.ResizeManager,
) *detail {
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Details (d)").SetBorder(true)

	detail := &detail{
		commands:       commands,
		view:           container,
		resizeManagers: resizeManagers,
	}

	detail.update(nil)

	return detail
}

func (d *detail) flush() {
	d.tree.SetRoot(nil)
}

func (d *detail) update(m *telemetry.MetricData) {
	hasFocus := d.view.HasFocus()
	d.view.Clear()
	d.tree = d.getMetricInfoTree(m)
	d.updateCommands()
	d.view.AddItem(d.tree, 0, 1, true)
	if hasFocus {
		navigation.Focus(d.view)
	}
}

func (d *detail) getMetricInfoTree(m *telemetry.MetricData) *tview.TreeView {
	if m == nil {
		return tview.NewTreeView()
	}
	root := tview.NewTreeNode("Metric")
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)

	mname := tview.NewTreeNode(fmt.Sprintf("name: %s", m.Metric.Name()))
	munit := tview.NewTreeNode(fmt.Sprintf("unit: %s", m.Metric.Unit()))
	mdesc := tview.NewTreeNode(fmt.Sprintf("description: %s", m.Metric.Description()))
	mtype := tview.NewTreeNode(fmt.Sprintf("type: %s", m.Metric.Type().String()))

	root.AddChild(mname)
	root.AddChild(munit)
	root.AddChild(mdesc)
	root.AddChild(mtype)

	// resource info
	rm := m.ResourceMetric
	r := rm.Resource()
	resource := tview.NewTreeNode("Resource")
	rdropped := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", r.DroppedAttributesCount()))
	resource.AddChild(rdropped)
	rschema := tview.NewTreeNode(fmt.Sprintf("schema url: %s", rm.SchemaUrl()))
	resource.AddChild(rschema)

	attrs := tview.NewTreeNode("Attributes")
	layout.AppendAttrsSorted(attrs, r.Attributes())
	resource.AddChild(attrs)

	// scope info
	scopes := tview.NewTreeNode("Scopes")
	sm := m.ScopeMetric
	s := sm.Scope()
	scope := tview.NewTreeNode(s.Name())
	sschema := tview.NewTreeNode(fmt.Sprintf("schema url: %s", sm.SchemaUrl()))
	scope.AddChild(sschema)

	scope.AddChild(tview.NewTreeNode(fmt.Sprintf("version: %s", s.Version())))
	scope.AddChild(tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", s.DroppedAttributesCount())))

	sattrs := tview.NewTreeNode("Attributes")
	layout.AppendAttrsSorted(sattrs, s.Attributes())
	scope.AddChild(sattrs)

	scopes.AddChild(scope)
	resource.AddChild(scopes)

	// metric
	metr := tview.NewTreeNode("Metrics")
	scopes.AddChild(metr)
	/// metadata
	meta := tview.NewTreeNode("Metadata")
	metr.AddChild(meta)
	layout.AppendAttrsSorted(meta, m.Metric.Metadata())

	/// datapoints
	dps := tview.NewTreeNode("Datapoints")
	metr.AddChild(dps)
	switch m.Metric.Type() {
	case pmetric.MetricTypeGauge:
		for dpi := 0; dpi < m.Metric.Gauge().DataPoints().Len(); dpi++ {
			dps.AddChild(numberDatapointNode(dpi, m.Metric.Gauge().DataPoints().At(dpi)))
		}
	case pmetric.MetricTypeSum:
		sum := m.Metric.Sum()
		root.AddChild(temporalityNode(sum.AggregationTemporality()))
		root.AddChild(tview.NewTreeNode(fmt.Sprintf("is monotonic: %v", sum.IsMonotonic())))
		for dpi := 0; dpi < sum.DataPoints().Len(); dpi++ {
			dps.AddChild(numberDatapointNode(dpi, sum.DataPoints().At(dpi)))
		}
	case pmetric.MetricTypeHistogram:
		hist := m.Metric.Histogram()
		root.AddChild(temporalityNode(hist.AggregationTemporality()))
		for dpi := 0; dpi < hist.DataPoints().Len(); dpi++ {
			dp := tview.NewTreeNode(fmt.Sprintf("%d", dpi))
			d := hist.DataPoints().At(dpi)
			addTimestampNodes(dp, d.StartTimestamp(), d.Timestamp())
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("count: %d", d.Count())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("bucket counts (%d): %v", d.BucketCounts().Len(), d.BucketCounts().AsRaw())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("explicit bounds (%d): %v", d.ExplicitBounds().Len(), d.ExplicitBounds().AsRaw())))
			dp.AddChild(optionalDoubleNode("max", d.HasMax(), d.Max()))
			dp.AddChild(optionalDoubleNode("min", d.HasMin(), d.Min()))
			dp.AddChild(optionalDoubleNode("sum", d.HasSum(), d.Sum()))
			dp.AddChild(flagsNode(d.Flags()))
			dp.AddChild(exemplarsNode(d.Exemplars()))
			dp.AddChild(attributesNode(d.Attributes()))
			dps.AddChild(dp)
		}
	case pmetric.MetricTypeExponentialHistogram:
		hist := m.Metric.ExponentialHistogram()
		root.AddChild(temporalityNode(hist.AggregationTemporality()))
		for dpi := 0; dpi < hist.DataPoints().Len(); dpi++ {
			dp := tview.NewTreeNode(fmt.Sprintf("%d", dpi))
			d := hist.DataPoints().At(dpi)
			addTimestampNodes(dp, d.StartTimestamp(), d.Timestamp())
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("count: %d", d.Count())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("scale: %d", d.Scale())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("zero count: %d", d.ZeroCount())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("zero threshold: %f", d.ZeroThreshold())))
			neg := tview.NewTreeNode("Negative")
			dp.AddChild(neg)
			neg.AddChild(tview.NewTreeNode(fmt.Sprintf("bucket counts: %v", d.Negative().BucketCounts().AsRaw())))
			neg.AddChild(tview.NewTreeNode(fmt.Sprintf("offset: %d", d.Negative().Offset())))
			pos := tview.NewTreeNode("Positive")
			dp.AddChild(pos)
			pos.AddChild(tview.NewTreeNode(fmt.Sprintf("bucket counts: %v", d.Positive().BucketCounts().AsRaw())))
			pos.AddChild(tview.NewTreeNode(fmt.Sprintf("offset: %d", d.Positive().Offset())))
			dp.AddChild(optionalDoubleNode("max", d.HasMax(), d.Max()))
			dp.AddChild(optionalDoubleNode("min", d.HasMin(), d.Min()))
			dp.AddChild(optionalDoubleNode("sum", d.HasSum(), d.Sum()))
			dp.AddChild(flagsNode(d.Flags()))
			dp.AddChild(exemplarsNode(d.Exemplars()))
			dp.AddChild(attributesNode(d.Attributes()))
			dps.AddChild(dp)
		}
	case pmetric.MetricTypeSummary:
		for dpi := 0; dpi < m.Metric.Summary().DataPoints().Len(); dpi++ {
			dp := tview.NewTreeNode(fmt.Sprintf("%d", dpi))
			d := m.Metric.Summary().DataPoints().At(dpi)
			addTimestampNodes(dp, d.StartTimestamp(), d.Timestamp())
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("count: %d", d.Count())))
			// quantile
			quants := tview.NewTreeNode("Quantile Values")
			dp.AddChild(quants)
			for qi := 0; qi < d.QuantileValues().Len(); qi++ {
				q := d.QuantileValues().At(qi)
				quant := tview.NewTreeNode(fmt.Sprintf("%d", qi))
				quants.AddChild(quant)
				quant.AddChild(tview.NewTreeNode(fmt.Sprintf("quantile: %f", q.Quantile())))
				quant.AddChild(tview.NewTreeNode(fmt.Sprintf("value: %f", q.Value())))
			}
			dp.AddChild(flagsNode(d.Flags()))
			dp.AddChild(attributesNode(d.Attributes()))
			dps.AddChild(dp)
		}
	}

	root.AddChild(resource)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
	})

	layout.AttachModalForTreeAttributes(tree, d.updateCommands)

	return tree
}

// numberDatapointNode renders a Gauge or Sum datapoint.
func numberDatapointNode(index int, d pmetric.NumberDataPoint) *tview.TreeNode {
	dp := tview.NewTreeNode(fmt.Sprintf("%d", index))
	addTimestampNodes(dp, d.StartTimestamp(), d.Timestamp())
	val := tview.NewTreeNode("Value")
	val.AddChild(tview.NewTreeNode(fmt.Sprintf("type: %s", d.ValueType().String())))
	val.AddChild(tview.NewTreeNode(fmt.Sprintf("int: %d", d.IntValue())))
	val.AddChild(tview.NewTreeNode(fmt.Sprintf("double: %f", d.DoubleValue())))
	dp.AddChild(val)
	dp.AddChild(flagsNode(d.Flags()))
	dp.AddChild(exemplarsNode(d.Exemplars()))
	dp.AddChild(attributesNode(d.Attributes()))
	return dp
}

func addTimestampNodes(dp *tview.TreeNode, start, ts pcommon.Timestamp) {
	dp.AddChild(tview.NewTreeNode(fmt.Sprintf("start timestamp: %s", start.String())))
	dp.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", ts.String())))
}

func temporalityNode(t pmetric.AggregationTemporality) *tview.TreeNode {
	return tview.NewTreeNode(fmt.Sprintf("aggregation temporality: %s", t.String()))
}

// optionalDoubleNode renders a field that the data model marks as optional,
// so that an unset value is not mistaken for a recorded zero.
func optionalDoubleNode(name string, isSet bool, value float64) *tview.TreeNode {
	if !isSet {
		return tview.NewTreeNode(fmt.Sprintf("%s: (not set)", name))
	}
	return tview.NewTreeNode(fmt.Sprintf("%s: %f", name, value))
}

func flagsNode(flags pmetric.DataPointFlags) *tview.TreeNode {
	flg := tview.NewTreeNode("Flags")
	flg.AddChild(tview.NewTreeNode(fmt.Sprintf("no recorded value: %v", flags.NoRecordedValue())))
	return flg
}

func exemplarsNode(exemplars pmetric.ExemplarSlice) *tview.TreeNode {
	exs := tview.NewTreeNode("Examplers")
	for ei := 0; ei < exemplars.Len(); ei++ {
		ex := tview.NewTreeNode(fmt.Sprintf("%d", ei))
		exs.AddChild(ex)
		e := exemplars.At(ei)
		ex.AddChild(tview.NewTreeNode(fmt.Sprintf("trace id: %s", e.TraceID())))
		ex.AddChild(tview.NewTreeNode(fmt.Sprintf("span id: %s", e.SpanID())))
		ex.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", e.Timestamp().String())))
		v := tview.NewTreeNode("Value")
		v.AddChild(tview.NewTreeNode(fmt.Sprintf("type: %s", e.ValueType().String())))
		v.AddChild(tview.NewTreeNode(fmt.Sprintf("int: %d", e.IntValue())))
		v.AddChild(tview.NewTreeNode(fmt.Sprintf("double: %f", e.DoubleValue())))
		ex.AddChild(v)
		fattrs := tview.NewTreeNode("Filtered Attributes")
		ex.AddChild(fattrs)
		layout.AppendAttrsSorted(fattrs, e.FilteredAttributes())
	}
	return exs
}

func attributesNode(attributes pcommon.Map) *tview.TreeNode {
	attrs := tview.NewTreeNode("Attributes")
	layout.AppendAttrsSorted(attrs, attributes)
	return attrs
}

func (d *detail) updateCommands() {
	keyMaps := layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			Description: "Toggle folding (parent), Show full text (child)",
		},
	}
	for _, rm := range d.resizeManagers {
		keyMaps.Merge(rm.KeyMaps())
	}
	layout.RegisterCommandList(d.commands, d.tree, nil, keyMaps)
}
