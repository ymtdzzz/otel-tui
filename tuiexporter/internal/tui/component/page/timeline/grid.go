package timeline

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/layout"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tui/component/navigation"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	spanNameColumnWidthResizeUnit = 5
	spanNameColumnWidthDefalt     = 30
)

type spanTreeNode struct {
	span     *telemetry.SpanData
	label    string
	box      *tview.Box
	children []*spanTreeNode
	expand   bool
}

type grid struct {
	commands      *tview.TextView
	gridView      *tview.Grid
	tcache        *telemetry.TraceCache
	snameWidth    int
	totalRow      int
	currentRow    int
	tree          []*spanTreeNode
	duration      time.Duration
	nodes         []*spanTreeNode
	items         []*tview.TextView
	resizeManager *layout.ResizeManager
	detail        *detail
	logPane       *logPane
}

func newGrid(
	commands *tview.TextView,
	tcache *telemetry.TraceCache,
	resizeManager *layout.ResizeManager,
	detail *detail,
	logPane *logPane,
) *grid {
	snameWidth := spanNameColumnWidthDefalt
	container := tview.NewGrid().
		SetColumns(snameWidth, 0).
		SetBorders(true)
	container.SetTitle("Trace Timeline (t)").SetBorder(true)

	grid := &grid{
		commands:      commands,
		gridView:      container,
		tcache:        tcache,
		snameWidth:    snameWidth,
		totalRow:      0,
		currentRow:    0,
		nodes:         []*spanTreeNode{},
		items:         []*tview.TextView{},
		resizeManager: resizeManager,
		detail:        detail,
		logPane:       logPane,
	}

	return grid
}

func (g *grid) updateGrid(traceID string) *telemetry.SpanData {
	g.totalRow = 0
	g.currentRow = 0
	g.nodes = []*spanTreeNode{}
	g.items = []*tview.TextView{}

	tree, duration := g.newSpanTree(traceID)
	g.tree = tree
	g.duration = duration

	g.placeSpans()

	g.updateCommands()

	if len(g.nodes) == 0 {
		return nil
	}
	return g.nodes[0].span
}

func (g *grid) prepareTimeline(duration time.Duration) {
	title := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("Spans")
	timeline := tview.NewBox().SetBorder(false).
		SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			// Draw a horizontal line across the middle of the box.
			centerY := y + height/2
			for cx := x + 1; cx < x+width-1; cx++ {
				screen.SetContent(cx, centerY, tview.BoxDrawingsLightHorizontal, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
			}

			// Write some text along the horizontal line.
			unit, count := calculateTimelineUnit(duration)
			for i := range count {
				ratio := float64(i) / float64(count)
				label := roundDownDuration(unit * time.Duration(i)).String()
				if i == 0 {
					label = "0"
				}
				tview.Print(screen, label, x+getXByRatio(ratio, width), centerY, width-2, tview.AlignLeft, tcell.ColorYellow)
			}

			// Space for other content.
			return x + 1, centerY + 1, width - 2, height - (centerY + 1 - y)
		})
	g.gridView.Clear().
		AddItem(title, 0, 0, 1, 1, 0, 0, false).
		AddItem(timeline, 0, 1, 1, 1, 0, 0, false)
}

func (g *grid) placeSpans() {
	g.prepareTimeline(g.duration)
	g.totalRow = 0

	var (
		tvs   []*tview.TextView
		nodes []*spanTreeNode
	)
	for _, n := range g.tree {
		g.totalRow = g.placeSpan(n, g.totalRow, 0, &tvs, &nodes)
	}
	g.nodes = nodes
	g.items = tvs
	if g.getCurrentSpan() != nil {
		navigation.Focus(g.items[g.currentRow])
	}

	rows := make([]int, g.totalRow+2)
	for i := 0; i < g.totalRow+1; i++ {
		rows[i] = 1
	}
	g.gridView.SetRows(rows...)
	log.Printf("totalRow: %d, tviews: %+v", g.totalRow, tvs)

	if g.getCurrentSpan() != nil {
		navigation.Focus(g.items[g.currentRow])
	}
}

func (g *grid) placeSpan(
	node *spanTreeNode,
	row, depth int,
	tvs *[]*tview.TextView,
	nodes *[]*spanTreeNode,
) int {
	row++
	label := node.label
	prefix := ""
	for i := range depth {
		if i == depth-1 {
			prefix = prefix + string(tview.BoxDrawingsLightUpAndRight)
			break
		}
		prefix = prefix + " "
	}
	exp := " "
	if len(node.children) > 0 {
		if node.expand {
			exp = "▼"
		} else {
			exp = "▶"
		}
	}
	tv := g.newTextView(prefix + exp + label)
	*tvs = append(*tvs, tv)
	*nodes = append(*nodes, node)
	g.gridView.AddItem(tv, row, 0, 1, 1, 0, 0, false)
	g.gridView.AddItem(node.box, row, 1, 1, 1, 0, 0, false)
	if !node.expand {
		return row
	}
	sort.SliceStable(node.children, func(i, j int) bool {
		return node.children[i].span.Span.StartTimestamp().AsTime().Before(
			node.children[j].span.Span.StartTimestamp().AsTime(),
		)
	})
	for _, child := range node.children {
		row = g.placeSpan(child, row, depth+1, tvs, nodes)
	}
	return row
}

func (g *grid) newSpanTree(traceID string) (rootNodes []*spanTreeNode, duration time.Duration) {
	spans, ok := g.tcache.GetSpansByTraceID(traceID)
	if !ok {
		return
	}

	start := time.Now().Add(time.Hour * 24)
	end := time.Time{}

	// store memo and calculate start and end time of the trace
	spanMemo := make(map[string]int)
	colorMemo := make(map[string]tcell.Color)
	nodes := []*spanTreeNode{}
	for idx, span := range spans {
		nodes = append(nodes, &spanTreeNode{span: span, expand: true})
		spanMemo[span.Span.SpanID().String()] = idx
		if span.Span.StartTimestamp().AsTime().Before(start) {
			start = span.Span.StartTimestamp().AsTime()
		}
		if span.Span.EndTimestamp().AsTime().After(end) {
			end = span.Span.EndTimestamp().AsTime()
		}
		// color is assigned by the service name
		sname := telemetry.GetServiceNameFromResource(span.ResourceSpan.Resource())
		if _, ok := colorMemo[sname]; !ok {
			colorMemo[sname] = layout.Colors[len(colorMemo)%len(layout.Colors)]
		}
	}
	duration = end.Sub(start)

	// generate span tree
	for _, span := range spans {
		current := span.Span.SpanID().String()
		node := nodes[spanMemo[current]]
		sname := telemetry.GetServiceNameFromResource(span.ResourceSpan.Resource())
		st, en := span.Span.StartTimestamp().AsTime().Sub(start), span.Span.EndTimestamp().AsTime().Sub(start)
		d := en - st
		node.box = createSpan(colorMemo[sname], duration, st, en)
		if span.Span.Status().Code() == ptrace.StatusCodeError {
			node.label = fmt.Sprintf("[!] %s %s", span.Span.Name(), d.String())
		} else {
			node.label = fmt.Sprintf("%s %s", span.Span.Name(), d.String())
		}

		parent := span.Span.ParentSpanID().String()
		_, parentExists := g.tcache.GetSpanByID(parent)
		if !parentExists {
			rootNodes = append(rootNodes, node)
			continue
		}
		parentIdx := spanMemo[parent]
		nodes[parentIdx].children = append(nodes[parentIdx].children, nodes[spanMemo[span.Span.SpanID().String()]])
	}

	// sort root spans by start time
	sort.SliceStable(rootNodes, func(i, j int) bool {
		return rootNodes[i].span.Span.StartTimestamp().AsTime().Before(
			rootNodes[j].span.Span.StartTimestamp().AsTime(),
		)
	})

	return rootNodes, duration
}

func (g *grid) newTextView(text string) *tview.TextView {
	tv := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText(text).
		SetWordWrap(false)

	tv.SetFocusFunc(func() {
		tv.SetBackgroundColor(tcell.ColorWhite)
		tv.SetTextColor(tcell.ColorBlack)
	})
	tv.SetBlurFunc(func() {
		tv.SetBackgroundColor(tcell.ColorNone)
		tv.SetTextColor(tcell.ColorDefault)
	})

	return tv
}

func (g *grid) updateCommands() {
	keyMaps := layout.KeyMaps{
		// FIXME: key 'j' and 'k' should be used to move the focus
		//   but these keys are captured by the parent grid.
		{
			Key: tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone),
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				if g.currentRow < g.totalRow-1 {
					g.currentRow++
					navigation.Focus(g.items[g.currentRow])

					currentSpan := g.getCurrentSpan()
					g.detail.update(currentSpan)
					g.logPane.updateLog(
						currentSpan.Span.TraceID().String(),
						currentSpan.Span.SpanID().String(),
					)
				}
				return nil
			},
		},
		{
			Key: tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone),
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				if g.currentRow > 0 {
					g.currentRow--
					navigation.Focus(g.items[g.currentRow])

					currentSpan := g.getCurrentSpan()
					g.detail.update(currentSpan)
					g.logPane.updateLog(
						currentSpan.Span.TraceID().String(),
						currentSpan.Span.SpanID().String(),
					)
				}
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			Description: "Toggle folding the child spans",
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				g.nodes[g.currentRow].expand = !g.nodes[g.currentRow].expand

				g.placeSpans()

				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone),
			Description: "Widen span name column",
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				_, _, w, _ := g.gridView.GetInnerRect()
				g.snameWidth = widenInLimit(spanNameColumnWidthResizeUnit, g.snameWidth, w)
				g.gridView.SetColumns(g.snameWidth, 0)
				return nil
			},
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone),
			Description: "Narrow span name column",
			Handler: func(event *tcell.EventKey) *tcell.EventKey {
				g.snameWidth = narrowInLimit(spanNameColumnWidthResizeUnit, g.snameWidth, spanNameColumnWidthDefalt)
				g.gridView.SetColumns(g.snameWidth, 0)
				return nil
			},
		},
	}
	keyMaps.Merge(g.resizeManager.KeyMaps())
	layout.RegisterCommandList(g.commands, g.gridView, func() {
		if g.getCurrentSpan() != nil {
			navigation.Focus(g.items[g.currentRow])
		}
	}, keyMaps)
}

func (g *grid) getCurrentSpan() *telemetry.SpanData {
	if g.currentRow < 0 || g.currentRow >= len(g.nodes) {
		return nil
	}
	return g.nodes[g.currentRow].span
}

func createSpan(color tcell.Color, total, start, end time.Duration) (span *tview.Box) {
	return tview.NewBox().SetBorder(false).
		SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			// Draw a horizontal line across the middle of the box.
			centerY := y + height/2
			sRatio := float64(start) / float64(total)
			eRatio := float64(end) / float64(total)
			s := x + getXByRatio(sRatio, width)
			e := x + getXByRatio(eRatio, width)
			if s == e {
				screen.SetContent(s, centerY, tview.BoxDrawingsHeavyVertical, nil, tcell.StyleDefault.Foreground(color))
			} else {
				for cx := s; cx < e; cx++ {
					screen.SetContent(cx, centerY, tview.BlockMediumShade, nil, tcell.StyleDefault.Foreground(color))
				}
			}

			// Space for other content.
			return x + 1, centerY + 1, width - 2, height - (centerY + 1 - y)
		})
}

func getXByRatio(ratio float64, width int) int {
	return int(float64(width) * ratio)
}

func calculateTimelineUnit(duration time.Duration) (unit time.Duration, count int) {
	// TODO: set count depends on the width
	count = 5
	unit = duration / time.Duration(count)
	return
}

func roundDownDuration(d time.Duration) time.Duration {
	if d < time.Microsecond {
		return d - (d % time.Nanosecond)
	} else if d < time.Millisecond {
		return d - (d % time.Microsecond)
	} else if d < time.Second {
		return d - (d % time.Millisecond)
	} else if d < time.Minute {
		return d - (d % time.Second)
	} else if d < time.Hour {
		return d - (d % time.Minute)
	}
	return d
}

func narrowInLimit(step, curr, limit int) int {
	if curr-step >= limit {
		return curr - step
	}
	return curr
}

func widenInLimit(step, curr, limit int) int {
	if curr+step <= limit {
		return curr + step
	}
	return curr
}
