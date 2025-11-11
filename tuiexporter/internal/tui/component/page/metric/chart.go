package metric

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
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

const nullValueFloat64 = math.MaxFloat64

type chart struct {
	commands       *tview.TextView
	view           *tview.Flex
	ch             *tview.Flex
	focusTargets   []layout.FocusableBox
	store          *telemetry.Store
	resizeManagers []*layout.ResizeManager
}

func newChart(
	commands *tview.TextView,
	store *telemetry.Store,
	resizeManagers []*layout.ResizeManager,
) *chart {
	container := tview.NewFlex().SetDirection(tview.FlexRow)
	container.SetTitle("Chart (c)").SetBorder(true)
	ch := tview.NewFlex().SetDirection(tview.FlexColumn)

	container.AddItem(ch, 0, 1, true)

	c := &chart{
		commands:       commands,
		view:           container,
		ch:             ch,
		focusTargets:   []layout.FocusableBox{ch},
		store:          store,
		resizeManagers: resizeManagers,
	}

	c.update(nil)

	return c
}

func (c *chart) flush() {
	c.ch.Clear()
}

func (c *chart) update(m *telemetry.MetricData) {
	c.ch.Clear()
	keyMaps := c.drawMetricChartByRow(m)
	c.updateCommands(keyMaps)
}

type ByTimestamp []*pmetric.NumberDataPoint

func (a ByTimestamp) Len() int      { return len(a) }
func (a ByTimestamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTimestamp) Less(i, j int) bool {
	return a[i].Timestamp().AsTime().Before(a[j].Timestamp().AsTime())
}

func (c *chart) drawMetricChartByRow(m *telemetry.MetricData) layout.KeyMaps {
	if m == nil {
		return layout.KeyMaps{}
	}

	switch m.Metric.Type() {
	case pmetric.MetricTypeGauge:
		return c.drawMetricNumberChart(m)
	case pmetric.MetricTypeSum:
		return c.drawMetricNumberChart(m)
	case pmetric.MetricTypeHistogram:
		return c.drawMetricHistogramChart(m)
	case pmetric.MetricTypeExponentialHistogram:
		return c.drawMetricNumberChart(m)
	case pmetric.MetricTypeSummary:
		return c.drawMetricNumberChart(m)
	}
	return layout.KeyMaps{}
}

func (c *chart) drawMetricHistogramChart(m *telemetry.MetricData) layout.KeyMaps {
	dpcount := m.Metric.Histogram().DataPoints().Len()
	chs := make([]*tvxwidgets.BarChart, dpcount)
	sides := make([]*tview.Flex, dpcount)
	for dpi := range dpcount {
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

	if dpcount == 0 {
		return layout.KeyMaps{}
	}
	idx := 0
	c.ch.AddItem(chs[idx], 0, 7, false).AddItem(sides[idx], 0, 3, false)
	c.focusTargets = []layout.FocusableBox{}
	for _, ch := range chs {
		c.focusTargets = append(c.focusTargets, ch)
	}

	return layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone),
			Hidden:      true,
			Description: "",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				if idx < dpcount-1 {
					idx++
				} else {
					idx = 0
				}
				c.ch.Clear().AddItem(chs[idx], 0, 7, false).AddItem(sides[idx], 0, 3, false)
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone),
			Hidden:      true,
			Description: "",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				if idx > 0 {
					idx--
				} else {
					idx = dpcount - 1
				}
				c.ch.Clear().AddItem(chs[idx], 0, 7, false).AddItem(sides[idx], 0, 3, false)
				return nil
			},
		},
	}
}

func (c *chart) drawMetricNumberChart(m *telemetry.MetricData) layout.KeyMaps {
	sname := telemetry.GetServiceNameFromResource(m.ResourceMetric.Resource())
	mcache := c.store.GetMetricCache()
	ms, ok := mcache.GetMetricsBySvcAndMetricName(sname, m.Metric.Name())
	if !ok {
		return layout.KeyMaps{}
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

	// TODO: Delete it after implementing drawMetric* for all types
	if !support {
		txt := tview.NewTextView().SetText("This metric type is not supported")
		c.ch.AddItem(txt, 0, 1, false)
		return layout.KeyMaps{}
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
	data, txts := c.getDataToDraw(dataMap, attrkeys[attrkeyidx], start, end)
	ch := tvxwidgets.NewPlot()
	ch.SetMarker(tvxwidgets.PlotMarkerBraille)
	ch.SetTitle(getTitle(attrkeyidx))
	ch.SetBorder(true)
	ch.SetData(data)
	ch.SetDrawXAxisLabel(false)
	ch.SetLineColor(layout.Colors)

	legend := tview.NewFlex().SetDirection(tview.FlexRow)
	legend.AddItem(txts, 0, 1, false)

	c.ch.AddItem(ch, 0, 7, true).AddItem(legend, 0, 3, false)
	c.focusTargets = []layout.FocusableBox{ch}

	return layout.KeyMaps{
		{
			Key:         tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone),
			Hidden:      true,
			Description: "",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				if attrkeyidx < len(attrkeys)-1 {
					attrkeyidx++
				} else {
					attrkeyidx = 0
				}
				ch.SetTitle(getTitle(attrkeyidx))
				data, txts := c.getDataToDraw(dataMap, attrkeys[attrkeyidx], start, end)
				legend.Clear()
				legend.AddItem(txts, 0, 1, false)
				ch.SetData(data)
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone),
			Hidden:      true,
			Description: "",
			Handler: func(_ *tcell.EventKey) *tcell.EventKey {
				if attrkeyidx > 0 {
					attrkeyidx--
				} else {
					attrkeyidx = len(attrkeys) - 1
				}
				ch.SetTitle(getTitle(attrkeyidx))
				data, txts := c.getDataToDraw(dataMap, attrkeys[attrkeyidx], start, end)
				legend.Clear()
				legend.AddItem(txts, 0, 1, false)
				ch.SetData(data)
				return nil
			},
		},
	}
}

func (c *chart) getDataToDraw(dataMap map[string]map[string][]*pmetric.NumberDataPoint, attrkey string, start, end time.Time) ([][]float64, *tview.TextView) {
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
			d[i][ii] = nullValueFloat64
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
		prevval := nullValueFloat64
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
		txts[i] = fmt.Sprintf("[%s]● %s: %s", layout.Colors[i].String(), attrkey, k)
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

func (c *chart) updateCommands(keyMaps layout.KeyMaps) {
	for _, rm := range c.resizeManagers {
		keyMaps.Merge(rm.KeyMaps())
	}
	for _, ft := range c.focusTargets {
		layout.RegisterCommandList(c.commands, ft, nil, keyMaps)
	}
}

// uint64ToInt converts uint64 into int. When the input is larger than math.MaxInt, it returns math.MaxInt.
func uint64ToInt(u uint64) int {
	if u >= math.MaxInt {
		return math.MaxInt
	}
	return int(u)
}
