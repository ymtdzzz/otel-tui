package component

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/datetime"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	TIMELINE_DETAILS_IDX                = 1 // index of details in the base flex container
	TIMELINE_TREE_TITLE                 = "Details (d)"
	SPAN_NAME_COLUMN_WIDTH_RESIZE_UNIT  = 5
	SPAN_NAME_COLUMN_WIDTH_DEFAULT      = 30
	DEFAULT_PROPORTION_TIMELINE_DETAILS = 21
	DEFAULT_PROPORTION_TIMELINE_GRID    = 29
)

var colors = []tcell.Color{
	// https://color.adobe.com/[otel-tui]-Span-Color-Theme-color-theme-08c8f7c5-7b93-4936-ae75-8f91fc045fd5
	tcell.ColorAliceBlue,
	tcell.ColorBurlyWood,
	tcell.ColorCadetBlue,
	tcell.ColorCoral,
	tcell.ColorCornsilk,
	tcell.ColorGold,
	tcell.ColorLightBlue,
	tcell.ColorLightGreen,
	tcell.ColorLemonChiffon,
	tcell.ColorMediumTurquoise,
}

type spanTreeNode struct {
	span     *telemetry.SpanData
	label    string
	box      *tview.Box
	children []*spanTreeNode
	expand   bool
}

func DrawTimeline(commands *tview.TextView, showModalFn showModalFunc, hideModalFn hideModalFunc, traceID string, tcache *telemetry.TraceCache, lcache *telemetry.LogCache, setFocusFn func(p tview.Primitive)) tview.Primitive {
	if traceID == "" || tcache == nil {
		return newTextView(commands, "No spans found")
	}
	_, ok := tcache.GetSpansByTraceID(traceID)
	if !ok {
		return newTextView(commands, "No spans found")
	}

	base := tview.NewFlex().SetDirection(tview.FlexRow)
	traceContainer := tview.NewFlex().SetDirection(tview.FlexColumn)

	// draw timeline
	tree, duration := newSpanTree(traceID, tcache)

	// place spans on the timeline
	snameWidth := SPAN_NAME_COLUMN_WIDTH_DEFAULT
	grid := tview.NewGrid().
		SetColumns(snameWidth, 0).
		SetBorders(true)
	grid.SetTitle("Trace Timeline (t)").SetBorder(true)
	clearTimeline(duration, grid)
	registerCommandList(commands, grid, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone),
			description: "Move up",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone),
			description: "Move down",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModCtrl),
			description: "Expand the width",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModCtrl),
			description: "Reduce the width",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone),
			description: "Back to Traces",
		},
	})

	var (
		tvs   []*tview.TextView
		nodes []*spanTreeNode
	)
	totalRow := 0
	for _, n := range tree {
		totalRow = placeSpan(commands, grid, n, totalRow, 0, &tvs, &nodes)
	}

	// details
	details := getSpanInfoTree(commands, showModalFn, hideModalFn, nodes[0].span, TIMELINE_TREE_TITLE)
	detailspro := DEFAULT_PROPORTION_TIMELINE_DETAILS
	gridpro := DEFAULT_PROPORTION_TIMELINE_GRID

	details.SetInputCapture(detailsInputFunc(traceContainer, grid, details, &gridpro, &detailspro))

	// logs
	logs := tview.NewTable().SetBorders(false).SetSelectable(true, false)
	var ldft *LogDataForTable
	logs.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlF {
			if ldft != nil {
				ldft.SetFullDatetime(!ldft.IsFullDatetime())
			}
			return nil
		}
		return event
	})

	updateLogTableFn := func(traceID, spanID string, all bool) {
		logCount := 0
		if lds, ok := lcache.GetLogsByTraceID(traceID); ok {
			if !all && spanID != "" {
				flds := []*telemetry.LogData{}
				for _, ld := range lds {
					if ld.Log.SpanID().String() == spanID {
						flds = append(flds, ld)
					}
				}
				lds = flds
			}
			logCount = len(lds)
			logData := NewLogDataForTableForTimeline(&lds)
			if ldft != nil {
				logData.SetFullDatetime(ldft.IsFullDatetime())
			}
			ldft = &logData
			logs.SetContent(&logData)
			attachModalForTableRows(logs, &logData, showModalFn, hideModalFn)
		}
		logs.SetBorder(true).SetTitle(fmt.Sprintf("Logs (l) -- %d logs found (L: toggle collapse, A: toggle filter by span)", logCount))
	}

	allLogs := false
	updateLogTableFn(traceID, nodes[0].span.Span.SpanID().String(), allLogs)
	registerCommandList(commands, logs, nil, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModCtrl),
			description: "Toggle full datetime",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone),
			description: "Back to Traces",
		},
	})

	rows := make([]int, totalRow+2)
	for i := 0; i < totalRow+1; i++ {
		rows[i] = 1
	}

	grid.SetRows(rows...)

	log.Printf("totalRow: %d, tviews: %+v", totalRow, tvs)

	// set key handler to grid
	currentRow := 0
	if totalRow > 0 {
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
					oldDetails := traceContainer.GetItem(TIMELINE_DETAILS_IDX)
					traceContainer.RemoveItem(oldDetails)
					details := getSpanInfoTree(commands, showModalFn, hideModalFn, nodes[currentRow].span, TIMELINE_TREE_TITLE)
					details.SetInputCapture(detailsInputFunc(traceContainer, grid, details, &gridpro, &detailspro))
					traceContainer.AddItem(details, 0, detailspro, false)

					// update log table
					updateLogTableFn(traceID, nodes[currentRow].span.Span.SpanID().String(), allLogs)
				}
				return nil
			case tcell.KeyUp:
				if currentRow > 0 {
					currentRow--
					setFocusFn(tvs[currentRow])
					details = getSpanInfoTree(commands, showModalFn, hideModalFn, nodes[currentRow].span, TIMELINE_TREE_TITLE)

					// update details
					oldDetails := traceContainer.GetItem(TIMELINE_DETAILS_IDX)
					traceContainer.RemoveItem(oldDetails)
					details := getSpanInfoTree(commands, showModalFn, hideModalFn, nodes[currentRow].span, TIMELINE_TREE_TITLE)
					details.SetInputCapture(detailsInputFunc(traceContainer, grid, details, &gridpro, &detailspro))
					traceContainer.AddItem(details, 0, detailspro, false)

					// update log table
					updateLogTableFn(traceID, nodes[currentRow].span.Span.SpanID().String(), allLogs)
				}
				return nil
			case tcell.KeyCtrlL:
				_, _, w, _ := grid.GetInnerRect()
				snameWidth = widenInLimit(SPAN_NAME_COLUMN_WIDTH_RESIZE_UNIT, snameWidth, w)
				grid.SetColumns(snameWidth, 0)
				return nil
			case tcell.KeyCtrlH:
				snameWidth = narrowInLimit(SPAN_NAME_COLUMN_WIDTH_RESIZE_UNIT, snameWidth, SPAN_NAME_COLUMN_WIDTH_DEFAULT)
				grid.SetColumns(snameWidth, 0)
				return nil
			case tcell.KeyEnter:
				nodes[currentRow].expand = !nodes[currentRow].expand

				tvs = []*tview.TextView{}
				nodes = []*spanTreeNode{}
				clearTimeline(duration, grid)
				totalRow = 0
				for _, n := range tree {
					totalRow = placeSpan(commands, grid, n, totalRow, 0, &tvs, &nodes)
				}
				rows := make([]int, totalRow+2)
				for i := 0; i < totalRow+1; i++ {
					rows[i] = 1
				}
				grid.SetRows(rows...)
				setFocusFn(tvs[currentRow])
				return nil
			}
			return event
		})
	}

	details.SetBorder(true).SetTitle("Details (d)")

	isLogCollapse := true
	traceContainer.AddItem(grid, 0, DEFAULT_PROPORTION_TIMELINE_GRID, true).
		AddItem(details, 0, DEFAULT_PROPORTION_TIMELINE_DETAILS, false)
	base.AddItem(traceContainer, 0, 1, true).
		AddItem(logs, 2, 1, false)

	base.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'd':
			setFocusFn(traceContainer.GetItem(TIMELINE_DETAILS_IDX))
			return nil
		case 't':
			setFocusFn(grid)
			return nil
		case 'l':
			setFocusFn(logs)
			return nil
		case 'L':
			isLogCollapse = !isLogCollapse
			logHeight := 10
			if isLogCollapse {
				logHeight = 2
			}
			base.Clear().AddItem(traceContainer, 0, 1, traceContainer.HasFocus()).
				AddItem(logs, logHeight, 1, logs.HasFocus())

			return nil
		case 'A':
			allLogs = !allLogs
			updateLogTableFn(traceID, nodes[currentRow].span.Span.SpanID().String(), allLogs)

			return nil
		}
		return event
	})

	return base
}

func clearTimeline(duration time.Duration, grid *tview.Grid) {
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
	grid.Clear().
		AddItem(title, 0, 0, 1, 1, 0, 0, false).
		AddItem(timeline, 0, 1, 1, 1, 0, 0, false)
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

func detailsInputFunc(traceContainer *tview.Flex, grid *tview.Grid, details *tview.TreeView, gridpro, detailspro *int) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlL:
			if *detailspro <= 1 {
				return nil
			}
			*gridpro++
			*detailspro--
			traceContainer.ResizeItem(grid, 0, *gridpro).
				ResizeItem(details, 0, *detailspro)
			return nil
		case tcell.KeyCtrlH:
			if *gridpro <= 1 {
				return nil
			}
			*gridpro--
			*detailspro++
			traceContainer.ResizeItem(grid, 0, *gridpro).
				ResizeItem(details, 0, *detailspro)
			return nil
		}
		return event
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

func placeSpan(commands *tview.TextView, grid *tview.Grid, node *spanTreeNode, row, depth int, tvs *[]*tview.TextView, nodes *[]*spanTreeNode) int {
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
	exp := " "
	if len(node.children) > 0 {
		if node.expand {
			exp = "▼"
		} else {
			exp = "▶"
		}
	}
	tv := newTextView(commands, prefix+exp+label)
	*tvs = append(*tvs, tv)
	*nodes = append(*nodes, node)
	grid.AddItem(tv, row, 0, 1, 1, 0, 0, false)
	grid.AddItem(node.box, row, 1, 1, 1, 0, 0, false)
	if !node.expand {
		return row
	}
	sort.SliceStable(node.children, func(i, j int) bool {
		return node.children[i].span.Span.StartTimestamp().AsTime().Before(
			node.children[j].span.Span.StartTimestamp().AsTime(),
		)
	})
	for _, child := range node.children {
		row = placeSpan(commands, grid, child, row, depth+1, tvs, nodes)
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
			colorMemo[sname] = colors[len(colorMemo)%len(colors)]
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

func newTextView(commands *tview.TextView, text string) *tview.TextView {
	tv := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText(text).
		SetWordWrap(false)
		// FIXME: The parent component Grid does not trigger FocusFunc so set it on child tvs
		//   But this is redundant. Is there any better ways?
	registerCommandList(commands, tv, func() {
		tv.SetBackgroundColor(tcell.ColorWhite)
		tv.SetTextColor(tcell.ColorBlack)
	}, KeyMaps{
		{
			key:         tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone),
			description: "Move up",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone),
			description: "Move down",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
			description: "Toggle folding the child spans",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModCtrl),
			description: "Expand the width",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModCtrl),
			description: "Reduce the width",
		},
		{
			key:         tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone),
			description: "Back to Traces",
		},
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

func getSpanInfoTree(commands *tview.TextView, showModalFn showModalFunc, hideModalFn hideModalFunc, span *telemetry.SpanData, title string) *tview.TreeView {
	traceID := span.Span.TraceID().String()
	sname := telemetry.GetServiceNameFromResource(span.ResourceSpan.Resource())
	root := tview.NewTreeNode(fmt.Sprintf("%s (%s)", sname, traceID))
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
	scodeText := ""
	if scode == ptrace.StatusCodeError {
		scodeText = fmt.Sprintf("code: %s ⚠️", scode)
	} else {
		scodeText = fmt.Sprintf("code: %s", scode)
	}
	scodeNode := tview.NewTreeNode(scodeText)
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

	startTime := datetime.GetFullTime(span.Span.StartTimestamp().AsTime())
	startTimeNode := tview.NewTreeNode(fmt.Sprintf("start time: %s", startTime))
	root.AddChild(startTimeNode)

	endTime := datetime.GetFullTime(span.Span.EndTimestamp().AsTime())
	endTimeNode := tview.NewTreeNode(fmt.Sprintf("end time: %s", endTime))
	root.AddChild(endTimeNode)

	dropped := span.ResourceSpan.Resource().DroppedAttributesCount()
	droppedNode := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", dropped))
	root.AddChild(droppedNode)

	attrs := tview.NewTreeNode("Attributes")
	appendAttrsSorted(attrs, span.Span.Attributes())
	root.AddChild(attrs)

	// events
	events := tview.NewTreeNode("Events")
	for ei := 0; ei < span.Span.Events().Len(); ei++ {
		event := span.Span.Events().At(ei)
		name := event.Name()
		eventNode := tview.NewTreeNode(name)

		timestamp := datetime.GetFullTime(event.Timestamp().AsTime())
		timestampNode := tview.NewTreeNode(fmt.Sprintf("timestamp: %s", timestamp))
		eventNode.AddChild(timestampNode)

		dropped := event.DroppedAttributesCount()
		droppedNode := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", dropped))
		eventNode.AddChild(droppedNode)

		attrs := tview.NewTreeNode("Attributes")
		appendAttrsSorted(attrs, event.Attributes())
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
		appendAttrsSorted(attrs, link.Attributes())
		linkNode.AddChild(attrs)

		links.AddChild(linkNode)
	}
	root.AddChild(links)

	// resource info
	rs := span.ResourceSpan
	r := rs.Resource()
	resource := tview.NewTreeNode("Resource")
	root.AddChild(resource)
	rdropped := tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", r.DroppedAttributesCount()))
	resource.AddChild(rdropped)
	rschema := tview.NewTreeNode(fmt.Sprintf("schema url: %s", rs.SchemaUrl()))
	resource.AddChild(rschema)

	rattrs := tview.NewTreeNode("Attributes")
	appendAttrsSorted(rattrs, r.Attributes())
	resource.AddChild(rattrs)

	// scope info
	scopes := tview.NewTreeNode("Scopes")
	for si := 0; si < rs.ScopeSpans().Len(); si++ {
		ss := rs.ScopeSpans().At(si)
		scope := tview.NewTreeNode(ss.Scope().Name())
		sschema := tview.NewTreeNode(fmt.Sprintf("schema url: %s", ss.SchemaUrl()))
		scope.AddChild(sschema)

		scope.AddChild(tview.NewTreeNode(fmt.Sprintf("version: %s", ss.Scope().Version())))
		scope.AddChild(tview.NewTreeNode(fmt.Sprintf("dropped attributes count: %d", ss.Scope().DroppedAttributesCount())))

		attrs := tview.NewTreeNode("Attributes")
		appendAttrsSorted(attrs, ss.Scope().Attributes())
		scope.AddChild(attrs)

		scopes.AddChild(scope)
	}
	resource.AddChild(scopes)

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
		{
			key:         tcell.NewEventKey(tcell.KeyEsc, ' ', tcell.ModNone),
			description: "Back to Traces",
		},
	})

	return tree
}
