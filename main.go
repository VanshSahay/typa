package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/harmonica"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

	// Camera (float); springs follow cursor for smooth scrolling.
	camX, camY     float64
	velX, velY     float64
	springX, springY harmonica.Spring

	dots  map[pos]struct{}
	score int
}

const targetDots = 5
const dotValue = 10

func (m *model) Init() tea.Cmd {
	m.cells = make(map[pos]rune)
	m.dots = make(map[pos]struct{})
	m.dx, m.dy = 1, 0
	m.springX = harmonica.NewSpring(harmonica.FPS(60), 9.0, 0.92)
	m.springY = harmonica.NewSpring(harmonica.FPS(60), 9.0, 0.92)
	m.camX, m.camY = float64(m.cx), float64(m.cy)
	return tick60()
}

func isOpposite(ax, ay, bx, by int) bool {
	return ax == -bx && ay == -by
}

// faceDirection sets facing to (dx,dy) unless that is exactly opposite to current facing.
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

func (m *model) spawnCollectible() {
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
		return
	}
}

func (m *model) ensureCollectibles() {
	for m.dots != nil && len(m.dots) < targetDots {
		m.spawnCollectible()
	}
}

func (m *model) tryCollect() {
	p := pos{m.cx, m.cy}
	if _, ok := m.dots[p]; !ok {
		return
	}
	delete(m.dots, p)
	m.score += dotValue
	m.spawnCollectible()
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
	if len(m.stack) == 0 {
		return
	}
	s := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	m.setCell(pos{s.x, s.y}, ' ')
	m.cx, m.cy = s.x, s.y
}

func (m *model) place(r rune) {
	p := pos{m.cx, m.cy}
	delete(m.dots, p)
	m.setCell(p, r)
	m.stack = append(m.stack, step{x: m.cx, y: m.cy, r: r})
	m.tryDirectionWord()
	m.advanceCursor()
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
		m.resize(msg.Width, msg.Height)
		m.ensureCollectibles()
		return m, tick60()

	case tea.KeyPressMsg:
		k := msg.Key()
		if k.Mod&(tea.ModCtrl|tea.ModAlt|tea.ModMeta|tea.ModSuper) != 0 {
			if k.String() == "ctrl+c" {
				return m, tea.Quit
			}
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
		case "esc":
			return m, tea.Quit
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

func (m *model) hintText() string {
	dir := "→"
	switch {
	case m.dx == 1 && m.dy == 0:
		dir = "→"
	case m.dx == -1 && m.dy == 0:
		dir = "←"
	case m.dx == 0 && m.dy == -1:
		dir = "↑"
	case m.dx == 0 && m.dy == 1:
		dir = "↓"
	}
	line1 := fmt.Sprintf("score %d  · typing moves %s  · ◎ collectibles (+ %d) — walk your cursor onto them", m.score, dir, dotValue)
	line2 := "arrows / words turn (no 180°)  ·  type over a ◎ to clear it without score  ·  esc quits"
	return line1 + "\n" + line2
}

func (m *model) viewportSize() (viewCols, viewRows int) {
	contentW := max(1, m.width-4)
	hintBlock := lipgloss.NewStyle().Width(contentW).Render(m.hintText())
	hintLines := strings.Count(hintBlock, "\n") + 1
	viewRows = max(1, m.height-hintLines-3)
	viewCols = contentW
	return viewCols, viewRows
}

func (m *model) renderViewport(viewCols, viewRows int) string {
	ink := lipgloss.NewStyle().Foreground(lipgloss.Color("#e4e4e7")).Bold(true)
	cur := lipgloss.NewStyle().
		Underline(true).
		Foreground(lipgloss.Color("#fafafa"))

	halfW, halfH := viewCols/2, viewRows/2
	centerX := math.Round(m.camX)
	centerY := math.Round(m.camY)

	var b strings.Builder
	for sy := 0; sy < viewRows; sy++ {
		for sx := 0; sx < viewCols; sx++ {
			wx := int(centerX) + (sx - halfW)
			wy := int(centerY) + (sy - halfH)
			p := pos{wx, wy}
			r := m.getCell(p)
			atCursor := wx == m.cx && wy == m.cy
			empty := r == ' ' || r == 0

			switch {
			case atCursor && empty:
				b.WriteString(cur.Render(" "))
			case atCursor && !empty:
				b.WriteString(cur.Render(string(r)))
			case empty:
				b.WriteByte(' ')
			default:
				b.WriteString(ink.Render(string(r)))
			}
		}
		if sy < viewRows-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (m *model) View() tea.View {
	if m.width < 4 || m.height < 4 {
		v := tea.NewView(lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a")).Render("terminal too small"))
		v.AltScreen = true
		return v
	}

	viewCols, viewRows := m.viewportSize()
	grid := m.renderViewport(viewCols, viewRows)

	innerW := max(1, m.width-4)
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#71717a")).
		Width(innerW).
		Render(m.hintText())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#52525b")).
		Padding(0, 1).
		Width(m.width).
		Render(grid + "\n" + hint)

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
