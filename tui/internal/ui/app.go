// Package ui contains all Bubbletea models and views for the Neto TUI.
package ui

import (
	"errors"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/neto-app/neto/tui/internal/client"
	"github.com/neto-app/neto/tui/internal/config"
)

// appState represents which screen is currently displayed.
type appState int

const (
	stateSetup appState = iota
	stateLogin
	stateChat
)

// App is the root Bubbletea model. It owns the state machine and delegates
// rendering to child models (setup, login, chat).
type App struct {
	state  appState
	cfg    *config.Config
	client *client.Client
	setup  SetupModel
	login  LoginModel
	chat   ChatModel
	width  int
	height int
}

// NewApp creates and initialises the root App model.
func NewApp() App {
	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, config.ErrNotConfigured) {
			return App{
				state: stateSetup,
				setup: NewSetupModel(),
			}
		}
		// Unexpected error — show setup so the user can reconfigure.
		return App{
			state: stateSetup,
			setup: NewSetupModel(),
		}
	}

	c := client.New(cfg, func(accessToken, refreshToken string) error {
		cfg.AccessToken = accessToken
		cfg.RefreshToken = refreshToken
		return config.Save(cfg)
	})
	if cfg.AccessToken != "" {
		return App{
			state:  stateChat,
			cfg:    cfg,
			client: c,
			chat:   NewChatModel(c, cfg),
		}
	}

	return App{
		state:  stateLogin,
		cfg:    cfg,
		client: c,
		login:  NewLoginModel(c, cfg),
	}
}

// Init implements tea.Model.
func (a App) Init() tea.Cmd {
	switch a.state {
	case stateSetup:
		return a.setup.Init()
	case stateLogin:
		return a.login.Init()
	case stateChat:
		return a.chat.Init()
	}
	return nil
}

// Update implements tea.Model.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Propagate inner width (descontando padding horizontal).
		inner := msg
		inner.Width = msg.Width - 2*paddingX
		switch a.state {
		case stateSetup:
			m, cmd := a.setup.Update(inner)
			a.setup = m.(SetupModel)
			return a, cmd
		case stateLogin:
			m, cmd := a.login.Update(inner)
			a.login = m.(LoginModel)
			return a, cmd
		case stateChat:
			m, cmd := a.chat.Update(inner)
			a.chat = m.(ChatModel)
			return a, cmd
		}

	case SetupDoneMsg:
		// User finished setup — move to login.
		a.cfg = msg.Config
		a.client = client.New(a.cfg, func(accessToken, refreshToken string) error {
			a.cfg.AccessToken = accessToken
			a.cfg.RefreshToken = refreshToken
			return config.Save(a.cfg)
		})
		a.state = stateLogin
		a.login = NewLoginModel(a.client, a.cfg)
		return a, a.login.Init()

	case LoginSuccessMsg:
		// User logged in — move to chat.
		a.cfg.AccessToken = msg.AccessToken
		a.cfg.RefreshToken = msg.RefreshToken
		_ = config.Save(a.cfg)
		a.client.SetToken(a.cfg.AccessToken)
		a.client.SetRefreshToken(a.cfg.RefreshToken)
		a.state = stateChat
		a.chat = NewChatModel(a.client, a.cfg)
		return a, a.chat.Init()
	}

	// Delegate to active child.
	switch a.state {
	case stateSetup:
		m, cmd := a.setup.Update(msg)
		a.setup = m.(SetupModel)
		return a, cmd
	case stateLogin:
		m, cmd := a.login.Update(msg)
		a.login = m.(LoginModel)
		return a, cmd
	case stateChat:
		m, cmd := a.chat.Update(msg)
		a.chat = m.(ChatModel)
		return a, cmd
	}

	return a, nil
}

// View implements tea.Model.
func (a App) View() string {
	var content string
	switch a.state {
	case stateSetup:
		content = a.setup.View()
	case stateLogin:
		content = a.login.View()
	case stateChat:
		content = a.chat.View()
	}
	return lipgloss.NewStyle().PaddingLeft(paddingX).PaddingRight(paddingX).Render(content)
}

// SetupDoneMsg is emitted by the setup screen when configuration is saved.
type SetupDoneMsg struct {
	Config *config.Config
}

// LoginSuccessMsg is emitted by the login screen when authentication succeeds.
type LoginSuccessMsg struct {
	AccessToken  string
	RefreshToken string
}

// SetupModel is the first-run configuration screen.
type SetupModel struct {
	inputs  []textinput.Model
	focused int
	err     string
}

// NewSetupModel creates a setup screen with three text inputs.
func NewSetupModel() SetupModel {
	fields := []struct{ placeholder, value string }{
		{"API URL (e.g. http://localhost:8080)", "http://localhost:8080"},
		{"Supabase URL (e.g. https://xxx.supabase.co)", ""},
		{"Supabase Anon Key", ""},
	}

	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.placeholder
		ti.SetValue(f.value)
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}
	return SetupModel{inputs: inputs}
}

// Init implements tea.Model.
func (m SetupModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "down":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.inputs)
			m.inputs[m.focused].Focus()
		case "shift+tab", "up":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused - 1 + len(m.inputs)) % len(m.inputs)
			m.inputs[m.focused].Focus()
		case "enter":
			if m.focused < len(m.inputs)-1 {
				m.inputs[m.focused].Blur()
				m.focused++
				m.inputs[m.focused].Focus()
			} else {
				// Validate and save
				apiURL := m.inputs[0].Value()
				supURL := m.inputs[1].Value()
				supKey := m.inputs[2].Value()
				if apiURL == "" || supURL == "" || supKey == "" {
					m.err = "All fields are required"
					return m, nil
				}
				cfg := &config.Config{
					APIURL:          apiURL,
					SupabaseURL:     supURL,
					SupabaseAnonKey: supKey,
				}
				if err := config.Save(cfg); err != nil {
					m.err = "Failed to save config: " + err.Error()
					return m, nil
				}
				return m, func() tea.Msg { return SetupDoneMsg{Config: cfg} }
			}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m SetupModel) View() string {
	s := "Neto — First-time setup\n\n"
	labels := []string{"API URL:", "Supabase URL:", "Supabase Anon Key:"}
	for i, inp := range m.inputs {
		s += labels[i] + "\n" + inp.View() + "\n\n"
	}
	if m.err != "" {
		s += "Error: " + m.err + "\n"
	}
	s += "\nTab/↓ next field • Enter submit • Ctrl+C quit"
	return s
}

// newSpinner returns a configured dots spinner.
func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return s
}
