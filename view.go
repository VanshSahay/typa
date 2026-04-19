package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func repeatStyle(st lipgloss.Style, r rune, n int) string {
	if n < 1 {
		return ""
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = r
	}
	return st.Render(string(b))
}

func (m *model) bottomHint() string {
	switch m.phase {
	case phaseTitle:
		return "↑/↓  navigate   ·   Enter   ·   Esc quit"
	case phaseName:
		return "Enter confirm   ·   Esc back"
	case phaseEnd:
		return "m menu   ·   Esc quit"
	default:
		return "Esc quit"
	}
}

func (m *model) viewportSize() (logicalCols, logicalRows int) {
	contentW := max(1, m.width-4)
	header := lipgloss.NewStyle().Width(contentW).Render(m.headerLine())
	headerLines := strings.Count(header, "\n") + 1
	hint := lipgloss.NewStyle().Width(contentW).Render(m.bottomHint())
	hintLines := strings.Count(hint, "\n") + 1
	border := 2
	avail := m.height - border - headerLines - hintLines - 1
	if avail < 1 {
		avail = 1
	}
	logicalRows = max(1, avail)
	logicalCols = max(1, contentW/scaleX)
	return logicalCols, logicalRows
}

func formatCountdown(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Second)
	sec := int(d / time.Second)
	mm := sec / 60
	ss := sec % 60
	return fmt.Sprintf("%02d:%02d", mm, ss)
}

func (m *model) headerLine() string {
	if m.phase != phasePlay {
		return ""
	}
	wpm := m.grossWPM()
	return fmt.Sprintf(
		"%s  ·  score %d  ·  wpm %.0f  ·  %s left",
		m.username,
		m.score,
		math.Round(wpm),
		formatCountdown(m.remainingRound()),
	)
}

func (m *model) renderViewport(logicalCols, logicalRows int) string {
	ink := lipgloss.NewStyle().Foreground(lipgloss.Color("#e4e4e7")).Bold(true)
	coin := lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Bold(true)
	cur := lipgloss.NewStyle().
		Underline(true).
		Foreground(lipgloss.Color("#fafafa"))
	curCoin := lipgloss.NewStyle().
		Underline(true).
		Foreground(lipgloss.Color("#fbbf24")).
		Bold(true)

	halfW, halfH := logicalCols/2, logicalRows/2
	centerX := math.Round(m.camX)
	centerY := math.Round(m.camY)

	var b strings.Builder
	for sy := 0; sy < logicalRows; sy++ {
		var row strings.Builder
		for sx := 0; sx < logicalCols; sx++ {
			wx := int(centerX) + (sx - halfW)
			wy := int(centerY) + (sy - halfH)
			p := pos{wx, wy}
			r := m.getCell(p)
			atCursor := wx == m.cx && wy == m.cy
			empty := r == ' ' || r == 0
			_, hasDot := m.dots[p]

			var seg string
			switch {
			case atCursor && empty && hasDot:
				seg = repeatStyle(curCoin, '◎', scaleX)
			case atCursor && empty && !hasDot:
				seg = repeatStyle(cur, ' ', scaleX)
			case atCursor && !empty:
				seg = repeatStyle(cur, r, scaleX)
			case empty && hasDot:
				seg = repeatStyle(coin, '◎', scaleX)
			case empty:
				seg = strings.Repeat(" ", scaleX)
			default:
				seg = repeatStyle(ink, r, scaleX)
			}
			row.WriteString(seg)
		}
		b.WriteString(row.String())
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func (m *model) viewTitle() tea.View {
	gold := lipgloss.NewStyle().Foreground(lipgloss.Color("#eab308")).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a"))
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#a3e635"))

	logo := gold.Render(`
████████╗██╗   ██╗██████╗  █████╗ 
╚══██╔══╝╚██╗ ██╔╝██╔══██╗██╔══██╗
   ██║    ╚████╔╝ ██████╔╝███████║
   ██║     ╚██╔╝  ██╔═══╝ ██╔══██║
   ██║      ██║   ██║     ██║  ██║
   ╚═╝      ╚═╝   ╚═╝     ╚═╝  ╚═╝
                                  `)

	tag := muted.Render("ssh / terminal typing chase · collect ◎ · survive your trail")

	start := "Start run"
	exit := "Exit"
	if m.titleSel == 0 {
		start = lipgloss.NewStyle().Reverse(true).Foreground(lipgloss.Color("#fafafa")).Render("  Start run")
		exit = lipgloss.NewStyle().Foreground(lipgloss.Color("#d4d4d8")).Render("  Exit")
	} else {
		start = lipgloss.NewStyle().Foreground(lipgloss.Color("#d4d4d8")).Render("  Start run")
		exit = lipgloss.NewStyle().Reverse(true).Foreground(lipgloss.Color("#fafafa")).Render("  Exit")
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, start, exit)
	body := lipgloss.JoinVertical(
		lipgloss.Center,
		strings.TrimSpace(logo),
		"",
		accent.Render("TYPA"),
		"",
		tag,
		"",
		"",
		menu,
	)

	// Shrink-wrap width + horizontal align so the bordered panel centers in the terminal (full Width(m.width) left-aligned text).
	innerW := max(1, min(m.width-4, 88))

	bodyBlock := lipgloss.NewStyle().
		Width(innerW).
		Align(lipgloss.Center).
		Render(body)

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#52525b")).
		Width(innerW).
		Align(lipgloss.Center).
		Render(m.bottomHint())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#52525b")).
		Padding(1, 2).
		Render(bodyBlock + "\n\n" + hint)

	screen := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	v := tea.NewView(screen)
	v.AltScreen = true
	return v
}

func (m *model) viewName() tea.View {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#e4e4e7")).Render("Pilot name")
	prompt := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#86efac")).Render("> "),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#fafafa")).Render(m.nameBuf),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a")).Blink(true).Render("▏"),
	)
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a")).Render("This name appears on the public leaderboard.")

	body := lipgloss.JoinVertical(lipgloss.Left, title, "", prompt, "", sub)

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#52525b")).
		Width(m.width - 4).
		Render(m.bottomHint())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#52525b")).
		Padding(1, 2).
		Width(m.width).
		Render(body + "\n\n" + hint)

	screen := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	v := tea.NewView(screen)
	v.AltScreen = true
	return v
}

func (m *model) gameOverView() string {
	var headline, accent lipgloss.Style
	var tagline, banner string

	switch m.endReason {
	case endTimeout:
		headline = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#fbbf24"))
		accent = lipgloss.NewStyle().Foreground(lipgloss.Color("#fde68a")).Bold(true)
		tagline = "Two minutes are up — run complete."
		banner = "TIME'S UP"
	default:
		headline = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f87171"))
		accent = lipgloss.NewStyle().Foreground(lipgloss.Color("#fca5a5")).Bold(true)
		tagline = "You crossed your own trail."
		banner = "SELF-COLLISION"
	}

	titleBlock := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#a3e635")).Bold(true).Render("TYPA"),
		"",
		headline.Render("RUN ENDED"),
		"",
		accent.Render(banner),
	)

	stats := lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		fmt.Sprintf("  Pilot    %s", m.username),
		fmt.Sprintf("  Score    %d", m.score),
		fmt.Sprintf("  WPM      %.0f", math.Round(m.grossWPM())),
		"",
	)
	statsStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e4e4e7")).
		Width(min(m.width-8, 46)).
		Render(stats)

	blurb := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a1a1aa")).
		Width(min(m.width-8, 56)).
		Align(lipgloss.Center).
		Render(tagline)

	block := lipgloss.JoinVertical(
		lipgloss.Center,
		titleBlock,
		"",
		"",
		statsStyled,
		"",
		"",
		blurb,
	)

	pad := max(4, (m.height-22)/2)
	return lipgloss.Place(
		m.width,
		m.height-4,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.NewStyle().Padding(pad, 3, pad, 3).Render(block),
	)
}

func (m *model) viewPlay() tea.View {
	logicalCols, logicalRows := m.viewportSize()
	grid := m.renderViewport(logicalCols, logicalRows)

	innerW := max(1, m.width-4)
	head := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#a1a1aa")).
		Width(innerW).
		Render(m.headerLine())

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#71717a")).
		Width(innerW).
		Align(lipgloss.Center).
		Render(m.bottomHint())

	body := head + "\n" + grid

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#52525b")).
		Padding(0, 1).
		Width(m.width).
		Render(body + "\n\n" + hint)

	screen := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)

	v := tea.NewView(screen)
	v.AltScreen = true
	return v
}

func (m *model) viewGameOver() tea.View {
	innerW := max(1, m.width-4)
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#71717a")).
		Width(innerW).
		Align(lipgloss.Center).
		Render(m.bottomHint())
	main := m.gameOverView()
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#52525b")).
		Padding(0, 1).
		Width(m.width).
		Render(main + "\n\n" + hint)
	screen := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	v := tea.NewView(screen)
	v.AltScreen = true
	return v
}

func (m *model) View() tea.View {
	if m.width < 4 || m.height < 4 {
		v := tea.NewView(lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a")).Render("terminal too small"))
		v.AltScreen = true
		return v
	}

	switch m.phase {
	case phaseTitle:
		return m.viewTitle()
	case phaseName:
		return m.viewName()
	case phaseEnd:
		return m.viewGameOver()
	default:
		return m.viewPlay()
	}
}
