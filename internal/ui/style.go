package ui

import "github.com/charmbracelet/lipgloss"

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
	ColorNoColor      = lipgloss.AdaptiveColor{Dark: "", Light: ""}
)

var (
	StyleTitle    = lipgloss.NewStyle().Foreground(ColorCream).Background(ColorIndigo)
	StyleSubtitle = lipgloss.NewStyle().Foreground(ColorCream).Background(ColorSubtleIndigo)
	StyleInfo     = lipgloss.NewStyle().Foreground(ColorCream).Background(ColorNoColor)

	StyleQuitMsg  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"})
	StyleErrorMsg = lipgloss.NewStyle().Foreground(ColorRed)
)
