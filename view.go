package main

import (
	"fmt"
	"math"
	"strings"

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

func (m *model) headerLine() string {
	if m.gameOver {
		return ""
	}
	return fmt.Sprintf("score %d", m.score)
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
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#f87171")).
		Render("GAME OVER")
	sub := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e4e4e7")).
		Render(fmt.Sprintf("final score %d", m.score))
	blurb := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a1a1aa")).
		Render("You crossed your own trail.")
	block := lipgloss.JoinVertical(lipgloss.Center, title, sub, "", blurb)
	return lipgloss.Place(
		m.width,
		m.height-3,
		lipgloss.Center,
		lipgloss.Center,
		block,
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
