package main

import "github.com/charmbracelet/harmonica"

// frameMsg is dispatched on each animation tick (60 fps).
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

// scaleX is terminal columns per logical map cell (horizontal size).
const scaleX = 1
