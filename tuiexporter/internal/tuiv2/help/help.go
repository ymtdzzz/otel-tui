package help

import (
	"maps"
	"slices"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type componentKeyMap struct {
	visible bool
	keys    []key.Binding
}

type componentKeys map[string]componentKeyMap

func (k componentKeys) keys() []key.Binding {
	ks := make([]key.Binding, 0, len(k))
	for _, ky := range slices.Sorted(maps.Keys(k)) {
		if k[ky].visible {
			ks = append(ks, k[ky].keys...)
		}
	}

	return ks
}

type keys []key.Binding

func (g keys) ShortHelp() []key.Binding {
	return g
}

// FullHelp is not currently used
func (g keys) FullHelp() [][]key.Binding {
	return nil
}

type Model struct {
	h          help.Model
	globalKeys keys
	traceKeys  componentKeys
	metricKeys componentKeys
}

func New(globalKeys []key.Binding) Model {
	h := help.New()
	h.ShowAll = false

	return Model{
		h:          h,
		globalKeys: globalKeys,
		traceKeys:  componentKeys{},
		metricKeys: componentKeys{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetTraceHelpKeysMsg:
		m.traceKeys[string(msg.ID)] = componentKeyMap{
			visible: msg.Visible,
			keys:    msg.Keys,
		}
	}

	return m, nil
}

func (m Model) TraceView() string {
	keys := m.globalKeys
	keys = append(keys, m.traceKeys.keys()...)

	return m.h.View(keys)
}

func (m Model) MetricView() string {
	keys := m.globalKeys
	keys = append(keys, m.metricKeys.keys()...)

	return m.h.View(keys)
}
