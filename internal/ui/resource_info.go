package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
)

type ResourceStatus string

const (
	// Once received one OperationStart hook message
	ResourceStatusStart ResourceStatus = "start"
	// Once received one OperationComplete hook message
	ResourceStatusComplete ResourceStatus = "complete"
	// Once received one OperationErrored hook message
	ResourceStatusErrored ResourceStatus = "error"

	// TODO: Support refresh? (refresh is an independent lifecycle than the resource apply lifecycle)
	// TODO: Support provision? (provision is a intermidiate stage in the resource apply lifecycle)
)

func ResourceStatusEmoji(status ResourceStatus) string {
	switch status {
	case ResourceStatusStart:
		return "🕛"
	case ResourceStatusComplete:
		return "✅"
	case ResourceStatusErrored:
		return "❌"
	default:
		return "❓"
	}
}

type ResourceInfoLocator struct {
	Module       string
	ResourceAddr string
	Action       string
}

type ResourceInfo struct {
	Loc       ResourceInfoLocator
	Status    ResourceStatus
	StartTime time.Time
	EndTime   time.Time
}

type ResourceInfoUpdate struct {
	Status  *ResourceStatus
	Endtime *time.Time
}

// ResourceInfos records the operation information for each resource's action.
type ResourceInfos []*ResourceInfo

func (infos ResourceInfos) Update(loc ResourceInfoLocator, update ResourceInfoUpdate) bool {
	for _, info := range infos {
		if info.Loc == loc {
			if update.Status != nil {
				info.Status = *update.Status
			}
			if update.Endtime != nil {
				info.EndTime = *update.Endtime
			}
			return true
		}
	}
	return false
}

// ToRows turns the ResourceInfos into table rows.
// The total is used to decorate the index as a fraction, if total > 0.
func (infos ResourceInfos) ToRows(total int) []table.Row {
	t := time.Now()
	var rows []table.Row
	for i, info := range infos {
		idx := strconv.Itoa(i + 1)
		if total > 0 {
			idx = fmt.Sprintf("%d/%d", i+1, total)
		}

		var dur time.Duration
		if info.EndTime.Equal(time.Time{}) {
			dur = t.Sub(info.StartTime).Truncate(time.Second)
		} else {
			dur = info.EndTime.Sub(info.StartTime).Truncate(time.Second)
		}

		module := "-"
		if info.Loc.Module != "" {
			module = info.Loc.Module
		}

		row := []string{
			idx,
			ResourceStatusEmoji(info.Status),
			string(info.Loc.Action),
			module,
			info.Loc.ResourceAddr,
			dur.String(),
		}
		rows = append(rows, row)
	}
	return rows
}

func (infos ResourceInfos) ToColumns(width int) []table.Column {
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
