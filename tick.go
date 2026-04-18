package main

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

func tick60() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return frameMsg{}
	})
}
