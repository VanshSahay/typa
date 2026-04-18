package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type step struct {
	x, y int
	r    rune
}

type model struct {
	width  int
	height int
	cols   int
	rows   int

	grid [][]rune
	cx   int
	cy   int
	dx   int
	dy   int

	stack []step
}

func (m *model) Init() tea.Cmd {
	m.dx, m.dy = 1, 0
	return nil
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (m *model) resize(w, h int) {
	m.width, m.height = w, h
	cols := max(4, w-4)
	rows := max(4, h-6)
	if cols == m.cols && rows == m.rows && m.grid != nil {
		return
	}

	next := make([][]rune, rows)
	for y := range next {
		next[y] = make([]rune, cols)
		for x := range next[y] {
			ch := ' '
			if m.grid != nil && y < len(m.grid) && m.grid[y] != nil && x < len(m.grid[y]) {
				ch = m.grid[y][x]
			}
			next[y][x] = ch
		}
	}
	m.grid = next
	m.cols, m.rows = cols, rows
	m.cx = clamp(m.cx, 0, m.cols-1)
	m.cy = clamp(m.cy, 0, m.rows-1)
}

func (m *model) advanceCursor() {
	m.cx = (m.cx + m.dx + m.cols) % m.cols
	m.cy = (m.cy + m.dy + m.rows) % m.rows
}

func (m *model) setDirection(dx, dy int) {
	m.dx, m.dy = dx, dy
}

// directionWords: longest first so we match "right" before any shared prefix issue.
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
			m.undoSteps(n)
			m.setDirection(dw.dx, dw.dy)
			return
		}
	}
}

func (m *model) undoSteps(n int) {
	if len(m.stack) < n || n <= 0 {
		return
	}
	start := len(m.stack) - n
	first := m.stack[start]
	for i := len(m.stack) - 1; i >= start; i-- {
		s := m.stack[i]
		m.grid[s.y][s.x] = ' '
	}
	m.stack = m.stack[:start]
	m.cx, m.cy = first.x, first.y
}

func (m *model) undoOne() {
	if len(m.stack) == 0 {
		return
	}
	s := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	m.grid[s.y][s.x] = ' '
	m.cx, m.cy = s.x, s.y
}

func (m *model) place(r rune) {
	m.grid[m.cy][m.cx] = r
	m.stack = append(m.stack, step{x: m.cx, y: m.cy, r: r})
	m.tryDirectionWord()
	m.advanceCursor()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		k := msg.Key()
		if k.Mod&(tea.ModCtrl|tea.ModAlt|tea.ModMeta|tea.ModSuper) != 0 {
			if k.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}

		switch k.Code {
		case tea.KeyUp:
			m.setDirection(0, -1)
			return m, nil
		case tea.KeyDown:
			m.setDirection(0, 1)
			return m, nil
		case tea.KeyLeft:
			m.setDirection(-1, 0)
			return m, nil
		case tea.KeyRight:
			m.setDirection(1, 0)
			return m, nil
		}

		switch k.String() {
		case "esc":
			return m, tea.Quit
		case "backspace", "ctrl+h":
			m.undoOne()
			return m, nil
		case "space":
			if m.grid != nil {
				m.place(' ')
			}
			return m, nil
		default:
			if k.Text != "" && m.grid != nil {
				for _, r := range k.Text {
					if unicode.IsPrint(r) {
						m.place(r)
					}
				}
			}
		}
	}

	return m, nil
}

func (m *model) renderGrid() string {
	ink := lipgloss.NewStyle().Foreground(lipgloss.Color("#e4e4e7"))
	floor := lipgloss.NewStyle().Foreground(lipgloss.Color("#3f3f46"))
	cur := lipgloss.NewStyle().
		Background(lipgloss.Color("#6366f1")).
		Foreground(lipgloss.Color("#fafafa"))

	var b strings.Builder
	for y := 0; y < m.rows; y++ {
		for x := 0; x < m.cols; x++ {
			r := m.grid[y][x]
			atCursor := x == m.cx && y == m.cy
			ch := string(r)
			if r == ' ' {
				ch = " "
				if atCursor {
					b.WriteString(cur.Render(" "))
					continue
				}
				b.WriteString(floor.Render("·"))
				continue
			}
			if atCursor {
				b.WriteString(cur.Render(ch))
			} else {
				b.WriteString(ink.Render(ch))
			}
		}
		if y < m.rows-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (m *model) hintBar() string {
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
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a"))
	return s.Render("typing moves "+dir+"  ·  arrows aim  ·  type up down left right to turn  ·  esc quits")
}

func (m *model) View() tea.View {
	w := m.width
	h := m.height
	if w < 10 {
		w = 80
	}
	if h < 6 {
		h = 24
	}

	innerW := w - 4
	if m.grid == nil {
		return tea.NewView(lipgloss.NewStyle().Foreground(lipgloss.Color("#71717a")).Render("resize the terminal…"))
	}

	grid := m.renderGrid()
	grid = lipgloss.NewStyle().Width(innerW).Render(grid)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(w - 2).
		Render(grid + "\n" + m.hintBar())

	v := tea.NewView(
		lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Top,
			lipgloss.NewStyle().MarginTop(1).Render(box),
		),
	)
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
