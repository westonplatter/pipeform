package ui

import (
	"github.com/magodo/pipeform/internal/terraform/views"
	"github.com/magodo/pipeform/internal/terraform/views/json"
)

type ViewState int

const (
	ViewStateUnknown ViewState = iota
	ViewStateIdle
	ViewStateRefresh
	ViewStatePlan
	ViewStateApply
	ViewStateSummary
)

func (s ViewState) String() string {
	switch s {
	case ViewStateIdle:
		return "idle"
	case ViewStateRefresh:
		return "refresh"
	case ViewStatePlan:
		return "plan"
	case ViewStateApply:
		return "apply"
	case ViewStateSummary:
		return "summary"
	default:
		return "unknown"
	}
}

func (s ViewState) NextState(msg views.Message) ViewState {
	switch s {
	case ViewStateIdle:
		switch msg.BaseMessage().Type {
		case json.MessageRefreshStart:
			return ViewStateRefresh
		case json.MessagePlannedChange:
			return ViewStatePlan
		case json.MessageApplyStart:
			return ViewStateApply
		case json.MessageChangeSummary:
			// There are two change summary messages, one after change, one after apply.
			// We only handle the one after apply, as the one after change is less interesting to show.
			if msg.(views.ChangeSummaryMsg).Changes.Operation == json.OperationApplied {
				return ViewStateSummary
			}
		}
		return ViewStateIdle

	case ViewStateRefresh:
		switch msg.BaseMessage().Type {
		case json.MessagePlannedChange:
			return ViewStatePlan
		case json.MessageChangeSummary:
			if msg.(views.ChangeSummaryMsg).Changes.Operation == json.OperationApplied {
				return ViewStateSummary
			}
		}
		return ViewStateRefresh
	case ViewStatePlan:
		switch msg.BaseMessage().Type {
		case json.MessageApplyStart:
			return ViewStateApply
		case json.MessageChangeSummary:
			if msg.(views.ChangeSummaryMsg).Changes.Operation == json.OperationApplied {
				return ViewStateSummary
			}
		}
		return ViewStatePlan
	case ViewStateApply:
		switch msg.BaseMessage().Type {
		case json.MessageChangeSummary:
			return ViewStateSummary
		}
		return ViewStateApply
	case ViewStateSummary:
		return ViewStateSummary
	default:
		return ViewStateUnknown
	}
}
