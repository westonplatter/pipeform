package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
