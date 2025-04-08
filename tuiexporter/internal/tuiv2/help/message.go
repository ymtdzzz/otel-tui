package help

import "github.com/charmbracelet/bubbles/key"

type ComponentID string

const (
	COMPONENT_ID_TRACE ComponentID = "component-id-trace"
)

type SetTraceHelpKeysMsg struct {
	Visible bool
	ID      ComponentID
	Keys    []key.Binding
}
