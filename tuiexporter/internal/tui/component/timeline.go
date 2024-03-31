package component

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

const (
	TIMELINE_DETAILS_IDX = 1 // index of details in the base flex container
	TIMELINE_TREE_TITLE  = "Details (d)"
)

type spanTreeNode struct {
	span     *telemetry.SpanData
	label    string
	box      *tview.Box
	children []*spanTreeNode
}

func DrawTimeline(traceID string, cache *telemetry.TraceCache, setFocusFn func(p tview.Primitive)) (tview.Primitive, KeyMaps) {
	if traceID == "" || cache == nil {
		return newTextView("No spans found"), KeyMaps{}
	}
	_, ok := cache.GetSpansByTraceID(traceID)
	if !ok {
		return newTextView("No spans found"), KeyMaps{}
	}

	base := tview.NewFlex().SetDirection(tview.FlexColumn)

	// draw timeline
	title := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("Spans")
	tree, duration := newSpanTree(traceID, cache)

	timeline := tview.NewBox().SetBorder(false).
		SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			// Draw a horizontal line across the middle of the box.
			centerY := y + height/2
			for cx := x + 1; cx < x+width-1; cx++ {
				screen.SetContent(cx, centerY, tview.BoxDrawingsLightHorizontal, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
			}

			// Write some text along the horizontal line.
			unit, count := calculateTimelineUnit(duration)
			for i := 0; i < count; i++ {
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

	// place spans on the timeline
	grid := tview.NewGrid().
		SetColumns(30, 0). // TODO: dynamic width
		SetBorders(true).
		AddItem(title, 0, 0, 1, 1, 0, 0, false).
		AddItem(timeline, 0, 1, 1, 1, 0, 0, false)
	grid.SetTitle("Trace Timeline (t)").SetBorder(true)

	var (
		tvs   []*tview.TextView
		nodes []*spanTreeNode
	)
	totalRow := 0
	for _, n := range tree {
		totalRow = placeSpan(grid, n, totalRow, 0, &tvs, &nodes)
	}

	// details
	details := getSpanInfoTree(nodes[0].span, TIMELINE_TREE_TITLE)

	rows := make([]int, totalRow+2)
	for i := 0; i < totalRow+1; i++ {
		rows[i] = 1
	}

	grid.SetRows(rows...)

	log.Printf("totalRow: %d, tviews: %+v", totalRow, tvs)

	// set key handler to grid
	if totalRow > 0 {
		currentRow := 0
		setFocusFn(tvs[currentRow])

		grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			// FIXME: key 'j' and 'k' should be used to move the focus
			//   but these keys are captured by the parent grid.
			switch event.Key() {
			case tcell.KeyDown:
				if currentRow < totalRow-1 {
					currentRow++
					setFocusFn(tvs[currentRow])

					// update details
					oldDetails := base.GetItem(TIMELINE_DETAILS_IDX)
					base.RemoveItem(oldDetails)
					details := getSpanInfoTree(nodes[currentRow].span, TIMELINE_TREE_TITLE)
					base.AddItem(details, 0, 3, false)
				}
				return nil
			case tcell.KeyUp:
				if currentRow > 0 {
					currentRow--
					setFocusFn(tvs[currentRow])
					details = getSpanInfoTree(nodes[currentRow].span, TIMELINE_TREE_TITLE)

					// update details
					oldDetails := base.GetItem(TIMELINE_DETAILS_IDX)
					base.RemoveItem(oldDetails)
					details := getSpanInfoTree(nodes[currentRow].span, TIMELINE_TREE_TITLE)
					base.AddItem(details, 0, 3, false)
				}
				return nil
			}
			return event
		})
	}

	details.SetBorder(true).SetTitle("Details (d)")

	base.AddItem(grid, 0, 7, true).
		AddItem(details, 0, 3, false)

	base.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'd':
			log.Printf("d key pressed")
			setFocusFn(base.GetItem(TIMELINE_DETAILS_IDX))
			return nil
		case 't':
			log.Printf("t key pressed")
			setFocusFn(grid)
			return nil
		}
		return event
	})

	return base, KeyMaps{
		*tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone):   "Move up",
		*tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone): "Move down",
	}
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

func placeSpan(grid *tview.Grid, node *spanTreeNode, row, depth int, tvs *[]*tview.TextView, nodes *[]*spanTreeNode) int {
	row++
	label := node.label
	prefix := ""
	for i := 0; i < depth; i++ {
		if i == depth-1 {
			prefix = prefix + string(tview.BoxDrawingsLightUpAndRight)
			break
		}
		prefix = prefix + " "
	}
	tv := newTextView(prefix + label)
	*tvs = append(*tvs, tv)
	*nodes = append(*nodes, node)
	grid.AddItem(tv, row, 0, 1, 1, 0, 0, false)
	grid.AddItem(node.box, row, 1, 1, 1, 0, 0, false)
	for _, child := range node.children {
		row = placeSpan(grid, child, row, depth+1, tvs, nodes)
	}
	return row
}

func newSpanTree(traceID string, cache *telemetry.TraceCache) (rootNodes []*spanTreeNode, duration time.Duration) {
	spans, ok := cache.GetSpansByTraceID(traceID)
	if !ok {
		return
	}

	start := time.Now().Add(time.Hour * 24)
	end := time.Time{}

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
		node.box = createSpan(current, duration, st, en)
		node.label = fmt.Sprintf("%s %s", span.Span.Name(), d.String())

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

func newTextView(text string) *tview.TextView {
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

func getXByRatio(ratio float64, width int) int {
	return int(float64(width) * ratio)
}

func createSpan(name string, total, start, end time.Duration) (span *tview.Box) {
	return tview.NewBox().SetBorder(false).
		SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			// Draw a horizontal line across the middle of the box.
			centerY := y + height/2
			sRatio := float64(start) / float64(total)
			eRatio := float64(end) / float64(total)
			s := x + getXByRatio(sRatio, width)
			e := x + getXByRatio(eRatio, width)
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

func getSpanInfoTree(span *telemetry.SpanData, title string) *tview.TreeView {
	traceID := span.Span.TraceID().String()
	sname, _ := span.ResourceSpan.Resource().Attributes().Get("service.name")
	root := tview.NewTreeNode(fmt.Sprintf("%s (%s)", sname.AsString(), traceID))
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	tree.SetBorder(true).SetTitle(title)

	spanID := span.Span.SpanID().String()
	spanNode := tview.NewTreeNode(fmt.Sprintf("span id: %s", spanID))
	root.AddChild(spanNode)

	parentSpanID := span.Span.ParentSpanID().String()
	parentSpanNode := tview.NewTreeNode(fmt.Sprintf("parent span id: %s", parentSpanID))
	root.AddChild(parentSpanNode)

	state := span.Span.TraceState().AsRaw()
	stateNode := tview.NewTreeNode(fmt.Sprintf("trace state: %s", state))
	root.AddChild(stateNode)

	status := tview.NewTreeNode("Status")
	smessage := span.Span.Status().Message()
	smessageNode := tview.NewTreeNode(fmt.Sprintf("message: %s", smessage))
	status.AddChild(smessageNode)
	scode := span.Span.Status().Code()
	scodeNode := tview.NewTreeNode(fmt.Sprintf("code: %s", scode))
	status.AddChild(scodeNode)
	root.AddChild(status)

	flags := span.Span.Flags()
	flagsNode := tview.NewTreeNode(fmt.Sprintf("flags: %d", flags))
	root.AddChild(flagsNode)

	name := span.Span.Name()
	nameNode := tview.NewTreeNode(fmt.Sprintf("name: %s", name))
	root.AddChild(nameNode)

	kind := span.Span.Kind()
	kindNode := tview.NewTreeNode(fmt.Sprintf("kind: %s", kind))
	root.AddChild(kindNode)

	duration := span.Span.EndTimestamp().AsTime().Sub(span.Span.StartTimestamp().AsTime())
	durationNode := tview.NewTreeNode(fmt.Sprintf("duration: %s", duration.String()))
	root.AddChild(durationNode)

	startTime := span.Span.StartTimestamp().AsTime().Format("2006/01/02 15:04:05.000000")
	startTimeNode := tview.NewTreeNode(fmt.Sprintf("start time: %s", startTime))
	root.AddChild(startTimeNode)

	endTime := span.Span.EndTimestamp().AsTime().Format("2006/01/02 15:04:05.000000")
	endTimeNode := tview.NewTreeNode(fmt.Sprintf("end time: %s", endTime))
	root.AddChild(endTimeNode)

	dropped := span.ResourceSpan.Resource().DroppedAttributesCount()
	droppedNode := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", dropped))
	root.AddChild(droppedNode)

	attrs := tview.NewTreeNode("Attributes")
	for k, v := range span.Span.Attributes().AsRaw() {
		attr := tview.NewTreeNode(fmt.Sprintf("%s: %s", k, v))
		attrs.AddChild(attr)
	}
	root.AddChild(attrs)

	// events
	events := tview.NewTreeNode("Events")
	for ei := 0; ei < span.Span.Events().Len(); ei++ {
		event := span.Span.Events().At(ei)
		name := event.Name()
		eventNode := tview.NewTreeNode(name)

		timestamp := event.Timestamp().AsTime().Format("2006/01/02 15:04:05.000000")
		timestampNode := tview.NewTreeNode(fmt.Sprintf("timestamp: %s", timestamp))
		eventNode.AddChild(timestampNode)

		dropped := event.DroppedAttributesCount()
		droppedNode := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", dropped))
		eventNode.AddChild(droppedNode)

		attrs := tview.NewTreeNode("Attributes")
		for k, v := range event.Attributes().AsRaw() {
			attr := tview.NewTreeNode(fmt.Sprintf("%s: %s", k, v))
			attrs.AddChild(attr)
		}
		eventNode.AddChild(attrs)

		events.AddChild(eventNode)
	}
	root.AddChild(events)

	// links
	links := tview.NewTreeNode("Links")
	for li := 0; li < span.Span.Links().Len(); li++ {
		link := span.Span.Links().At(li)
		linkNode := tview.NewTreeNode(fmt.Sprintf("link %d", li))

		linkTraceID := link.TraceID().String()
		linkTraceIDNode := tview.NewTreeNode(fmt.Sprintf("trace id: %s", linkTraceID))
		linkNode.AddChild(linkTraceIDNode)

		linkSpanID := link.SpanID().String()
		linkSpanIDNode := tview.NewTreeNode(fmt.Sprintf("span id: %s", linkSpanID))
		linkNode.AddChild(linkSpanIDNode)

		linkFlags := link.Flags()
		linkFlagsNode := tview.NewTreeNode(fmt.Sprintf("flags: %d", linkFlags))
		linkNode.AddChild(linkFlagsNode)

		linkState := link.TraceState().AsRaw()
		linkStateNode := tview.NewTreeNode(fmt.Sprintf("trace state: %s", linkState))
		linkNode.AddChild(linkStateNode)

		linkDropped := link.DroppedAttributesCount()
		linkDroppedNode := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", linkDropped))
		linkNode.AddChild(linkDroppedNode)

		attrs := tview.NewTreeNode("Attributes")
		for k, v := range link.Attributes().AsRaw() {
			attr := tview.NewTreeNode(fmt.Sprintf("%s: %s", k, v))
			attrs.AddChild(attr)
		}
		linkNode.AddChild(attrs)

		links.AddChild(linkNode)
	}
	root.AddChild(links)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
	})

	return tree
}
