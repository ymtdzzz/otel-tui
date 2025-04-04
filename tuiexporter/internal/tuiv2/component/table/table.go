package table

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/help"
)

type Model[T any] struct {
	id         TableID
	styles     styles
	keyMap     keyMap
	focus      bool
	focusQuery bool
	queryInput textinput.Model
	query      string
	viewport   viewport.Model
	table      table.Model
	data       *[]T
	mapper     CellMappers[T]
	onChange   func(selected T, idx int) tea.Cmd
}

func New[T any](
	id TableID,
	data *[]T,
	mapper CellMappers[T],
	onChange func(selected T, idx int) tea.Cmd,
) Model[T] {
	keyMap := defaultKeyMap()
	styles := defaultStyles()

	t := table.New()

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.headerBorderColor).
		BorderTop(true).
		BorderBottom(true).
		Bold(true)

	t.SetStyles(s)

	vp := viewport.New(0, 0)
	vp.SetHorizontalStep(1)

	qi := textinput.New()
	qi.Placeholder = keyMap.focusInput.Keys()[0] + " to query"
	qi.Prompt = " Query: "
	qi.Width = 20

	m := Model[T]{
		id:         id,
		styles:     styles,
		keyMap:     keyMap,
		queryInput: qi,
		viewport:   vp,
		table:      t,
		data:       data,
		mapper:     mapper,
		onChange:   onChange,
	}
	m.updateTable()

	return m
}

func (m *Model[T]) Focus(focus bool) {
	m.focus = focus
	if focus {
		m.table.Focus()
	} else {
		m.table.Blur()
	}
}

func (m Model[T]) Focused() bool {
	return m.focus
}

func (m Model[T]) QueryFocused() bool {
	return m.focusQuery
}

func (m Model[T]) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
	)
}

func (m Model[T]) Update(msg tea.Msg) (Model[T], tea.Cmd) {
	var (
		_    tea.Cmd
		cmds []tea.Cmd
		curr = m.table.Cursor()
	)

	switch msg := msg.(type) {
	case UpdateTableMsg:
		if m.id == msg.ID {
			m.updateTable()
		}
	case tuiv2.UpdateHelpKeysMsg:
		switch m.id {
		case TABLE_ID_TRACE:
			cmds = append(cmds, tuiv2.SetTraceHelpKeysCmd(
				m.focus,
				help.ComponentID(m.id),
				m.keyMap.ShortHelp(m.focusQuery),
			))
		}
	}

	if !m.focus {
		// IF not focused, pass messages other than key messages
		if _, ok := msg.(tea.KeyMsg); !ok {
			cmds = append(cmds, m.handleMsg(msg, curr)...)
		}
		return m, tea.Batch(cmds...)
	}

	m.viewport.SetContent(m.table.View())
	cmds = append(cmds, m.handleMsg(msg, curr)...)

	return m, tea.Batch(cmds...)
}

func (m Model[T]) View() string {
	m.viewport.SetContent(m.table.View())
	table := m.viewport.View()

	return lipgloss.JoinVertical(lipgloss.Top, m.queryInput.View(), table)
}

func (m *Model[T]) UpdateLayout(width, height int) {
	height -= 1

	m.table.SetHeight(height)

	m.viewport.Width = width
	m.viewport.Height = height
}

func (m *Model[T]) handleMsg(msg tea.Msg, curr int) (cmds []tea.Cmd) {
	var cmd tea.Cmd

	if m.focusQuery {
		m.queryInput, cmd = m.queryInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.String() == "/":
			m.focusQuery = true
			m.queryInput.Focus()
			m.table.Blur()
			cmds = append(cmds, tuiv2.UpdateHelpKeysCmd())
		case msg.String() == "esc":
			m.focusQuery = false
			m.queryInput.Blur()
			m.table.Focus()
			cmds = append(cmds, tuiv2.UpdateHelpKeysCmd())
		case msg.String() == "enter":
			m.focusQuery = false
			m.queryInput.Blur()
			m.table.Focus()
			m.query = m.queryInput.Value()
			cmds = append(cmds, func() tea.Msg {
				return tuiv2.ApplySpanFilterMsg{
					Query: m.query,
				}
			})
			cmds = append(cmds, tuiv2.UpdateHelpKeysCmd())
		}
	}

	if !m.focusQuery {
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	next := m.table.Cursor()
	if curr != next && curr >= 0 && curr < len(*m.data) {
		cmds = append(cmds, m.onChange((*m.data)[curr], next))
	}

	return cmds
}

func (m *Model[T]) updateTable() {
	maxlen := make([]int, len(m.mapper))
	rows := make([]table.Row, len((*m.data)))
	for i := range *m.data {
		rows[i], maxlen = m.toRow(i, maxlen)
	}
	cols := getColumns(m.mapper, maxlen)
	m.table.SetColumns(m.adjustColWidth(cols))
	m.table.SetRows(rows)
	if len(m.table.Rows()) == 0 {
		m.viewport.SetContent("No data")
	}
}

func (m Model[T]) toRow(row int, maxlen []int) (table.Row, []int) {
	if row >= 0 && row < len(*(m.data)) {
		rowData := make(table.Row, len(m.mapper))
		d := (*m.data)[row]
		for i, m := range m.mapper {
			rowData[i] = m.GetTextRowFn(d)
			if len(rowData[i]) > maxlen[i] {
				maxlen[i] = len(rowData[i])
			}
		}
		return rowData, maxlen
	}
	return table.Row{}, maxlen
}

func (m Model[T]) adjustColWidth(cols []table.Column) []table.Column {
	totalWidth := 0
	for _, c := range cols {
		totalWidth += c.Width
	}
	if totalWidth < m.viewport.Width {
		cols[len(cols)-1].Width = m.viewport.Width - totalWidth
	}
	return cols
}
