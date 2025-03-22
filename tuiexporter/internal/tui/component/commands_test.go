package component

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestKeyMaps_keyTexts(t *testing.T) {
	tests := []struct {
		name    string
		keymaps KeyMaps
		want    string
	}{
		{
			name: "single non-arrow key",
			keymaps: KeyMaps{
				{
					key:         tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
					arrow:       false,
					description: "quit",
				},
			},
			want: " [yellow]q[white]: quit",
		},
		{
			name: "single arrow key",
			keymaps: KeyMaps{
				{
					key:         tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone),
					arrow:       true,
					description: "navigate",
				},
			},
			want: " [yellow]→←↑↓[white]: navigate",
		},
		{
			name: "multiple keys",
			keymaps: KeyMaps{
				{
					key:         tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
					arrow:       false,
					description: "quit",
				},
				{
					key:         tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone),
					arrow:       true,
					description: "navigate",
				},
			},
			want: " [yellow]q[white]: quit | [yellow]→←↑↓[white]: navigate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.keymaps.keyTexts()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewCommandList(t *testing.T) {
	cmdList := newCommandList()
	assert.NotNil(t, cmdList)
}

func TestAttachCommandList(t *testing.T) {
	tests := []struct {
		name      string
		commands  *tview.TextView
		primitive tview.Primitive
		wantItems int
	}{
		{
			name:      "with nil commands",
			commands:  nil,
			primitive: tview.NewBox(),
			wantItems: 0,
		},
		{
			name:      "with commands and primitive",
			commands:  tview.NewTextView(),
			primitive: tview.NewBox(),
			wantItems: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := attachCommandList(tt.commands, tt.primitive)
			assert.NotNil(t, got)
			assert.Equal(t, tt.wantItems, got.GetItemCount())

			if tt.wantItems > 0 {
				item := got.GetItem(0)
				assert.Equal(t, tt.primitive, item)

				commands := got.GetItem(1)
				assert.Equal(t, tt.commands, commands)
			}
		})
	}
}

type mockFocusable struct {
	*tview.Box
	focusFunc func()
}

func newMockFocusable() *mockFocusable {
	return &mockFocusable{
		Box: tview.NewBox(),
	}
}

func (m *mockFocusable) SetFocusFunc(fn func()) *tview.Box {
	m.focusFunc = fn
	return m.Box
}

func TestRegisterCommandList(t *testing.T) {
	tests := []struct {
		name     string
		commands *tview.TextView
		keys     KeyMaps
	}{
		{
			name:     "with nil commands",
			commands: nil,
			keys:     KeyMaps{},
		},
		{
			name:     "with valid commands",
			commands: tview.NewTextView(),
			keys: KeyMaps{
				{
					key:         tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
					arrow:       false,
					description: "quit",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			focusableMock := newMockFocusable()
			origFocusFn := func() {}
			registerCommandList(tt.commands, focusableMock, origFocusFn, tt.keys)

			if tt.commands == nil {
				// focusableMock should not have focus function
				assert.Nil(t, focusableMock.focusFunc)
				return
			}

			assert.NotNil(t, focusableMock.focusFunc)

			focusableMock.focusFunc()
			text := tt.commands.GetText(false)
			assert.Equal(t, tt.keys.keyTexts(), text)
		})
	}
}
