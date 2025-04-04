package tree

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2"
	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/tuiv2/help"
)

type OnSelectedFn func(label string) tea.Cmd
type OnKeyInputFn func(keyMsg tea.KeyMsg) tea.Cmd

type TreeNode struct {
	Label    string
	Children []*TreeNode
	Expanded bool
	Parent   *TreeNode
}

type Model struct {
	id         TreeID
	styles     styles
	keyMap     keyMap
	root       *TreeNode
	tree       *tree.Tree
	flattened  []*TreeNode
	cursor     int
	viewport   viewport.Model
	focus      bool
	onSelected OnSelectedFn
	onKeyInput OnKeyInputFn
}

func New(id TreeID, onSelected OnSelectedFn, onKeyInput OnKeyInputFn) Model {
	m := Model{
		id:         id,
		styles:     defaultStyles(),
		keyMap:     defaultKeyMap(),
		tree:       newTree(),
		flattened:  []*TreeNode{},
		cursor:     0,
		viewport:   viewport.New(0, 0),
		focus:      false,
		onSelected: onSelected,
		onKeyInput: onKeyInput,
	}
	m.updateView()

	return m
}

func (m *Model) Focus(focus bool) {
	m.focus = focus
}

func (m Model) Focused() bool {
	return m.focus
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case UpdateTreeMsg:
		m.root = msg.Root
		m.updateView()
	case tuiv2.UpdateHelpKeysMsg:
		switch m.id {
		case TREE_ID_TRACE:
			cmds = append(cmds, tuiv2.SetTraceHelpKeysCmd(
				m.focus,
				help.ComponentID(m.id),
				m.keyMap.ShortHelp(),
			))
		}
	}

	if !m.focus {
		// IF not focused, pass messages other than key messages
		if _, ok := msg.(tea.KeyMsg); !ok {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		onKeyInput := true

		switch {
		case key.Matches(msg, m.keyMap.scrollToTop):
			m.cursor = 0
			m.viewport.GotoTop()
			m.updateView()
		case key.Matches(msg, m.keyMap.scrollToBottom):
			m.cursor = len(m.flattened) - 1
			m.viewport.GotoBottom()
			m.updateView()
		case key.Matches(msg, m.keyMap.halfPageUp):
			m.cursor -= m.viewport.Height / 2
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.viewport.HalfPageUp()
			m.updateView()
		case key.Matches(msg, m.keyMap.halfPageDown):
			m.cursor += m.viewport.Height / 2
			if m.cursor >= len(m.flattened) {
				m.cursor = len(m.flattened) - 1
			}
			m.viewport.HalfPageDown()
			m.updateView()
		case key.Matches(msg, m.keyMap.scrollRight):
			m.viewport.ScrollRight(1)
			m.updateView()
		case key.Matches(msg, m.keyMap.scrollLeft):
			m.viewport.ScrollLeft(1)
			m.updateView()
		case key.Matches(msg, m.keyMap.up):
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.viewport.YOffset+1 {
					m.viewport.SetYOffset(m.viewport.YOffset - 1)
				}
			}
			m.updateView()
		case key.Matches(msg, m.keyMap.down):
			if m.cursor < len(m.flattened)-1 {
				m.cursor++
				if m.cursor >= m.viewport.Height+m.viewport.YOffset-1 {
					m.viewport.SetYOffset(m.viewport.YOffset + 1)
				}
			}
			m.updateView()
		case key.Matches(msg, m.keyMap.selectNode):
			node := m.flattened[m.cursor]
			node.Expanded = !node.Expanded
			if len(node.Children) == 0 {
				if m.onSelected != nil {
					cmds = append(cmds, m.onSelected(node.Label))
				}
			}
			m.updateView()
			onKeyInput = false
		}

		if onKeyInput {
			if m.onKeyInput != nil {
				cmds = append(cmds, m.onKeyInput(msg))
			}
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if len(m.flattened) == 0 {
		m.viewport.SetContent("No data")
	}
	return m.viewport.View()
}

func (m *Model) updateView() {
	m.flattenAll(m.root, "")
	m.viewport.SetContent(m.tree.String())
}

func (m *Model) UpdateLayout(width, height int) {
	m.viewport.Width = width
	m.viewport.Height = height
}

func (m *Model) flattenAll(root *TreeNode, _ string) {
	flattened := []*TreeNode{}
	if root == nil {
		m.tree, m.flattened = newTree(), flattened
		return
	}

	t := newTree().Root(m.renderLabel(root.Label, 0))
	flattened = append(flattened, root)
	if root.Expanded {
		t = m.flatten(root, t, &flattened)
	}

	m.tree, m.flattened = t, flattened
}

func (m *Model) flatten(n *TreeNode, t *tree.Tree, acc *[]*TreeNode) *tree.Tree {
	children := make([]any, 0, len(n.Children))
	for _, child := range n.Children {
		*acc = append(*acc, child)
		if !child.Expanded || len(child.Children) == 0 {
			children = append(children, m.renderLabel(child.Label, len(*acc)-1))
		} else {
			subTree := newTree().Root(m.renderLabel(child.Label, len(*acc)-1))
			subTree = m.flatten(child, subTree, acc)
			children = append(children, subTree)
		}
	}

	return t.Child(children...)
}

func (m Model) renderLabel(label string, cur int) string {
	if cur == m.cursor {
		return m.styles.focusedStyle.Render(label)
	}
	return label
}

func newTree() *tree.Tree {
	return tree.New().
		Enumerator(tree.RoundedEnumerator)
}
