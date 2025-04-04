package trace

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/component/table"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/component/tree"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/help"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

const (
	MINIMUM_INDIVISUAL_WIDTH = 20
	OFFSET_STEP              = 2
)

var traceCellMappers = table.CellMappers[*telemetry.SpanData]{
	{
		Header: "Service Name",
		GetTextRowFn: func(data *telemetry.SpanData) string {
			return data.GetServiceName()
		},
	},
	{
		Header: "Latency",
		GetTextRowFn: func(data *telemetry.SpanData) string {
			return data.GetDurationText()
		},
	},
	{
		Header: "Received At",
		GetTextRowFn: func(data *telemetry.SpanData) string {
			return data.GetReceivedAtText()
		},
	},
	{
		Header: "Span Name",
		GetTextRowFn: func(data *telemetry.SpanData) string {
			return data.GetSpanName()
		},
	},
}

type Model struct {
	keyMap        keyMap
	styles        styles
	focus         bool
	store         *telemetry.Store
	svcSpans      table.Model[*telemetry.SpanData]
	details       tree.Model
	modal         viewport.Model
	width, height int
	splitOffset   int
}

func New(store *telemetry.Store) Model {
	keyMap := defaultKeyMap()

	details := tree.New(tree.TREE_ID_TRACE, func(label string) tea.Cmd {
		return func() tea.Msg {
			return tuiv2.SetTextModalMsg{
				Type: tuiv2.MODAL_TYPE_TRACE,
				Text: label,
			}
		}
	}, func(msg tea.KeyMsg) tea.Cmd {
		return func() tea.Msg {
			if key.Matches(msg,
				keyMap.modalScrollUp,
				keyMap.modalScrollDown,
				keyMap.moveSplitRight,
				keyMap.moveSplitLeft,
			) {
				return nil
			}
			return tuiv2.SetTextModalMsg{
				Type: tuiv2.MODAL_TYPE_TRACE,
				Text: "",
			}
		}
	})

	traces := table.New[*telemetry.SpanData](
		table.TABLE_ID_TRACE,
		(*[]*telemetry.SpanData)(store.GetFilteredSvcSpans()),
		traceCellMappers,
		func(selected *telemetry.SpanData, idx int) tea.Cmd {
			return func() tea.Msg {
				return tree.UpdateTreeMsg{
					Root: getTraceInfoTreeNode(store.GetFilteredServiceSpansByIdx(idx)),
				}
			}
		},
	)
	traces.Focus(true)

	modal := viewport.New(0, 0)
	modal.KeyMap = viewport.KeyMap{}

	m := Model{
		keyMap:   keyMap,
		styles:   defaultStyles(),
		store:    store,
		svcSpans: traces,
		details:  details,
		modal:    modal,
	}

	return m
}

func (m *Model) Focus(focus bool) {
	m.focus = focus
}

func (m Model) Focused() bool {
	return m.focus
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.svcSpans.Init(),
		m.details.Init(),
		m.modal.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	if !m.focus {
		// IF not focused, pass messages other than key messages
		if _, ok := msg.(tea.KeyMsg); !ok {
			cmds = append(cmds, m.handleMsg(msg)...)
		}
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.focusTraces):
			m.svcSpans.Focus(true)
			m.details.Focus(false)
			cmds = append(cmds, closeModalCmd())
			cmds = append(cmds, tuiv2.UpdateHelpKeysCmd())
		case key.Matches(msg, m.keyMap.focusDetails):
			if !m.svcSpans.QueryFocused() {
				m.svcSpans.Focus(false)
				m.details.Focus(true)
				cmds = append(cmds, tuiv2.UpdateHelpKeysCmd())
			}
		case key.Matches(msg, m.keyMap.moveSplitRight, m.keyMap.modalScrollRight):
			if !m.isModalVisible() {
				m.splitOffset += OFFSET_STEP
				m.UpdateLayout(0, 0)
			} else {
				m.modal.ScrollRight(1)
			}
		case key.Matches(msg, m.keyMap.moveSplitLeft, m.keyMap.modalScrollLeft):
			if !m.isModalVisible() {
				m.splitOffset -= OFFSET_STEP
				m.UpdateLayout(0, 0)
			} else {
				m.modal.ScrollLeft(1)
			}
		case key.Matches(msg, m.keyMap.modalScrollUp):
			if m.isModalVisible() {
				m.modal.ScrollUp(1)
			}
		case key.Matches(msg, m.keyMap.modalScrollDown):
			if m.isModalVisible() {
				m.modal.ScrollDown(1)
			}
		}
	case tuiv2.SetTextModalMsg:
		if msg.Type == tuiv2.MODAL_TYPE_TRACE {
			if msg.Text == "" {
				km := m.modal.KeyMap
				width, height := m.modal.Width, m.modal.Height
				m.modal = viewport.New(width, height)
				m.modal.KeyMap = km
			} else {
				m.modal.SetContent(tuiv2.PrettyJSONFromKVFormat(msg.Text))
			}
		}
	}

	cmds = append(cmds, m.handleMsg(msg)...)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	// Left
	m.styles.left = m.styles.focusedStyle(m.styles.left, m.svcSpans.Focused())
	left := m.styles.left.Render(m.svcSpans.View())
	left = tuiv2.RenderWithTitle(left, "Traces [t]", m.styles.left)

	// Right
	m.styles.right = m.styles.focusedStyle(m.styles.right, m.details.Focused())
	right := m.styles.right.Render(m.details.View())
	right = tuiv2.RenderWithTitle(right, "Details [d]", m.styles.right)

	layout := lipgloss.JoinHorizontal(lipgloss.Left, left, right)

	modal := ""
	if m.isModalVisible() {
		modal = m.styles.modal.Render(m.modal.View())
		modal = tuiv2.RenderWithTitle(modal, "Scroll (^hjkl)", m.styles.modal)
	}
	layoutWithModal := tuiv2.Composite(modal, layout, tuiv2.Right, tuiv2.Bottom, 0, 0)

	return layoutWithModal
}

func (m *Model) UpdateLayout(width, height int) {
	if width == 0 && height == 0 {
		width = m.width
		height = m.height
	} else {
		m.width = width
		m.height = height
	}

	leftWidth := int(float64(width) * 0.6)
	leftWidth = leftWidth - m.styles.left.GetHorizontalFrameSize() + m.splitOffset
	if leftWidth < MINIMUM_INDIVISUAL_WIDTH {
		leftWidth = MINIMUM_INDIVISUAL_WIDTH
		m.splitOffset += OFFSET_STEP
	}
	rightWidth := width - leftWidth
	if rightWidth < MINIMUM_INDIVISUAL_WIDTH {
		rightWidth = MINIMUM_INDIVISUAL_WIDTH
		leftWidth = width - rightWidth
		m.splitOffset -= OFFSET_STEP
	}
	rightWidth = rightWidth - m.styles.right.GetHorizontalFrameSize()
	height = height - m.styles.left.GetVerticalFrameSize()

	m.styles.left = m.styles.left.Width(leftWidth).Height(height)
	m.styles.right = m.styles.right.Width(rightWidth).Height(height)

	m.svcSpans.UpdateLayout(leftWidth, height)
	m.details.UpdateLayout(rightWidth, height)

	// modal
	m.styles.modal = m.styles.modal.
		Width(rightWidth + 1).
		Height(height / 3)
	m.modal.Width = m.styles.modal.GetWidth() - m.styles.modal.GetHorizontalBorderSize()
	m.modal.Height = m.styles.modal.GetHeight()
}

func (m *Model) handleMsg(msg tea.Msg) (cmds []tea.Cmd) {
	var cmd tea.Cmd

	switch msg.(type) {
	case tuiv2.UpdateHelpKeysMsg:
		cmds = append(cmds, tuiv2.SetTraceHelpKeysCmd(
			m.focus,
			help.COMPONENT_ID_TRACE,
			m.keyMap.ShortHelp(),
		))
	}

	m.svcSpans, cmd = m.svcSpans.Update(msg)
	cmds = append(cmds, cmd)
	m.details, cmd = m.details.Update(msg)
	cmds = append(cmds, cmd)
	m.modal, cmd = m.modal.Update(msg)
	cmds = append(cmds, cmd)

	return cmds
}

func (m *Model) isModalVisible() bool {
	return m.modal.TotalLineCount() > 0
}

func getTraceInfoTreeNode(spans []*telemetry.SpanData) *tree.TreeNode {
	root := &tree.TreeNode{
		Label:    "No data",
		Children: []*tree.TreeNode{},
		Expanded: true,
	}

	if len(spans) == 0 {
		return root
	}
	traceID := spans[0].Span.TraceID().String()
	sname := telemetry.GetServiceNameFromResource(spans[0].ResourceSpan.Resource())
	root.Label = fmt.Sprintf("%s (%s)", sname, traceID)

	// statistics
	statistics := &tree.TreeNode{
		Label:    "Statistics",
		Children: []*tree.TreeNode{},
		Parent:   root,
		Expanded: true,
	}
	root.Children = append(root.Children, statistics)
	statistics.Children = append(statistics.Children, &tree.TreeNode{
		Label:    fmt.Sprintf("span count: %d", len(spans)),
		Parent:   statistics,
		Expanded: true,
	})

	// resource info
	rs := spans[0].ResourceSpan
	r := rs.Resource()
	resource := &tree.TreeNode{
		Label:    "Resource",
		Children: []*tree.TreeNode{},
		Parent:   root,
		Expanded: true,
	}
	root.Children = append(root.Children, resource)
	resource.Children = append(resource.Children, &tree.TreeNode{
		Label:    fmt.Sprintf("dropped attributes count: %d", r.DroppedAttributesCount()),
		Parent:   resource,
		Expanded: true,
	})
	resource.Children = append(resource.Children, &tree.TreeNode{
		Label:    fmt.Sprintf("schema url: %s", rs.SchemaUrl()),
		Parent:   resource,
		Expanded: true,
	})

	attrs := &tree.TreeNode{
		Label:    "Attributes",
		Children: []*tree.TreeNode{},
		Parent:   resource,
		Expanded: true,
	}
	resource.Children = append(resource.Children, attrs)
	appendAttrsSorted(attrs, r.Attributes())

	// scope info
	scopes := &tree.TreeNode{
		Label:    "Scopes",
		Children: []*tree.TreeNode{},
		Parent:   resource,
		Expanded: true,
	}
	resource.Children = append(resource.Children, scopes)
	for si := 0; si < rs.ScopeSpans().Len(); si++ {
		ss := rs.ScopeSpans().At(si)
		scope := &tree.TreeNode{
			Label:    ss.Scope().Name(),
			Children: []*tree.TreeNode{},
			Parent:   scopes,
			Expanded: true,
		}
		scopes.Children = append(scopes.Children, scope)
		scope.Children = append(scope.Children, &tree.TreeNode{
			Label:    fmt.Sprintf("schema url: %s", ss.SchemaUrl()),
			Parent:   scope,
			Expanded: true,
		})
		scope.Children = append(scope.Children, &tree.TreeNode{
			Label:    fmt.Sprintf("version: %s", ss.Scope().Version()),
			Parent:   scope,
			Expanded: true,
		})
		scope.Children = append(scope.Children, &tree.TreeNode{
			Label:    fmt.Sprintf("dropped attributes count: %d", ss.Scope().DroppedAttributesCount()),
			Parent:   scope,
			Expanded: true,
		})

		attrs := &tree.TreeNode{
			Label:    "Attributes",
			Children: []*tree.TreeNode{},
			Parent:   scope,
			Expanded: true,
		}
		scope.Children = append(scope.Children, attrs)
		appendAttrsSorted(attrs, ss.Scope().Attributes())
	}

	return root
}

func appendAttrsSorted(parent *tree.TreeNode, attrs pcommon.Map) {
	keys := make([]string, 0, attrs.Len())
	attrs.Range(func(k string, _ pcommon.Value) bool {
		keys = append(keys, k)
		return true
	})
	sort.Strings(keys)

	for _, k := range keys {
		v, _ := attrs.Get(k)
		parent.Children = append(parent.Children, &tree.TreeNode{
			Label:    fmt.Sprintf("%s: %s", k, v.AsString()),
			Parent:   parent,
			Expanded: true,
		})
	}
}

func closeModalCmd() tea.Cmd {
	return func() tea.Msg {
		return tuiv2.SetTextModalMsg{
			Type: tuiv2.MODAL_TYPE_TRACE,
			Text: "",
		}
	}
}
