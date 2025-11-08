package layout

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
					Key:         tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
					Arrow:       false,
					Description: "quit",
				},
			},
			want: " [yellow]q[white]: quit",
		},
		{
			name: "single arrow key",
			keymaps: KeyMaps{
				{
					Key:         tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone),
					Arrow:       true,
					Description: "navigate",
				},
			},
			want: " [yellow]→←↑↓[white]: navigate",
		},
		{
			name: "multiple keys",
			keymaps: KeyMaps{
				{
					Key:         tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
					Arrow:       false,
					Description: "quit",
				},
				{
					Key:         tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),
					Arrow:       false,
					Description: "",
				},
				{
					Key:         tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone),
					Arrow:       true,
					Description: "navigate",
				},
				{
					Key:         tcell.NewEventKey(tcell.KeyRight, 'b', tcell.ModNone),
					Arrow:       true,
					Description: "hidden key",
					Hidden:      true,
				},
				{
					Key:         tcell.NewEventKey(tcell.KeyCtrlH, ' ', tcell.ModNone),
					Arrow:       false,
					Description: "narrow width",
				},
			},
			want: " [yellow]q[white]: quit | [yellow]→←↑↓[white]: navigate | [yellow]Ctrl-H[white]: narrow width",
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
	cmdList := NewCommandList()
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
			got := AttachCommandList(tt.commands, tt.primitive)
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

type MockFocusableBox struct {
	mock.Mock
	*tview.Box
}

func NewMockFocusableBox() *MockFocusableBox {
	return &MockFocusableBox{Box: tview.NewBox()}
}

// func (m *MockFocusableBox) SetFocusFunc(fn func()) *tview.Box {
// 	m.Called(fn)
// 	return m.Box
// }
//
// func (m *MockFocusableBox) SetInputCapture(fn func(*tcell.EventKey) *tcell.EventKey) *tview.Box {
// 	m.Called(fn)
// 	return m.Box
// }

func (m *MockFocusableBox) Handle(event *tcell.EventKey) *tcell.EventKey {
	m.Called(event)
	return nil
}

func (m *MockFocusableBox) OriginalOnFocus() {
	m.Called()
}

func TestRegisterCommandList(t *testing.T) {
	t.Run("commands is nil", func(t *testing.T) {
		mockBox := NewMockFocusableBox()
		RegisterCommandList(nil, mockBox, nil, nil)

		mockBox.AssertNotCalled(t, "SetFocusFunc")
		mockBox.AssertNotCalled(t, "SetInputCapture")
	})

	t.Run("commands and keymaps exist", func(t *testing.T) {
		tests := []struct {
			name               string
			key                tcell.EventKey
			desc               string
			setOriginalOnFocus bool
			wantCommandText    string
		}{
			{
				name:               "quit by q key and call original on focus",
				key:                *tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
				desc:               "Quit",
				setOriginalOnFocus: true,
				wantCommandText:    "[yellow]q[white]: Quit",
			},
			{
				name:               "Refresh by Ctrl-R key without original on focus",
				key:                *tcell.NewEventKey(tcell.KeyCtrlR, ' ', tcell.ModNone),
				desc:               "Refresh",
				setOriginalOnFocus: false,
				wantCommandText:    "[yellow]Ctrl-R[white]: Refresh",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mockBox := NewMockFocusableBox()
				commands := tview.NewTextView()
				keys := KeyMaps{
					{
						Key:         &tt.key,
						Description: tt.desc,
						Handler:     mockBox.Handle,
					},
				}

				mockBox.On("Handle", mock.AnythingOfType("*tcell.EventKey")).Once()
				if tt.setOriginalOnFocus {
					mockBox.On("OriginalOnFocus").Once()
					RegisterCommandList(commands, mockBox, mockBox.OriginalOnFocus, keys)
				} else {
					RegisterCommandList(commands, mockBox, nil, keys)
				}

				mockBox.Focus(nil)
				handler := mockBox.InputHandler()
				handler(&tt.key, nil)

				mockBox.AssertNumberOfCalls(t, "Handle", 1)
				if tt.setOriginalOnFocus {
					mockBox.AssertNumberOfCalls(t, "OriginalOnFocus", 1)
				} else {
					mockBox.AssertNotCalled(t, "OriginalOnFocus")
				}
				assert.Contains(t, commands.GetText(false), tt.wantCommandText)
			})
		}

		t.Run("merged keymaps", func(t *testing.T) {
			mockBox := NewMockFocusableBox()
			commands := tview.NewTextView()
			keys1 := KeyMaps{
				{
					Key:         tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
					Description: "Quit",
					Handler:     mockBox.Handle,
				},
			}
			keys2 := KeyMaps{
				{
					Key:         tcell.NewEventKey(tcell.KeyCtrlR, ' ', tcell.ModNone),
					Description: "Refresh",
					Handler:     mockBox.Handle,
				},
			}
			keys1.Merge(keys2)

			mockBox.On("Handle", mock.AnythingOfType("*tcell.EventKey")).Twice()
			RegisterCommandList(commands, mockBox, nil, keys1)

			mockBox.Focus(nil)
			handler := mockBox.InputHandler()
			handler(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone), nil)
			handler(tcell.NewEventKey(tcell.KeyCtrlR, ' ', tcell.ModNone), nil)

			mockBox.AssertNumberOfCalls(t, "Handle", 2)
			assert.Equal(t, " [yellow]q[white]: Quit | [yellow]Ctrl-R[white]: Refresh", commands.GetText(false))
		})

		t.Run("no matching key in keymaps", func(t *testing.T) {
			mockBox := NewMockFocusableBox()
			commands := tview.NewTextView()
			keys := KeyMaps{
				{
					Key:         tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
					Description: "Quit",
					Handler:     mockBox.Handle,
				},
			}

			RegisterCommandList(commands, mockBox, nil, keys)

			mockBox.Focus(nil)
			handler := mockBox.InputHandler()
			handler(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone), nil)

			mockBox.AssertNotCalled(t, "Handle")
		})
	})
}
