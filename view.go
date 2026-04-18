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
	return "Press Esc to exit."
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
	if m.gameOver {
		return ""
	}
	wpm := m.grossWPM()
	return fmt.Sprintf(
		"score %d  ·  wpm %.0f  ·  %s left",
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

func (m *model) gameOverView() string {
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a"))

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
		headline.Render("RUN ENDED"),
		"",
		accent.Render(banner),
	)

	stats := lipgloss.JoinVertical(
		lipgloss.Left,
		muted.Render("RESULTS"),
		"",
		fmt.Sprintf("  Score       %d", m.score),
		fmt.Sprintf("  WPM         %.0f", math.Round(m.grossWPM())),
		"",
		muted.Render("  Round limit  02:00"),
	)
	statsStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e4e4e7")).
		Width(min(m.width-8, 42)).
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

	pad := max(6, (m.height-18)/2)
	return lipgloss.Place(
		m.width,
		m.height-3,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.NewStyle().Padding(pad, 2, pad, 2).Render(block),
	)
}

func (m *model) View() tea.View {
	if m.width < 4 || m.height < 4 {
		v := tea.NewView(lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a")).Render("terminal too small"))
		v.AltScreen = true
		return v
	}

	if m.gameOver {
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

	body := grid
	if strings.TrimSpace(m.headerLine()) != "" {
		body = head + "\n" + grid
	}

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
