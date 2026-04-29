package ui

import (
	"net/url"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	// paddingX is the horizontal padding applied to the entire app view.
	paddingX = 2
	// appVersion is the current TUI version shown in the header and status bar.
	appVersion = "1.0.0"
)

var (
	// colorAccent is the brand green used for interactive elements.
	colorAccent = lipgloss.Color("#00D9A3")
	// colorMuted is for secondary text.
	colorMuted = lipgloss.Color("#6C7A89")
	// colorError is for error messages.
	colorError = lipgloss.Color("#FF5F5F")
	// colorUser is for user messages.
	colorUser = lipgloss.Color("#7EB8F7")

	styleHeader = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleHeaderSub = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleSeparator = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleUserMsg = lipgloss.NewStyle().
			Foreground(colorUser)

	styleError = lipgloss.NewStyle().
			Foreground(colorError)

	styleLabel = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleHint = lipgloss.NewStyle().
			Foreground(colorMuted).
			Faint(true)
)

// logo is the ASCII art shown on the login screen and welcome box.
const logo = ` ███╗   ██╗███████╗████████╗ ██████╗ 
 ████╗  ██║██╔════╝╚══██╔══╝██╔═══██╗
 ██╔██╗ ██║█████╗     ██║   ██║   ██║
 ██║╚██╗██║██╔══╝     ██║   ██║   ██║
 ██║ ╚████║███████╗   ██║   ╚██████╔╝
 ╚═╝  ╚═══╝╚══════╝   ╚═╝    ╚═════╝`

// styledLogo returns the logo with accent color + subtitle (used on the login screen).
func styledLogo() string {
	art := lipgloss.NewStyle().Foreground(colorAccent).Render(logo)
	sub := styleHeaderSub.Render(" personal finance · AI-powered")
	return art + "\n" + sub
}

// styledHeader returns the full-width header line with the app name embedded in the rule.
// Example: ── Neto v1.0.0 ────────────────────────────────────────────────────
func styledHeader(width int) string {
	prefix := styleSeparator.Render("── ")
	name := styleHeader.Render("Neto v" + appVersion)
	space := styleSeparator.Render(" ")
	usedW := lipgloss.Width(prefix) + lipgloss.Width(name) + lipgloss.Width(space)
	fill := repeatRune('─', width-usedW)
	return prefix + name + space + styleSeparator.Render(fill)
}

// styledSeparator returns a full-width horizontal rule.
func styledSeparator(width int) string {
	return styleSeparator.Render(repeatRune('─', width))
}

// styledStatusBar renders a full-width bar with left and right content.
func styledStatusBar(width int, left, right string) string {
	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	gap := width - lw - rw
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// styledWelcomeBox renders a two-column welcome panel styled like Claude Code's welcome screen.
// Left column: logo + info. Right column: example commands + shortcuts.
func styledWelcomeBox(width int, apiURL string) string {
	// Compute left column width from the actual rendered logo lines.
	leftColW := 0
	for _, l := range strings.Split(logo, "\n") {
		w := lipgloss.Width(l)
		if w > leftColW {
			leftColW = w
		}
	}
	leftColW += 2 // small breathing room

	// Box inner width: outer width − border(2) − horizontal padding(2×1=2).
	boxInnerW := width - 4
	if boxInnerW < 40 {
		boxInnerW = 40
	}

	// " │ " divider = 3 chars.
	rightColW := boxInnerW - leftColW - 3

	// Collapse to single column when the terminal is too narrow.
	singleCol := rightColW < 24
	if singleCol {
		rightColW = boxInnerW
	}

	// --- Right column (tips + shortcuts) ---
	titleR1 := lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Render("Comandos de ejemplo")
	cmds := styleHeaderSub.Render(
		"  Gasté 50k en luz    — gasto\n" +
			"  ¿Cuánto gasté?      — consultar\n" +
			"  Mostrar cuentas     — listar\n" +
			"  ¿Cuánto debo?       — deudas\n" +
			"  Meta 100k           — objetivo",
	)
	titleR2 := lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Render("Atajos")
	keys := styleHeaderSub.Render(
		"  Enter        envía\n" +
			"  Alt+Enter    nueva línea\n" +
			"  Ctrl+H       ayuda\n" +
			"  Ctrl+Q       salir",
	)
	rightContent := titleR1 + "\n\n" + cmds + "\n\n" + titleR2 + "\n\n" + keys
	rightCol := lipgloss.NewStyle().Width(rightColW).Render(rightContent)

	var inner string
	if singleCol {
		inner = rightCol
	} else {
		// --- Left column (logo + info) ---
		titleL := lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Render("¡Bienvenido!")
		logoRendered := lipgloss.NewStyle().Foreground(colorAccent).Render(logo)
		sub := styleHeaderSub.Render("personal finance · AI-powered")
		host := styleHeaderSub.Render(shortHost(apiURL))
		leftContent := titleL + "\n\n" + logoRendered + "\n\n" + sub + "\n" + host
		leftCol := lipgloss.NewStyle().Width(leftColW).Render(leftContent)

		// Vertical divider: one "│" per row of the tallest column.
		lh := lipgloss.Height(leftCol)
		rh := lipgloss.Height(rightCol)
		maxH := lh
		if rh > maxH {
			maxH = rh
		}
		divLines := make([]string, maxH)
		for i := range divLines {
			divLines[i] = " │ "
		}
		divCol := styleSeparator.Render(strings.Join(divLines, "\n"))

		inner = lipgloss.JoinHorizontal(lipgloss.Top, leftCol, divCol, rightCol)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(0, 1).
		Render(inner)
}

// shortHost extracts the hostname from a URL for compact display.
func shortHost(rawURL string) string {
	if rawURL == "" {
		return "localhost"
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Host
}

func repeatRune(r rune, n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = r
	}
	return string(b)
}
