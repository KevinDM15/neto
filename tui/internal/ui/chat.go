package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/neto-app/neto/tui/internal/client"
	"github.com/neto-app/neto/tui/internal/config"
)

const helpText = `
Comandos de ejemplo:
  Gasté 50k en luz          — registrar gasto
  ¿Cuánto gasté este mes?   — consultar gastos
  Mostrar cuentas           — listar cuentas
  ¿Cuánto debo?             — ver deudas
  Meta de ahorro 100k       — crear meta

Enter envía  •  Alt+Enter nueva línea  •  Ctrl+H ayuda  •  Ctrl+Q salir
`

// maxInputLines is the maximum number of visible lines in the textarea.
const maxInputLines = 4

// chatMessage holds a single message in the conversation history.
type chatMessage struct {
	role    string // "user" or "assistant"
	content string
}

// ChatModel is the Bubbletea model for the main chat interface.
type ChatModel struct {
	client         *client.Client
	cfg            *config.Config
	messages       []chatMessage
	viewport       viewport.Model
	input          textarea.Model
	spinner        spinner.Model
	loading        bool
	err            string
	showHelp       bool
	conversationID string
	confirm        *ConfirmModel
	width          int
	height         int
}

// NewChatModel creates a new ChatModel.
func NewChatModel(c *client.Client, cfg *config.Config) ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Escribe un mensaje… (Enter envía, Alt+Enter nueva línea)"
	ta.ShowLineNumbers = false
	ta.SetHeight(1)

	// Strip all default styling so it inherits the terminal's dark background.
	noStyle := lipgloss.NewStyle()
	ta.FocusedStyle.Base = noStyle
	ta.FocusedStyle.CursorLine = noStyle
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(colorMuted)
	ta.BlurredStyle.Base = noStyle
	ta.BlurredStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(colorMuted)

	ta.Focus()

	vp := viewport.New(80, 20)
	vp.SetContent("")

	return ChatModel{
		client:   c,
		cfg:      cfg,
		input:    ta,
		viewport: vp,
		spinner:  newSpinner(),
	}
}

// chatResponseMsg wraps the API chat response.
type chatResponseMsg struct {
	resp *client.ChatResponse
	err  error
}

// Init implements tea.Model.
func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update implements tea.Model.
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.input.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		// Confirmation overlay takes priority.
		if m.confirm != nil {
			updated, cmd := m.confirm.Update(msg)
			cm := updated.(ConfirmModel)
			m.confirm = &cm

			switch cm.Result {
			case ConfirmResultYes:
				m.confirm = nil
				return m, m.sendConfirm(cm.Pending)
			case ConfirmResultNo:
				m.confirm = nil
				m.appendMsg("assistant", "Acción cancelada.")
			}
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "ctrl+h":
			m.showHelp = !m.showHelp
			return m, nil
		case "esc":
			m.err = ""
			m.showHelp = false
			return m, nil
		case "alt+enter":
			// Insertar nueva línea en el textarea.
			var taCmd tea.Cmd
			syntheticEnter := tea.KeyMsg{Type: tea.KeyEnter}
			m.input, taCmd = m.input.Update(syntheticEnter)
			return m, taCmd
		case "enter":
			if m.loading || strings.TrimSpace(m.input.Value()) == "" {
				return m, nil
			}
			text := strings.TrimSpace(m.input.Value())
			m.input.Reset()
			m.input.SetHeight(1)
			m.appendMsg("user", text)
			m.loading = true
			m.err = ""
			return m, tea.Batch(m.spinner.Tick, m.sendChat(text))
		}

	case chatResponseMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.conversationID = msg.resp.ConversationID
		if msg.resp.PendingConfirmation != nil {
			cm := NewConfirmModel(msg.resp.PendingConfirmation)
			m.confirm = &cm
			m.appendMsg("assistant", msg.resp.Reply)
			return m, cm.Init()
		}
		m.appendMsg("assistant", msg.resp.Reply)
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	// Viewport scrolling.
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)

	// Input — grow height as user types, up to maxInputLines.
	var taCmd tea.Cmd
	if !m.loading {
		m.input, taCmd = m.input.Update(msg)
		lines := strings.Count(m.input.Value(), "\n") + 1
		if lines > maxInputLines {
			lines = maxInputLines
		}
		if lines < 1 {
			lines = 1
		}
		m.input.SetHeight(lines)
	}

	return m, tea.Batch(vpCmd, taCmd)
}

// View implements tea.Model.
func (m ChatModel) View() string {
	if m.showHelp {
		return helpText
	}

	inputLines := strings.Count(m.input.Value(), "\n") + 1
	if inputLines > maxInputLines {
		inputLines = maxInputLines
	}

	// chrome = header(1) + sep(1) + sep(1) + input(n) + bottom sep(1)
	// +1 when there is a spinner or error line between the separator and input.
	chrome := 4 + inputLines
	if m.loading || m.err != "" {
		chrome++
	}
	maxVP := m.height - chrome
	if maxVP < 1 {
		maxVP = 1
	}

	content := m.viewportContent()
	contentLines := strings.Count(content, "\n") + 1
	vpHeight := contentLines
	if vpHeight > maxVP {
		vpHeight = maxVP
	}
	if vpHeight < 1 {
		vpHeight = 1
	}

	m.viewport.Height = vpHeight
	m.viewport.SetContent(content)
	if len(m.messages) > 0 {
		m.viewport.GotoBottom()
	}

	var sb strings.Builder
	sb.WriteString(styledCompactHeader(m.width) + "\n")
	sb.WriteString(styledSeparator(m.width) + "\n")
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")
	sb.WriteString(styledSeparator(m.width) + "\n")

	if m.confirm != nil {
		sb.WriteString(m.confirm.View())
		return sb.String()
	}

	if m.loading {
		sb.WriteString(fmt.Sprintf(" %s  \n", m.spinner.View()))
	} else if m.err != "" {
		sb.WriteString(styleError.Render(fmt.Sprintf(" ⚠ %s  (Esc para cerrar)", m.err)) + "\n")
	}

	sb.WriteString(m.input.View())
	sb.WriteString("\n")
	sb.WriteString(styledSeparator(m.width))
	return sb.String()
}

// viewportContent returns the string to render inside the viewport.
func (m *ChatModel) viewportContent() string {
	if len(m.messages) == 0 {
		return m.welcomeView()
	}
	var sb strings.Builder
	for _, msg := range m.messages {
		if msg.role == "user" {
			line := styleUserMsg.Render("> " + msg.content)
			if m.width > 0 {
				pad := m.width - lipgloss.Width(line) - 1
				if pad < 0 {
					pad = 0
				}
				sb.WriteString(strings.Repeat(" ", pad) + line + "\n")
			} else {
				sb.WriteString(line + "\n")
			}
		} else {
			sb.WriteString(renderMarkdown(msg.content, m.width) + "\n")
		}
	}
	return sb.String()
}

// sendChat returns a command that sends the user message to the API.
func (m ChatModel) sendChat(text string) tea.Cmd {
	c := m.client
	convID := m.conversationID
	return func() tea.Msg {
		resp, err := c.Chat(context.Background(), client.ChatRequest{
			ConversationID: convID,
			Message:        text,
		})
		return chatResponseMsg{resp: resp, err: err}
	}
}

// sendConfirm returns a command that confirms a pending tool call.
func (m ChatModel) sendConfirm(pending *client.PendingConfirmation) tea.Cmd {
	c := m.client
	convID := m.conversationID
	return func() tea.Msg {
		resp, err := c.Chat(context.Background(), client.ChatRequest{
			ConversationID: convID,
			Confirm:        true,
			PendingTool:    pending,
		})
		return chatResponseMsg{resp: resp, err: err}
	}
}

// appendMsg adds a message to the conversation history.
func (m *ChatModel) appendMsg(role, content string) {
	m.messages = append(m.messages, chatMessage{role: role, content: content})
}

// renderMarkdown renders markdown text for terminal display.
// Falls back to plain text if glamour fails.
func renderMarkdown(content string, width int) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}
	out, err := r.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimRight(out, "\n")
}

// welcomeView returns the logo shown at the top when the chat has no messages yet.
func (m *ChatModel) welcomeView() string {
	art := lipgloss.NewStyle().Foreground(colorAccent).Render(logo)
	sub := styleHeaderSub.Render("  personal finance · AI-powered")
	return art + "\n" + sub
}
