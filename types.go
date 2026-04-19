package main

import (
	"time"

	"github.com/charmbracelet/harmonica"
)

// frameMsg is dispatched on each animation tick (60 fps).
type frameMsg struct{}

type pos struct {
	x, y int
}

type step struct {
	x, y int
	r    rune
}

type sessionPhase uint8

const (
	phaseTitle sessionPhase = iota
	phaseName
	phasePlay
	phaseEnd
)

type model struct {
	width  int
	height int

	phase    sessionPhase
	titleSel int // 0 = start, 1 = exit

	username string
	nameBuf  string

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

	strokes     int
	typingStart time.Time
	runEnd      time.Time

	roundStart      time.Time
	endReason       endReason
	scoreSubmitted  bool // POST once per run to API
}

type endReason uint8

const (
	endNone endReason = iota
	endCollision
	endTimeout
)

const targetDots = 5
const dotValue = 10

// RoundDuration is the wall-clock limit for a run (from round start in phasePlay).
const RoundDuration = 2 * time.Minute

// scaleX is terminal columns per logical map cell (horizontal size).
const scaleX = 1
