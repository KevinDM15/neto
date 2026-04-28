package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/neto-app/neto/tui/internal/client"
	"github.com/neto-app/neto/tui/internal/config"
)

// LoginModel is the Bubbletea model for the email/password login screen.
type LoginModel struct {
	client   *client.Client
	cfg      *config.Config
	email    textinput.Model
	password textinput.Model
	focused  int // 0 = email, 1 = password
	loading  bool
	spinner  spinner.Model
	err      string
}

// NewLoginModel creates a new LoginModel.
func NewLoginModel(c *client.Client, cfg *config.Config) LoginModel {
	email := textinput.New()
	email.Placeholder = "email@example.com"
	email.Focus()

	pass := textinput.New()
	pass.Placeholder = "password"
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = '•'

	return LoginModel{
		client:   c,
		cfg:      cfg,
		email:    email,
		password: pass,
		spinner:  newSpinner(),
	}
}

// loginResultMsg carries the result of a login attempt.
type loginResultMsg struct {
	accessToken  string
	refreshToken string
	err          error
}

// Init implements tea.Model.
func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "down":
			if m.focused == 0 {
				m.email.Blur()
				m.focused = 1
				m.password.Focus()
			}
		case "shift+tab", "up":
			if m.focused == 1 {
				m.password.Blur()
				m.focused = 0
				m.email.Focus()
			}
		case "enter":
			if m.focused == 0 {
				m.email.Blur()
				m.focused = 1
				m.password.Focus()
				return m, nil
			}
			// Submit
			m.loading = true
			m.err = ""
			email := m.email.Value()
			pass := m.password.Value()
			c := m.client
			return m, tea.Batch(
				m.spinner.Tick,
				func() tea.Msg {
					at, rt, err := c.Login(context.Background(), email, pass)
					return loginResultMsg{accessToken: at, refreshToken: rt, err: err}
				},
			)
		}

	case loginResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		return m, func() tea.Msg {
			return LoginSuccessMsg{AccessToken: msg.accessToken, RefreshToken: msg.refreshToken}
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	if m.focused == 0 {
		m.email, cmd = m.email.Update(msg)
	} else {
		m.password, cmd = m.password.Update(msg)
	}
	return m, cmd
}

// View implements tea.Model.
func (m LoginModel) View() string {
	if m.loading {
		return fmt.Sprintf("\n  %s Logging in…\n", m.spinner.View())
	}

	s := "Neto — Login\n\n"
	s += "Email:\n" + m.email.View() + "\n\n"
	s += "Password:\n" + m.password.View() + "\n\n"
	if m.err != "" {
		s += "Error: " + m.err + "\n\n"
	}
	s += "Tab/↓ next field • Enter submit • Ctrl+C quit"
	return s
}
