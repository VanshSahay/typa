package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/harmonica"
)

type frameMsg struct{}

type pos struct {
	x, y int
}

type step struct {
	x, y int
	r    rune
}

type model struct {
	width  int
	height int

	cells map[pos]rune
	cx    int
	cy    int
	dx    int
	dy    int

	stack []step

	camX, camY       float64
	velX, velY       float64
	springX, springY harmonica.Spring

	dots     map[pos]struct{}
	score    int
	gameOver bool
}

const targetDots = 5
const dotValue = 10

// Visual scale: each logical cell is drawn scaleX terminal columns wide (taller letters).
// Do not duplicate whole rows vertically — that looked like two parallel text streams.
const scaleX = 1

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

func (m *model) Init() tea.Cmd {
	m.cells = make(map[pos]rune)
	m.dots = make(map[pos]struct{})
	m.dx, m.dy = 1, 0
	m.springX = harmonica.NewSpring(harmonica.FPS(60), 9.0, 0.92)
	m.springY = harmonica.NewSpring(harmonica.FPS(60), 9.0, 0.92)
	m.camX, m.camY = float64(m.cx), float64(m.cy)
	m.ensureCollectibles()
	return tick60()
}

func isOpposite(ax, ay, bx, by int) bool {
	return ax == -bx && ay == -by
}

func (m *model) faceDirection(dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}
	if m.dx == dx && m.dy == dy {
		return
	}
	if isOpposite(m.dx, m.dy, dx, dy) {
		return
	}
	m.dx, m.dy = dx, dy
}

func (m *model) spawnCollectible() bool {
	for range 120 {
		dx := rand.Intn(55) - 27
		dy := rand.Intn(55) - 27
		if dx*dx+dy*dy < 8*8 {
			continue
		}
		p := pos{m.cx + dx, m.cy + dy}
		if m.getCell(p) != ' ' {
			continue
		}
		if _, taken := m.dots[p]; taken {
			continue
		}
		m.dots[p] = struct{}{}
		return true
	}
	return false
}

func (m *model) ensureCollectibles() {
	if m.gameOver {
		return
	}
	for m.dots != nil && len(m.dots) < targetDots {
		if !m.spawnCollectible() {
			break
		}
	}
}

func (m *model) tryCollect() {
	p := pos{m.cx, m.cy}
	if _, ok := m.dots[p]; !ok {
		return
	}
	delete(m.dots, p)
	m.score += dotValue
	m.ensureCollectibles()
}

func tick60() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return frameMsg{}
	})
}

func (m *model) resize(w, h int) {
	m.width, m.height = w, h
}

func (m *model) setCell(p pos, r rune) {
	if r == ' ' {
		delete(m.cells, p)
	} else {
		m.cells[p] = r
	}
}

func (m *model) getCell(p pos) rune {
	if r, ok := m.cells[p]; ok {
		return r
	}
	return ' '
}

func (m *model) advanceCursor() {
	m.cx += m.dx
	m.cy += m.dy
}

var directionWords = []struct {
	word string
	dx   int
	dy   int
}{
	{"right", 1, 0},
	{"down", 0, 1},
	{"left", -1, 0},
	{"up", 0, -1},
}

func (m *model) tryDirectionWord() {
	for _, dw := range directionWords {
		n := len(dw.word)
		if len(m.stack) < n {
			continue
		}
		var b strings.Builder
		start := len(m.stack) - n
		for i := start; i < len(m.stack); i++ {
			b.WriteRune(m.stack[i].r)
		}
		if strings.EqualFold(b.String(), dw.word) {
			m.faceDirection(dw.dx, dw.dy)
			return
		}
	}
}

func (m *model) undoOne() {
	if m.gameOver {
		return
	}
	if len(m.stack) == 0 {
		return
	}
	s := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	m.setCell(pos{s.x, s.y}, ' ')
	m.cx, m.cy = s.x, s.y
}

func (m *model) place(r rune) {
	if m.gameOver {
		return
	}
	p := pos{m.cx, m.cy}
	delete(m.dots, p)
	m.setCell(p, r)
	m.stack = append(m.stack, step{x: m.cx, y: m.cy, r: r})
	m.tryDirectionWord()
	m.advanceCursor()

	next := pos{m.cx, m.cy}
	if m.getCell(next) != ' ' {
		m.gameOver = true
		return
	}
	m.tryCollect()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case frameMsg:
		m.camX, m.velX = m.springX.Update(m.camX, m.velX, float64(m.cx))
		m.camY, m.velY = m.springY.Update(m.camY, m.velY, float64(m.cy))
		return m, tick60()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.gameOver {
			m.resize(msg.Width, msg.Height)
			m.ensureCollectibles()
		} else {
			m.resize(msg.Width, msg.Height)
		}
		return m, tick60()

	case tea.KeyPressMsg:
		k := msg.Key()
		// Key repeat would place the same glyph twice in one tick; ignore for typed text.
		if k.IsRepeat && k.Text != "" {
			return m, tick60()
		}
		if k.Mod&(tea.ModCtrl|tea.ModAlt|tea.ModMeta|tea.ModSuper) != 0 {
			if k.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, tick60()
		}
		switch k.String() {
		case "esc":
			return m, tea.Quit
		}

		if m.gameOver {
			return m, tick60()
		}

		switch k.Code {
		case tea.KeyUp:
			m.faceDirection(0, -1)
			return m, tick60()
		case tea.KeyDown:
			m.faceDirection(0, 1)
			return m, tick60()
		case tea.KeyLeft:
			m.faceDirection(-1, 0)
			return m, tick60()
		case tea.KeyRight:
			m.faceDirection(1, 0)
			return m, tick60()
		}

		switch k.String() {
		case "backspace", "ctrl+h":
			m.undoOne()
			return m, tick60()
		case "space":
			if m.cells != nil {
				m.place(' ')
			}
			return m, tick60()
		default:
			if k.Text != "" && m.cells != nil {
				for _, r := range k.Text {
					if unicode.IsPrint(r) {
						m.place(r)
					}
				}
			}
			return m, tick60()
		}
	}

	return m, tick60()
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

func main() {
	p := tea.NewProgram(&model{})

	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
