package component

import (
	"fmt"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

type spanTreeNode struct {
	span     *telemetry.SpanData
	label    string
	box      *tview.Box
	children []*spanTreeNode
}

func DrawTimeline(traceID string, cache *telemetry.TraceCache) tview.Primitive {
	if traceID == "" || cache == nil {
		return NewPrimitive("No spans found")
	}
	_, ok := cache.GetSpansByTraceID(traceID)
	if !ok {
		return NewPrimitive("No spans found")
	}

	title := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("Spans")
	tree, duration := newSpanTree(traceID, cache)

	// draw timeline
	timeline := tview.NewBox().SetBorder(false).
		SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			// Draw a horizontal line across the middle of the box.
			centerY := y + height/2
			for cx := x + 1; cx < x+width-1; cx++ {
				screen.SetContent(cx, centerY, tview.BoxDrawingsLightHorizontal, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
			}

			// Write some text along the horizontal line.
			durunit := duration.Milliseconds() / 10
			tview.Print(screen, "0", x+GetXByRatio(0, width), centerY, width-2, tview.AlignLeft, tcell.ColorYellow)
			for i := 1; i < 10; i++ {
				tview.Print(screen, fmt.Sprintf("%dms", durunit*int64(i)), x+GetXByRatio(float64(i)*0.1, width), centerY, width-2, tview.AlignLeft, tcell.ColorYellow)
			}

			// Space for other content.
			return x + 1, centerY + 1, width - 2, height - (centerY + 1 - y)
		})

	// place spans on the timeline
	grid := tview.NewGrid().
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(title, 0, 0, 1, 1, 0, 0, false).
		AddItem(timeline, 0, 1, 1, 1, 0, 0, false)

	row := 0
	for _, n := range tree {
		row = placeSpan(grid, n, row, 0)
	}

	rows := make([]int, row+2, row+2)
	for i := 0; i < row+1; i++ {
		rows[i] = 1
	}

	grid.SetRows(rows...)

	return grid
}

func placeSpan(grid *tview.Grid, node *spanTreeNode, row, depth int) int {
	row++
	label := node.label
	for i := 0; i < depth; i++ {
		label = ">" + label
	}
	grid.AddItem(NewPrimitive(label), row, 0, 1, 1, 0, 0, false)
	grid.AddItem(node.box, row, 1, 1, 1, 0, 0, false)
	for _, child := range node.children {
		row = placeSpan(grid, child, row, depth+1)
	}
	return row
}

func newSpanTree(traceID string, cache *telemetry.TraceCache) (rootNodes []*spanTreeNode, duration time.Duration) {
	spans, ok := cache.GetSpansByTraceID(traceID)
	if !ok {
		return
	}

	start := time.Now().Add(time.Hour * 24)
	end := time.Now().Add(-time.Hour * 24)

	// store memo and calculate start and end time of the trace
	spanMemo := make(map[string]int)
	nodes := []*spanTreeNode{}
	for idx, span := range spans {
		nodes = append(nodes, &spanTreeNode{span: span})
		spanMemo[span.Span.SpanID().String()] = idx
		if span.Span.StartTimestamp().AsTime().Before(start) {
			start = span.Span.StartTimestamp().AsTime()
		}
		if span.Span.EndTimestamp().AsTime().After(end) {
			end = span.Span.EndTimestamp().AsTime()
		}
	}
	duration = end.Sub(start)

	// generate span tree
	for _, span := range spans {
		current := span.Span.SpanID().String()
		node := nodes[spanMemo[current]]
		st, en := span.Span.StartTimestamp().AsTime().Sub(start), span.Span.EndTimestamp().AsTime().Sub(start)
		d := en - st
		node.box = CreateSpan(current, int(duration.Milliseconds()), int(st.Milliseconds()), int(en.Milliseconds()))
		node.label = fmt.Sprintf("%s %dms", span.Span.Name(), d.Milliseconds())

		parent := span.Span.ParentSpanID().String()
		_, parentExists := cache.GetSpanByID(parent)
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

func NewPrimitive(text string) tview.Primitive {
	return tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText(text)
}

func GetXByRatio(ratio float64, width int) int {
	return int(float64(width) * ratio)
}

func CreateSpan(name string, total, start, end int) (span *tview.Box) {
	return tview.NewBox().SetBorder(false).
		SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			// Draw a horizontal line across the middle of the box.
			centerY := y + height/2
			sRatio := float64(start) / float64(total)
			eRatio := float64(end) / float64(total)
			s := x + GetXByRatio(sRatio, width)
			e := x + GetXByRatio(eRatio, width)
			if s == e {
				screen.SetContent(s, centerY, tview.BoxDrawingsHeavyVertical, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
			} else {
				for cx := s; cx < e; cx++ {
					screen.SetContent(cx, centerY, tview.BlockMediumShade, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
				}
			}

			// Space for other content.
			return x + 1, centerY + 1, width - 2, height - (centerY + 1 - y)
		})
}
