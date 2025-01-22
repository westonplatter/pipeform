package state

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/magodo/pipeform/internal/terraform/views/json"
)

type PlanInfo struct {
	Resource json.ResourceAddr
	Action   json.ChangeAction

	PrevResource *json.ResourceAddr
	Reason       json.ChangeReason
}

type PlanInfos []*PlanInfo

func (infos PlanInfos) ToRows() []table.Row {
	var rows []table.Row
	for i, info := range infos {
		var comment string
		switch info.Action {
		case json.ActionDelete, json.ActionReplace:
			comment = string(info.Reason)
		case json.ActionMove:
			if info.PrevResource != nil {
				source := info.PrevResource.Addr
				if info.PrevResource.Module != "" {
					source = fmt.Sprintf("%s (%s)", source, info.PrevResource.Module)
				}
				comment = fmt.Sprintf("Moved from %s", source)
			}
		}
		row := []string{
			strconv.Itoa(i + 1),
			info.Resource.Module,
			info.Resource.Addr,
			string(info.Action),
			comment,
		}
		rows = append(rows, row)
	}
	return rows
}

func (infos PlanInfos) ToColumns(width int) []table.Column {
	const indexWidth = 6
	const actionWidth = 8

	dynamicWidth := width - indexWidth - actionWidth

	commentWidth := dynamicWidth / 3
	moduleWidth := dynamicWidth / 3
	resourceWidth := dynamicWidth / 3

	return []table.Column{
		{Title: "Index", Width: indexWidth},
		{Title: "Module", Width: moduleWidth},
		{Title: "Resource", Width: resourceWidth},
		{Title: "Action", Width: actionWidth},
		// Comment is a combination of "reason" (for delete/replace) and a modified version of "previous_resource" (for move)
		{Title: "Comment", Width: commentWidth},
	}
}
