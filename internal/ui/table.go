package ui

import "github.com/charmbracelet/bubbles/table"

func TableColumn(width int) []table.Column {
	const statusWidth = 6
	const actionWidth = 8
	const timeWidth = 24

	dynamicWidth := width - statusWidth - actionWidth - timeWidth

	indexWidth := dynamicWidth / 5
	moduleWidth := dynamicWidth / 5 * 2
	resourceWidth := dynamicWidth / 5 * 2

	return []table.Column{
		{Title: "Index", Width: indexWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Action", Width: actionWidth},
		{Title: "Module", Width: moduleWidth},
		{Title: "Resource", Width: resourceWidth},
		{Title: "Time", Width: timeWidth},
	}
}
