package app

import (
	"strings"

	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/component/table"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/help"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/metric"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/trace"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	TAB_TRACE = iota
	TAB_METRIC
)

type Model struct {
	keyMap        keyMap
	currentTab    int
	width, height int
	store         *telemetry.Store
	trace         trace.Model
	metric        metric.Model
	help          help.Model
}

func New(store *telemetry.Store) tea.Model {
	trace := trace.New(store)
	trace.Focus(true)

	m := Model{
		keyMap:     defaultKeyMap(),
		currentTab: TAB_TRACE,
		store:      store,
		trace:      trace,
		metric:     metric.New(),
	}
	m.help = help.New(m.keyMap.ShortHelp())

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.trace.Init(),
		m.metric.Init(),
		m.help.Init(),
		tuiv2.UpdateHelpKeysCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.changeTabKey):
			cmds = append(cmds, func() tea.Msg {
				return rotateTabMsg{}
			})
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		helpHeight := 1
		m.trace.UpdateLayout(msg.Width, msg.Height-lipgloss.Height(m.renderTab())-helpHeight)
	case rotateTabMsg:
		cmds = append(cmds, m.rotateTab())
	case PushTracesMsg:
		m.store.AddSpan(msg.Traces)
		cmds = append(cmds, func() tea.Msg {
			return table.UpdateTableMsg{
				ID: table.TABLE_ID_TRACE,
			}
		})
	case tuiv2.ApplySpanFilterMsg:
		cmds = append(cmds, func() tea.Msg {
			m.store.ApplyFilterTraces(msg.Query, telemetry.SORT_TYPE_NONE)
			return table.UpdateTableMsg{
				ID: table.TABLE_ID_TRACE,
			}
		})
	}

	m.trace, cmd = m.trace.Update(msg)
	cmds = append(cmds, cmd)
	m.metric, cmd = m.metric.Update(msg)
	cmds = append(cmds, cmd)
	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	help := ""
	switch m.currentTab {
	case TAB_TRACE:
		help = m.help.TraceView()
	case TAB_METRIC:
		help = m.help.MetricView()
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.renderTab(),
		m.renderMain(),
		help,
	)
}

func (m Model) renderTab() string {
	var traceTab, metricTab string

	switch m.currentTab {
	case TAB_TRACE:
		traceTab = activeTab.Render("Trace")
		metricTab = tab.Render("Metric")
	case TAB_METRIC:
		traceTab = tab.Render("Trace")
		metricTab = activeTab.Render("Metric")
	}

	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		traceTab,
		metricTab,
	)
	gap := tabGap.Render(strings.Repeat(" ", max(0, m.width-lipgloss.Width(row)-2)))

	return lipgloss.JoinHorizontal(lipgloss.Bottom, traceTab, metricTab, gap)
}

func (m Model) renderMain() string {
	switch m.currentTab {
	case TAB_TRACE:
		return m.trace.View()
	case TAB_METRIC:
		return m.metric.View()
	}

	return ""
}

func (m *Model) rotateTab() tea.Cmd {
	switch m.currentTab {
	case TAB_TRACE:
		m.currentTab = TAB_METRIC
		m.trace.Focus(false)
	case TAB_METRIC:
		m.currentTab = TAB_TRACE
		m.trace.Focus(true)
	}
	return tuiv2.UpdateHelpKeysCmd()
}
