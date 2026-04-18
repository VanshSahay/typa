package main

import (
	"math/rand"
	"strings"
	"unicode"

	"github.com/charmbracelet/harmonica"
	tea "charm.land/bubbletea/v2"
)

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

// --- tea.Model input loop

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
		return m.handleKey(msg)
	}

	return m, tick60()
}

func (m *model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.Key()
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
