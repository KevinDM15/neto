package ui

import "github.com/charmbracelet/lipgloss"

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

// logo is the ASCII art shown on the login screen.
const logo = ` ███╗   ██╗███████╗████████╗ ██████╗ 
 ████╗  ██║██╔════╝╚══██╔══╝██╔═══██╗
 ██╔██╗ ██║█████╗     ██║   ██║   ██║
 ██║╚██╗██║██╔══╝     ██║   ██║   ██║
 ██║ ╚████║███████╗   ██║   ╚██████╔╝
 ╚═╝  ╚═══╝╚══════╝   ╚═╝    ╚═════╝`

// styledLogo returns the logo with accent color + subtitle.
func styledLogo() string {
	art := lipgloss.NewStyle().Foreground(colorAccent).Render(logo)
	sub := styleHeaderSub.Render(" personal finance · AI-powered")
	return art + "\n" + sub
}

// styledCompactHeader returns the one-line header for the chat screen.
func styledCompactHeader(width int) string {
	left := styleHeader.Render("◆ Neto") + "  " + styleHeaderSub.Render("personal finance · AI-powered")
	hint := styleHint.Render("Ctrl+H help  Ctrl+Q quit")
	gap := width - lipgloss.Width(left) - lipgloss.Width(hint)
	if gap < 1 {
		gap = 1
	}
	return left + lipgloss.NewStyle().Width(gap).Render("") + hint
}

// styledSeparator returns a full-width horizontal rule.
func styledSeparator(width int) string {
	return styleSeparator.Render(repeatRune('─', width))
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
