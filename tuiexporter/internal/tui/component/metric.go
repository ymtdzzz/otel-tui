package component

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

const NULL_VALUE_FLOAT64 = math.MaxFloat64

var defaultMetricCellMappers = cellMappers[telemetry.MetricData]{
	0: {
		header: "Service Name",
		getTextRowFn: func(data *telemetry.MetricData) string {
			return data.GetServiceName()
		},
	},
	1: {
		header: "Metric Name",
		getTextRowFn: func(data *telemetry.MetricData) string {
			return data.GetMetricName()
		},
	},
	2: {
		header: "Metric Type",
		getTextRowFn: func(data *telemetry.MetricData) string {
			return data.GetMetricTypeText()
		},
	},
	3: {
		header: "Data Point Count",
		getTextRowFn: func(data *telemetry.MetricData) string {
			return data.GetDataPointNum()
		},
	},
}

type MetricDataForTable struct {
	tview.TableContentReadOnly
	metrics *[]*telemetry.MetricData
	mapper  cellMappers[telemetry.MetricData]
}

func NewMetricDataForTable(metrics *[]*telemetry.MetricData) MetricDataForTable {
	return MetricDataForTable{
		metrics: metrics,
		mapper:  defaultMetricCellMappers,
	}
}

// implementations for tview Virtual Table
// see: https://github.com/rivo/tview/wiki/VirtualTable
func (m MetricDataForTable) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return m.getHeaderCell(column)
	}
	if row > 0 && row <= len(*m.metrics) {
		return getCellFromData(m.mapper, (*m.metrics)[row-1], column)
	}
	return tview.NewTableCell("N/A")
}

func (m MetricDataForTable) GetRowCount() int {
	return len(*m.metrics) + 1
}

func (m MetricDataForTable) GetColumnCount() int {
	return len(m.mapper)
}

func (m MetricDataForTable) getHeaderCell(column int) *tview.TableCell {
	cell := tview.NewTableCell("N/A").
		SetSelectable(false).
		SetTextColor(tcell.ColorYellow)
	h, ok := m.mapper[column]
	if !ok {
		return cell
	}
	cell.SetText(h.header)

	return cell
}

func getMetricInfoTree(commands *tview.TextView, showModalFn showModalFunc, hideModalFn hideModalFunc, m *telemetry.MetricData) *tview.TreeView {
	if m == nil {
		return nil
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
	appendAttrsSorted(attrs, r.Attributes())
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
	appendAttrsSorted(sattrs, s.Attributes())
	scope.AddChild(sattrs)

	scopes.AddChild(scope)
	resource.AddChild(scopes)

	// metric
	metr := tview.NewTreeNode("Metrics")
	scopes.AddChild(metr)
	/// metadata
	meta := tview.NewTreeNode("Metadata")
	metr.AddChild(meta)
	appendAttrsSorted(meta, m.Metric.Metadata())

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
				appendAttrsSorted(fattrs, e.FilteredAttributes())
			}
			// attributes
			attrs := tview.NewTreeNode("Attributes")
			appendAttrsSorted(attrs, d.Attributes())
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
				appendAttrsSorted(fattrs, e.FilteredAttributes())
			}
			// attributes
			attrs := tview.NewTreeNode("Attributes")
			appendAttrsSorted(attrs, d.Attributes())
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
				appendAttrsSorted(fattrs, e.FilteredAttributes())
			}
			// attributes
			attrs := tview.NewTreeNode("Attributes")
			appendAttrsSorted(attrs, d.Attributes())
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
				appendAttrsSorted(fattrs, e.FilteredAttributes())
			}
			// attributes
			attrs := tview.NewTreeNode("Attributes")
			appendAttrsSorted(attrs, d.Attributes())
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
			appendAttrsSorted(attrs, d.Attributes())
			dp.AddChild(attrs)

			dps.AddChild(dp)
		}
	}

	root.AddChild(resource)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
	})

	attachModalForTreeAttributes(tree, showModalFn, hideModalFn)

	registerCommandList(commands, tree, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModCtrl),
			description: "Reduce the width",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModCtrl),
			description: "Expand the width",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			description: "Toggle folding the child nodes",
		},
	})

	return tree
}

type ByTimestamp []*pmetric.NumberDataPoint

func (a ByTimestamp) Len() int      { return len(a) }
func (a ByTimestamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTimestamp) Less(i, j int) bool {
	return a[i].Timestamp().AsTime().Before(a[j].Timestamp().AsTime())
}

func drawMetricChartByRow(commands *tview.TextView, store *telemetry.Store, row int) tview.Primitive {
	m := store.GetFilteredMetricByIdx(row)
	switch m.Metric.Type() {
	case pmetric.MetricTypeGauge:
		return drawMetricNumberChart(commands, store, m)
	case pmetric.MetricTypeSum:
		return drawMetricNumberChart(commands, store, m)
	case pmetric.MetricTypeHistogram:
		return drawMetricHistogramChart(commands, m)
	case pmetric.MetricTypeExponentialHistogram:
		return drawMetricNumberChart(commands, store, m)
	case pmetric.MetricTypeSummary:
		return drawMetricNumberChart(commands, store, m)
	}
	return nil
}

func drawMetricHistogramChart(commands *tview.TextView, m *telemetry.MetricData) tview.Primitive {
	dpcount := m.Metric.Histogram().DataPoints().Len()
	chs := make([]*tvxwidgets.BarChart, dpcount)
	sides := make([]*tview.Flex, dpcount)
	for dpi := 0; dpi < dpcount; dpi++ {
		dp := m.Metric.Histogram().DataPoints().At(dpi)
		ch := tvxwidgets.NewBarChart()
		ch.SetBorder(true)
		ch.SetTitle(fmt.Sprintf("Data point [%d / %d] ( <- | -> )", dpi+1, dpcount))
		side := tview.NewFlex().SetDirection(tview.FlexRow)
		sts := tview.NewFlex().SetDirection(tview.FlexRow)
		sts.SetBorder(true).SetTitle("Statistics")
		txt := tview.NewFlex().SetDirection(tview.FlexRow)
		txt.SetBorder(true).SetTitle("Attributes")
		for bci := 0; bci < dp.BucketCounts().Len(); bci++ {
			var label string

			if dp.ExplicitBounds().Len() == 0 {
				label = "inf"
			} else {
				switch {
				case bci == 0:
					label = fmt.Sprintf("~%.1f", dp.ExplicitBounds().At(0))
				case bci == dp.BucketCounts().Len()-1:
					label = fmt.Sprintf("%.1f~", dp.ExplicitBounds().At(bci-1))
				default:
					label = fmt.Sprintf("%.1f", dp.ExplicitBounds().At(bci))
				}
			}

			ch.AddBar(label, uint64ToInt(dp.BucketCounts().At(bci)), tcell.ColorYellow)
		}
		sts.AddItem(tview.NewTextView().SetText(fmt.Sprintf("● max: %.1f", dp.Max())), 1, 1, false)
		sts.AddItem(tview.NewTextView().SetText(fmt.Sprintf("● min: %.1f", dp.Min())), 1, 1, false)
		sts.AddItem(tview.NewTextView().SetText(fmt.Sprintf("● sum: %.1f", dp.Sum())), 1, 1, false)
		dp.Attributes().Range(func(k string, v pcommon.Value) bool {
			txt.AddItem(tview.NewTextView().SetText(fmt.Sprintf("● %s: %s", k, v.AsString())), 2, 1, false)
			return true
		})
		side.AddItem(sts, 5, 1, false).AddItem(txt, 0, 1, false)
		chs[dpi] = ch
		sides[dpi] = side
	}

	chart := tview.NewFlex().SetDirection(tview.FlexColumn)
	if dpcount == 0 {
		return chart
	}
	idx := 0
	chart.AddItem(chs[idx], 0, 7, false).AddItem(sides[idx], 0, 3, false)

	chart.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRight:
			if idx < dpcount-1 {
				idx++
			} else {
				idx = 0
			}
			chart.Clear().AddItem(chs[idx], 0, 7, false).AddItem(sides[idx], 0, 3, false)
			return nil
		case tcell.KeyLeft:
			if idx > 0 {
				idx--
			} else {
				idx = dpcount - 1
			}
			chart.Clear().AddItem(chs[idx], 0, 7, false).AddItem(sides[idx], 0, 3, false)
			return nil
		}
		return event
	})

	registerCommandList(commands, chart, nil, KeyMaps{})

	return chart
}

func drawMetricNumberChart(commands *tview.TextView, store *telemetry.Store, m *telemetry.MetricData) tview.Primitive {
	sname := telemetry.GetServiceNameFromResource(m.ResourceMetric.Resource())
	mcache := store.GetMetricCache()
	ms, ok := mcache.GetMetricsBySvcAndMetricName(sname, m.Metric.Name())
	if !ok {
		return nil
	}

	// attribute name and value map
	dataMap := make(map[string]map[string][]*pmetric.NumberDataPoint, 1)
	attrkeys := []string{}

	support := true
	start := time.Unix(1<<63-62135596801, 999999999)
	end := time.Unix(0, 0)
	for _, m := range ms {
		var (
			attrs map[string]any
			dp    pmetric.NumberDataPoint
		)

		switch m.Metric.Type() {
		case pmetric.MetricTypeGauge:
			for dpi := 0; dpi < m.Metric.Gauge().DataPoints().Len(); dpi++ {
				dp = m.Metric.Gauge().DataPoints().At(dpi)
				attrs = dp.Attributes().AsRaw()
				dpts := dp.Timestamp().AsTime()
				if dpts.Before(start) {
					start = dpts
				}
				if dpts.After(end) {
					end = dpts
				}
			}
		case pmetric.MetricTypeSum:
			for dpi := 0; dpi < m.Metric.Sum().DataPoints().Len(); dpi++ {
				dp = m.Metric.Sum().DataPoints().At(dpi)
				attrs = dp.Attributes().AsRaw()
				dpts := dp.Timestamp().AsTime()
				if dpts.Before(start) {
					start = dpts
				}
				if dpts.After(end) {
					end = dpts
				}
			}
		default:
			support = false
		}
		if !support {
			break
		}

		if len(attrs) > 0 {
			// sort keys
			keys := make([]string, 0, len(attrs))
			for k := range attrs {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				v := attrs[k]
				vstr := fmt.Sprintf("%s", v)
				if attrkey, ok := dataMap[k]; ok {
					if _, ok := attrkey[vstr]; ok {
						dataMap[k][vstr] = append(dataMap[k][vstr], &dp)
					} else {
						dataMap[k][vstr] = []*pmetric.NumberDataPoint{&dp}
					}
				} else {
					attrkeys = append(attrkeys, k)
					dataMap[k] = map[string][]*pmetric.NumberDataPoint{vstr: {&dp}}
				}
			}
		} else {
			k := "N/A"
			vstr := "N/A"
			if attrkey, ok := dataMap[k]; ok {
				if _, ok := attrkey[vstr]; ok {
					dataMap[k][vstr] = append(dataMap[k][vstr], &dp)
				} else {
					dataMap[k][vstr] = []*pmetric.NumberDataPoint{&dp}
				}
			} else {
				attrkeys = append(attrkeys, k)
				dataMap[k] = map[string][]*pmetric.NumberDataPoint{vstr: {&dp}}
			}
		}
	}

	chart := tview.NewFlex().SetDirection(tview.FlexColumn)

	// TODO: Delete it after implementing drawMetric* for all types
	if !support {
		txt := tview.NewTextView().SetText("This metric type is not supported")
		chart.AddItem(txt, 0, 1, false)
		return chart
	}

	for k := range dataMap {
		for kk := range dataMap[k] {
			sort.Sort(ByTimestamp(dataMap[k][kk]))
		}
	}

	getTitle := func(idx int) string {
		return fmt.Sprintf("%s [%d / %d] ( <- | -> )", attrkeys[idx], idx+1, len(attrkeys))
	}

	// Draw a chart of the first attribute
	attrkeyidx := 0
	data, txts := getDataToDraw(dataMap, attrkeys[attrkeyidx], start, end)
	ch := tvxwidgets.NewPlot()
	ch.SetMarker(tvxwidgets.PlotMarkerBraille)
	ch.SetTitle(getTitle(attrkeyidx))
	ch.SetBorder(true)
	ch.SetData(data)
	ch.SetDrawXAxisLabel(false)
	ch.SetLineColor(colors)

	legend := tview.NewFlex().SetDirection(tview.FlexRow)
	legend.AddItem(txts, 0, 1, false)

	ch.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRight:
			if attrkeyidx < len(attrkeys)-1 {
				attrkeyidx++
			} else {
				attrkeyidx = 0
			}
			ch.SetTitle(getTitle(attrkeyidx))
			data, txts := getDataToDraw(dataMap, attrkeys[attrkeyidx], start, end)
			legend.Clear()
			legend.AddItem(txts, 0, 1, false)
			ch.SetData(data)
			return nil
		case tcell.KeyLeft:
			if attrkeyidx > 0 {
				attrkeyidx--
			} else {
				attrkeyidx = len(attrkeys) - 1
			}
			ch.SetTitle(getTitle(attrkeyidx))
			data, txts := getDataToDraw(dataMap, attrkeys[attrkeyidx], start, end)
			legend.Clear()
			legend.AddItem(txts, 0, 1, false)
			ch.SetData(data)
			return nil
		}
		return event
	})

	chart.AddItem(ch, 0, 7, true).AddItem(legend, 0, 3, false)

	registerCommandList(commands, ch, nil, KeyMaps{})

	return chart
}

func getDataToDraw(dataMap map[string]map[string][]*pmetric.NumberDataPoint, attrkey string, start, end time.Time) ([][]float64, *tview.TextView) {
	// Sort keys
	keys := make([]string, 0, len(dataMap[attrkey]))
	for k := range dataMap[attrkey] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// Count datapoints
	dpnum := 0
	for _, k := range keys {
		dpnum += len(dataMap[attrkey][k])
	}
	d := make([][]float64, len(keys))
	for i := range d {
		d[i] = make([]float64, dpnum)
	}
	// Set null value
	for i := range d {
		for ii := range d[i] {
			d[i][ii] = NULL_VALUE_FLOAT64
		}
	}
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	wholedur := end.Sub(start).Nanoseconds()
	type locateMap struct {
		prevpos int
		prevval float64
		pos     int
		val     float64
	}
	locatedposmap := make(map[int][]locateMap, len(keys))
	// Set values to timestamp relative position.
	// Note that this process keeps values between corresponding positions null value.
	// ex: [1.2 1.3 null 1.6 1.1 null null 2.5]
	txts := make([]string, len(keys))
	for i, k := range keys {
		prevpos := -1
		prevval := NULL_VALUE_FLOAT64
		for _, dp := range dataMap[attrkey][k] {
			// Get timestamp and locate it to relative position
			dur := dp.Timestamp().AsTime().Sub(start).Nanoseconds()
			var ratio float64
			if dur == 0 {
				ratio = 0
			} else {
				ratio = float64(dur) / float64(wholedur)
			}
			pos := int(math.Round(float64(dpnum) * ratio))
			if pos >= len(d[i]) {
				pos = len(d[i]) - 1
			}
			if pos < 0 {
				pos = 0
			}
			var val float64
			switch dp.ValueType() {
			case pmetric.NumberDataPointValueTypeDouble:
				val = dp.DoubleValue()
			case pmetric.NumberDataPointValueTypeInt:
				val = float64(dp.IntValue())
			}
			d[i][pos] = val
			locatedposmap[i] = append(locatedposmap[i], locateMap{
				prevpos: prevpos,
				prevval: prevval,
				pos:     pos,
				val:     val,
			})
			prevpos = pos
			prevval = val
		}
		txts[i] = fmt.Sprintf("[%s]● %s: %s", colors[i].String(), attrkey, k)
	}
	tv.SetText(strings.Join(txts, "\n"))
	// Replace null value with appropriate value for smooth line
	// ex: [1.2 1.3 1.45 1.6 1.1 1.56 2.02 2.5]
	for i := range d {
		for c, pmap := range locatedposmap[i] {
			// Fill after the last element
			if c == len(locatedposmap[i])-1 && pmap.pos < dpnum {
				for j := pmap.pos + 1; j < dpnum; j++ {
					d[i][j] = pmap.val
				}
			}
			// Fill before the first element
			if pmap.prevpos == -1 {
				for j := 0; j < pmap.pos; j++ {
					d[i][j] = pmap.val
				}
				continue
			}
			split := pmap.pos - pmap.prevpos
			diff := pmap.val - pmap.prevval
			step := diff / float64(split+1)
			curr := pmap.prevval
			for j := pmap.prevpos + 1; j < pmap.pos; j++ {
				curr += step
				d[i][j] = curr
			}
		}
	}
	return d, tv
}

// uint64ToInt converts uint64 into int. When the input is larger than math.MaxInt, it returns math.MaxInt.
func uint64ToInt(u uint64) int {
	if u >= math.MaxInt {
		return math.MaxInt
	}
	return int(u)
}
