package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/neto-app/neto/tui/internal/client"
	"github.com/neto-app/neto/tui/internal/config"
)

const helpText = `
Available commands (examples):
  Gasté 50k en luz          — record an expense
  Cuánto gasté este mes?    — query spending
  Mostrar cuentas           — list accounts
  Cuánto debo?              — check debts
  Meta de ahorro 100k       — set a savings goal

Ctrl+H  help  •  Ctrl+Q / Ctrl+C  quit
`

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
	input          textinput.Model
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
	ti := textinput.New()
	ti.Placeholder = "Escribe un mensaje…"
	ti.Focus()

	vp := viewport.New(80, 20)
	vp.SetContent("")

	return ChatModel{
		client:   c,
		cfg:      cfg,
		input:    ti,
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
	m.refreshViewport()
	return textinput.Blink
}

// Update implements tea.Model.
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 8 // header + 2 separators + input + bottom sep + status
		m.refreshViewport()
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
				m.refreshViewport()
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
		case "enter":
			if m.loading || m.input.Value() == "" {
				return m, nil
			}
			text := m.input.Value()
			m.input.Reset()
			m.appendMsg("user", text)
			m.refreshViewport()
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
			// Show confirmation overlay.
			cm := NewConfirmModel(msg.resp.PendingConfirmation)
			m.confirm = &cm
			m.appendMsg("assistant", msg.resp.Reply)
			m.refreshViewport()
			return m, cm.Init()
		}
		m.appendMsg("assistant", msg.resp.Reply)
		m.refreshViewport()
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

	// Input.
	var tiCmd tea.Cmd
	if !m.loading {
		m.input, tiCmd = m.input.Update(msg)
	}

	return m, tea.Batch(vpCmd, tiCmd)
}

// View implements tea.Model.
func (m ChatModel) View() string {
	if m.showHelp {
		return helpText
	}

	// chrome = header + sep + sep + input line + bottom sep
	const chrome = 5
	maxVP := m.height - chrome
	if maxVP < 1 {
		maxVP = 1
	}

	// Count content lines to size viewport dynamically.
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
		sb.WriteString(fmt.Sprintf(" %s  ", m.spinner.View()))
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

// refreshViewport is a no-op; View() recalculates content on every render.
func (m *ChatModel) refreshViewport() {}

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
	sep := styledSeparator(m.width)
	return art + "\n" + sub + "\n" + sep
}
