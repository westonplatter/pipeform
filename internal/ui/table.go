package ui

import "github.com/charmbracelet/bubbles/table"

func TableColumn(width int) []table.Column {
	const statusWidth = 6
	const actionWidth = 8

	indexWidth := (width - statusWidth - actionWidth) / 5
	moduleWidth := (width - statusWidth - actionWidth) / 5 * 2
	resourceWidth := (width - statusWidth - actionWidth) / 5 * 2

	return []table.Column{
		{Title: "Index", Width: indexWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Action", Width: actionWidth},
		{Title: "Module", Width: moduleWidth},
		{Title: "Resource", Width: resourceWidth},
	}
}
