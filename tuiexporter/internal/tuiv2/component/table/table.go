package table

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type Model[T any] struct {
	keyMap   keyMap
	focus    bool
	viewport viewport.Model
	table    table.Model
	data     *[]T
	mapper   CellMappers[T]
	onChange func(selected T, idx int) tea.Cmd
}

func New[T any](
	data *[]T,
	mapper CellMappers[T],
	onChange func(selected T, idx int) tea.Cmd,
) Model[T] {
	t := table.New()

	m := Model[T]{
		keyMap:   defaultKeyMap(),
		viewport: viewport.New(0, 0),
		table:    t,
		data:     data,
		mapper:   mapper,
		onChange: onChange,
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

func (m Model[T]) Init() tea.Cmd {
	return nil
}

func (m Model[T]) Update(msg tea.Msg) (Model[T], tea.Cmd) {
	var (
		_    tea.Cmd
		cmds []tea.Cmd
		curr = m.table.Cursor()
	)

	switch msg.(type) {
	case UpdateTableMsg:
		m.updateTable()
	}

	if !m.focus {
		// IF not focused, pass messages other than key messages
		if _, ok := msg.(tea.KeyMsg); !ok {
			cmds = append(cmds, m.handleMsg(msg, curr)...)
		}
		return m, tea.Batch(cmds...)
	}

	cmds = append(cmds, m.handleMsg(msg, curr)...)

	return m, tea.Batch(cmds...)
}

func (m Model[T]) View() string {
	m.viewport.SetContent(m.table.View())
	return m.viewport.View()
}

func (m *Model[T]) UpdateLayout(width, height int) {
	m.table.SetHeight(height)

	m.viewport.Width = width
	m.viewport.Height = height
}

func (m *Model[T]) handleMsg(msg tea.Msg, curr int) (cmds []tea.Cmd) {
	var cmd tea.Cmd

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

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
	m.table.SetColumns(getColumns(m.mapper, maxlen))
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
