package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Colors for dark and light backgrounds.
var (
	ColorIndigo       = lipgloss.AdaptiveColor{Dark: "#7571F9", Light: "#5A56E0"}
	ColorSubtleIndigo = lipgloss.AdaptiveColor{Dark: "#514DC1", Light: "#7D79F6"}
	ColorCream        = lipgloss.AdaptiveColor{Dark: "#FFFDF5", Light: "#FFFDF5"}
	ColorYellowGreen  = lipgloss.AdaptiveColor{Dark: "#ECFD65", Light: "#04B575"}
	ColorFuschia      = lipgloss.AdaptiveColor{Dark: "#EE6FF8", Light: "#EE6FF8"}
	ColorGreen        = lipgloss.AdaptiveColor{Dark: "#04B575", Light: "#04B575"}
	ColorRed          = lipgloss.AdaptiveColor{Dark: "#ED567A", Light: "#FF4672"}
	ColorFaintRed     = lipgloss.AdaptiveColor{Dark: "#C74665", Light: "#FF6F91"}
	ColorGrey         = lipgloss.AdaptiveColor{Light: "#B2B2B2", Dark: "#4A4A4A"}
	ColorNoColor      = lipgloss.AdaptiveColor{Dark: "", Light: ""}
)

var (
	StyleTitle    = lipgloss.NewStyle().Foreground(ColorCream).Background(ColorIndigo)
	StyleSubtitle = lipgloss.NewStyle().Foreground(ColorCream).Background(ColorSubtleIndigo)
	StyleComment  = lipgloss.NewStyle().Foreground(ColorGrey)

	StyleTableFunc = func() table.Styles {
		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		return s
	}
	StyleTableBase = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	StyleActiveDot   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	StyleInactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")

	StyleQuitMsg  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"})
	StyleErrorMsg = lipgloss.NewStyle().Foreground(ColorRed)
)
