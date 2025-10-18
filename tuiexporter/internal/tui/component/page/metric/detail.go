package metric

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type detail struct {
	commands       *tview.TextView
	view           *tview.Flex
	tree           *tview.TreeView
	showModalFn    layout.ShowModalFunc
	hideModalFn    layout.HideModalFunc
	resizeManagers []*layout.ResizeManager
}

func newDetail(
	commands *tview.TextView,
	showModalFn layout.ShowModalFunc,
	hideModalFn layout.HideModalFunc,
	resizeManagers []*layout.ResizeManager,
) *detail {
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Details (d)").SetBorder(true)

	detail := &detail{
		commands:       commands,
		view:           container,
		showModalFn:    showModalFn,
		hideModalFn:    hideModalFn,
		resizeManagers: resizeManagers,
	}

	detail.update(nil)

	return detail
}

func (d *detail) flush() {
	d.view.Clear()
}

func (d *detail) update(m *telemetry.MetricData) {
	d.view.Clear()
	d.tree = d.getMetricInfoTree(m)
	d.updateCommands()
	d.view.AddItem(d.tree, 0, 1, true)
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
			dp := tview.NewTreeNode(fmt.Sprintf("%d", dpi))
			d := m.Metric.Gauge().DataPoints().At(dpi)
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("start timestamp: %s", d.StartTimestamp().String())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", d.Timestamp().String())))
			// value
			val := tview.NewTreeNode("Value")
			val.AddChild(tview.NewTreeNode(fmt.Sprintf("type: %s", d.ValueType().String())))
			val.AddChild(tview.NewTreeNode(fmt.Sprintf("int: %d", d.IntValue())))
			val.AddChild(tview.NewTreeNode(fmt.Sprintf("double: %f", d.DoubleValue())))
			dp.AddChild(val)
			// flags
			flg := tview.NewTreeNode("Flags")
			flg.AddChild(tview.NewTreeNode(fmt.Sprintf("no recorded value: %v", d.Flags().NoRecordedValue())))
			dp.AddChild(flg)
			// exampler
			exs := tview.NewTreeNode("Examplers")
			dp.AddChild(exs)
			for ei := 0; ei < d.Exemplars().Len(); ei++ {
				ex := tview.NewTreeNode(fmt.Sprintf("%d", ei))
				exs.AddChild(ex)
				e := d.Exemplars().At(ei)
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("trace id: %s", e.TraceID())))
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("span id: %s", e.SpanID())))
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", e.Timestamp().String())))
				// value
				v := tview.NewTreeNode("Value")
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("type: %s", e.ValueType().String())))
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("int: %d", e.IntValue())))
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("double: %f", e.DoubleValue())))
				ex.AddChild(v)
				// filtered attributes
				fattrs := tview.NewTreeNode("Filtered Attributes")
				ex.AddChild(fattrs)
				layout.AppendAttrsSorted(fattrs, e.FilteredAttributes())
			}
			// attributes
			attrs := tview.NewTreeNode("Attributes")
			layout.AppendAttrsSorted(attrs, d.Attributes())
			dp.AddChild(attrs)

			dps.AddChild(dp)
		}
	case pmetric.MetricTypeSum:
		for dpi := 0; dpi < m.Metric.Sum().DataPoints().Len(); dpi++ {
			dp := tview.NewTreeNode(fmt.Sprintf("%d", dpi))
			d := m.Metric.Sum().DataPoints().At(dpi)
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("start timestamp: %s", d.StartTimestamp().String())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", d.Timestamp().String())))
			// value
			val := tview.NewTreeNode("Value")
			val.AddChild(tview.NewTreeNode(fmt.Sprintf("type: %s", d.ValueType().String())))
			val.AddChild(tview.NewTreeNode(fmt.Sprintf("int: %d", d.IntValue())))
			val.AddChild(tview.NewTreeNode(fmt.Sprintf("double: %f", d.DoubleValue())))
			dp.AddChild(val)
			// flags
			flg := tview.NewTreeNode("Flags")
			flg.AddChild(tview.NewTreeNode(fmt.Sprintf("no recorded value: %v", d.Flags().NoRecordedValue())))
			dp.AddChild(flg)
			// exampler
			exs := tview.NewTreeNode("Examplers")
			dp.AddChild(exs)
			for ei := 0; ei < d.Exemplars().Len(); ei++ {
				ex := tview.NewTreeNode(fmt.Sprintf("%d", ei))
				exs.AddChild(ex)
				e := d.Exemplars().At(ei)
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("trace id: %s", e.TraceID())))
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("span id: %s", e.SpanID())))
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", e.Timestamp().String())))
				// value
				v := tview.NewTreeNode("Value")
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("type: %s", e.ValueType().String())))
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("int: %d", e.IntValue())))
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("double: %f", e.DoubleValue())))
				ex.AddChild(v)
				// filtered attributes
				fattrs := tview.NewTreeNode("Filtered Attributes")
				ex.AddChild(fattrs)
				layout.AppendAttrsSorted(fattrs, e.FilteredAttributes())
			}
			// attributes
			attrs := tview.NewTreeNode("Attributes")
			layout.AppendAttrsSorted(attrs, d.Attributes())
			dp.AddChild(attrs)

			dps.AddChild(dp)
		}
	case pmetric.MetricTypeHistogram:
		for dpi := 0; dpi < m.Metric.Histogram().DataPoints().Len(); dpi++ {
			dp := tview.NewTreeNode(fmt.Sprintf("%d", dpi))
			d := m.Metric.Histogram().DataPoints().At(dpi)
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("start timestamp: %s", d.StartTimestamp().String())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", d.Timestamp().String())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("count: %d", d.Count())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("bucket counts (%d): %v", d.BucketCounts().Len(), d.BucketCounts().AsRaw())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("explicit bounds (%d): %v", d.ExplicitBounds().Len(), d.ExplicitBounds().AsRaw())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("max: %f", d.Max())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("min: %f", d.Min())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("sum: %f", d.Sum())))
			// flags
			flg := tview.NewTreeNode("Flags")
			flg.AddChild(tview.NewTreeNode(fmt.Sprintf("no recorded value: %v", d.Flags().NoRecordedValue())))
			dp.AddChild(flg)
			// exampler
			exs := tview.NewTreeNode("Examplers")
			dp.AddChild(exs)
			for ei := 0; ei < d.Exemplars().Len(); ei++ {
				ex := tview.NewTreeNode(fmt.Sprintf("%d", ei))
				exs.AddChild(ex)
				e := d.Exemplars().At(ei)
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("trace id: %s", e.TraceID())))
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("span id: %s", e.SpanID())))
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", e.Timestamp().String())))
				// value
				v := tview.NewTreeNode("Value")
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("type: %s", e.ValueType().String())))
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("int: %d", e.IntValue())))
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("double: %f", e.DoubleValue())))
				ex.AddChild(v)
				// filtered attributes
				fattrs := tview.NewTreeNode("Filtered Attributes")
				ex.AddChild(fattrs)
				layout.AppendAttrsSorted(fattrs, e.FilteredAttributes())
			}
			// attributes
			attrs := tview.NewTreeNode("Attributes")
			layout.AppendAttrsSorted(attrs, d.Attributes())
			dp.AddChild(attrs)

			dps.AddChild(dp)
		}
	case pmetric.MetricTypeExponentialHistogram:
		for dpi := 0; dpi < m.Metric.ExponentialHistogram().DataPoints().Len(); dpi++ {
			dp := tview.NewTreeNode(fmt.Sprintf("%d", dpi))
			d := m.Metric.ExponentialHistogram().DataPoints().At(dpi)
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("start timestamp: %s", d.StartTimestamp().String())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", d.Timestamp().String())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("count: %d", d.Count())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("scale: %d", d.Scale())))
			neg := tview.NewTreeNode("Negative")
			dp.AddChild(neg)
			neg.AddChild(tview.NewTreeNode(fmt.Sprintf("bucket counts: %v", d.Negative().BucketCounts().AsRaw())))
			neg.AddChild(tview.NewTreeNode(fmt.Sprintf("offset: %d", d.Negative().Offset())))
			pos := tview.NewTreeNode("Positive")
			dp.AddChild(pos)
			pos.AddChild(tview.NewTreeNode(fmt.Sprintf("bucket counts: %v", d.Positive().BucketCounts().AsRaw())))
			pos.AddChild(tview.NewTreeNode(fmt.Sprintf("offset: %d", d.Positive().Offset())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("max: %f", d.Max())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("min: %f", d.Min())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("sum: %f", d.Sum())))
			// flags
			flg := tview.NewTreeNode("Flags")
			flg.AddChild(tview.NewTreeNode(fmt.Sprintf("no recorded value: %v", d.Flags().NoRecordedValue())))
			dp.AddChild(flg)
			// exampler
			exs := tview.NewTreeNode("Examplers")
			dp.AddChild(exs)
			for ei := 0; ei < d.Exemplars().Len(); ei++ {
				ex := tview.NewTreeNode(fmt.Sprintf("%d", ei))
				exs.AddChild(ex)
				e := d.Exemplars().At(ei)
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("trace id: %s", e.TraceID())))
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("span id: %s", e.SpanID())))
				ex.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", e.Timestamp().String())))
				// value
				v := tview.NewTreeNode("Value")
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("type: %s", e.ValueType().String())))
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("int: %d", e.IntValue())))
				v.AddChild(tview.NewTreeNode(fmt.Sprintf("double: %f", e.DoubleValue())))
				ex.AddChild(v)
				// filtered attributes
				fattrs := tview.NewTreeNode("Filtered Attributes")
				ex.AddChild(fattrs)
				layout.AppendAttrsSorted(fattrs, e.FilteredAttributes())
			}
			// attributes
			attrs := tview.NewTreeNode("Attributes")
			layout.AppendAttrsSorted(attrs, d.Attributes())
			dp.AddChild(attrs)

			dps.AddChild(dp)
		}
	case pmetric.MetricTypeSummary:
		for dpi := 0; dpi < m.Metric.Summary().DataPoints().Len(); dpi++ {
			dp := tview.NewTreeNode(fmt.Sprintf("%d", dpi))
			d := m.Metric.Summary().DataPoints().At(dpi)
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("start timestamp: %s", d.StartTimestamp().String())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("timestamp: %s", d.Timestamp().String())))
			dp.AddChild(tview.NewTreeNode(fmt.Sprintf("count: %d", d.Count())))
			d.QuantileValues().At(0).Quantile()
			d.QuantileValues().At(0).Value()
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
			// flags
			flg := tview.NewTreeNode("Flags")
			flg.AddChild(tview.NewTreeNode(fmt.Sprintf("no recorded value: %v", d.Flags().NoRecordedValue())))
			dp.AddChild(flg)
			// attributes
			attrs := tview.NewTreeNode("Attributes")
			layout.AppendAttrsSorted(attrs, d.Attributes())
			dp.AddChild(attrs)

			dps.AddChild(dp)
		}
	}

	root.AddChild(resource)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
	})

	layout.AttachModalForTreeAttributes(tree, d.showModalFn, d.hideModalFn)

	return tree
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
	layout.RegisterCommandList2(d.commands, d.tree, nil, keyMaps)
}
