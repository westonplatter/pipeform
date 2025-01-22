package state

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/magodo/pipeform/internal/terraform/views/json"
)

type ResourceOperationStatus string

const (
	// Once received one OperationStart hook message
	ResourceOperationStatusStart ResourceOperationStatus = "start"
	// Once received one OperationComplete hook message
	ResourceOperationStatusComplete ResourceOperationStatus = "complete"
	// Once received one OperationErrored hook message
	ResourceOperationStatusErrored ResourceOperationStatus = "error"

	// TODO: Support refresh? (refresh is an independent lifecycle than the resource apply lifecycle)
	// TODO: Support provision? (provision is a intermidiate stage in the resource apply lifecycle)
)

func resourceOperationStatusEmoji(status ResourceOperationStatus) string {
	switch status {
	case ResourceOperationStatusStart:
		return "🕛"
	case ResourceOperationStatusComplete:
		return "✅"
	case ResourceOperationStatusErrored:
		return "❌"
	default:
		return "❓"
	}
}

type ResourceOperationInfoLocator struct {
	Module       string
	ResourceAddr string
	Action       string
}

type ResourceOperationInfo struct {
	Idx             int
	RawResourceAddr json.ResourceAddr
	Loc             ResourceOperationInfoLocator
	Status          ResourceOperationStatus
	StartTime       time.Time
	EndTime         time.Time
}

type ResourceOperationInfoUpdate struct {
	Status  *ResourceOperationStatus
	Endtime *time.Time
}

// ResourceOperationInfos records the operation information for each resource's action.
type ResourceOperationInfos []*ResourceOperationInfo

func (infos ResourceOperationInfos) Find(loc ResourceOperationInfoLocator) *ResourceOperationInfo {
	for _, info := range infos {
		if info.Loc == loc {
			return info
		}
	}
	return nil
}

func (infos ResourceOperationInfos) Update(loc ResourceOperationInfoLocator, update ResourceOperationInfoUpdate) *ResourceOperationInfo {
	info := infos.Find(loc)
	if info == nil {
		return nil
	}
	if update.Status != nil {
		info.Status = *update.Status
	}
	if update.Endtime != nil {
		info.EndTime = *update.Endtime
	}
	return info
}

// ToRows turns the ResourceInfos into table rows.
// The total is used to decorate the index as a fraction, if total > 0.
func (infos ResourceOperationInfos) ToRows(total int) []table.Row {
	now := time.Now()
	var rows []table.Row
	for _, info := range infos {
		idx := strconv.Itoa(info.Idx)
		if total > 0 {
			idx = fmt.Sprintf("%d/%d", info.Idx, total)
		}

		dur := info.Duration(now)

		module := "-"
		if info.Loc.Module != "" {
			module = info.Loc.Module
		}

		row := []string{
			idx,
			resourceOperationStatusEmoji(info.Status),
			string(info.Loc.Action),
			module,
			info.Loc.ResourceAddr,
			dur.String(),
		}
		rows = append(rows, row)
	}
	return rows
}

func (infos ResourceOperationInfos) ToColumns(width int) []table.Column {
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

func (infos ResourceOperationInfos) ToCsv(stage string) []string {
	var out []string
	now := time.Now()
	for _, info := range infos {
		key, _ := info.RawResourceAddr.ResourceKey.MarshalJSON()
		line := []string{
			strconv.FormatInt(info.StartTime.Unix(), 10),
			strconv.FormatInt(info.EndTime.Unix(), 10),
			stage,
			info.Loc.Action,
			info.Loc.Module,
			info.RawResourceAddr.ResourceType,
			info.RawResourceAddr.ResourceName,
			string(key),
			string(info.Status),
			strconv.FormatInt(int64(info.Duration(now).Seconds()), 10),
		}
		out = append(out, strings.Join(line, ","))
	}
	return out
}

func (info ResourceOperationInfo) Duration(now time.Time) time.Duration {
	var dur time.Duration
	if info.EndTime.Equal(time.Time{}) {
		dur = now.Sub(info.StartTime).Truncate(time.Second)
	} else {
		dur = info.EndTime.Sub(info.StartTime).Truncate(time.Second)
	}
	return dur
}
