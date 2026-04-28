package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/neto-app/neto/tui/internal/client"
)

// ConfirmResult indicates the user's choice in the confirmation dialog.
type ConfirmResult int

const (
	// ConfirmResultPending means the user has not decided yet.
	ConfirmResultPending ConfirmResult = iota
	// ConfirmResultYes means the user confirmed.
	ConfirmResultYes
	// ConfirmResultNo means the user cancelled.
	ConfirmResultNo
)

// ConfirmModel is a Bubbletea model that presents a confirmation dialog
// when the API returns a pending_confirmation response.
type ConfirmModel struct {
	Pending *client.PendingConfirmation
	Result  ConfirmResult
}

// NewConfirmModel creates a new ConfirmModel for the given pending confirmation.
func NewConfirmModel(p *client.PendingConfirmation) ConfirmModel {
	return ConfirmModel{Pending: p}
}

// Init implements tea.Model.
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "y", "enter":
			m.Result = ConfirmResultYes
		case "n", "esc":
			m.Result = ConfirmResultNo
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m ConfirmModel) View() string {
	if m.Pending == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("┌─────────────────────────────────┐\n")
	sb.WriteString(fmt.Sprintf("│ Confirmar: %-22s│\n", m.Pending.Tool))
	sb.WriteString("├─────────────────────────────────┤\n")
	for k, v := range m.Pending.Preview {
		sb.WriteString(fmt.Sprintf("│  %-13s %16v │\n", k+":", v))
	}
	sb.WriteString("├─────────────────────────────────┤\n")
	sb.WriteString("│  [y/Enter] Confirmar  [n/Esc]  │\n")
	sb.WriteString("└─────────────────────────────────┘\n")
	return sb.String()
}
