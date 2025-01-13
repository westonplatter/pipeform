package state

import (
	gojson "encoding/json"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/magodo/pipeform/internal/terraform/views/json"
)

type OutputInfo struct {
	Name string

	Sensitive bool
	Type      string
	ValueStr  gojson.RawMessage
	Action    json.ChangeAction
}

type OutputInfos []*OutputInfo

func (infos OutputInfos) ToRows() []table.Row {
	var rows []table.Row
	for i, info := range infos {
		row := []string{
			strconv.Itoa(i + 1),
			info.Name,
			info.Type,
			fmt.Sprintf("%t", info.Sensitive),
			string(info.ValueStr),
		}
		rows = append(rows, row)
	}
	return rows
}

func (infos OutputInfos) ToColumns(width int) []table.Column {
	const indexWidth = 6
	const typeWidth = 8
	const sensitiveWidth = 10

	dynamicWidth := width - indexWidth - typeWidth - sensitiveWidth

	nameWidth := dynamicWidth / 2
	valueWidth := dynamicWidth / 2

	return []table.Column{
		{Title: "Index", Width: indexWidth},
		{Title: "Name", Width: nameWidth},
		{Title: "Type", Width: typeWidth},
		{Title: "Sensitive", Width: sensitiveWidth},
		{Title: "Value", Width: valueWidth},
	}
}
