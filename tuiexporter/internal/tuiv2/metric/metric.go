package metric

import tea "github.com/charmbracelet/bubbletea"

type Model struct{}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		_    tea.Cmd
		cmds []tea.Cmd
	)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return "This is metric view"
}
