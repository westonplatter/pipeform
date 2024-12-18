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
		return "IDLE"
	case ViewStateRefresh:
		return "REFRESH"
	case ViewStatePlan:
		return "PLAN"
	case ViewStateApply:
		return "APPLY"
	case ViewStateSummary:
		return "SUMMARY"
	default:
		return "UNKNOWN"
	}
}

func (s ViewState) NextState(msg views.Message) (ViewState, bool) {
	switch s {
	case ViewStateIdle:
		switch msg.BaseMessage().Type {
		case json.MessageRefreshStart:
			return ViewStateRefresh, true
		case json.MessagePlannedChange:
			return ViewStatePlan, true
		case json.MessageApplyStart:
			return ViewStateApply, true
		case json.MessageChangeSummary:
			// There are two change summary messages, one after change, one after apply.
			// We only handle the one after apply, as the one after change is less interesting to show.
			if msg.(views.ChangeSummaryMsg).Changes.Operation == json.OperationApplied {
				return ViewStateSummary, true
			}
		}

	case ViewStateRefresh:
		switch msg.BaseMessage().Type {
		case json.MessagePlannedChange:
			return ViewStatePlan, true
		case json.MessageChangeSummary:
			if msg.(views.ChangeSummaryMsg).Changes.Operation == json.OperationApplied {
				return ViewStateSummary, true
			}
		}

	case ViewStatePlan:
		switch msg.BaseMessage().Type {
		case json.MessageApplyStart:
			return ViewStateApply, true
		case json.MessageChangeSummary:
			if msg.(views.ChangeSummaryMsg).Changes.Operation == json.OperationApplied {
				return ViewStateSummary, true
			}
		}

	case ViewStateApply:
		switch msg.BaseMessage().Type {
		case json.MessageChangeSummary:
			return ViewStateSummary, true
		}
	}

	return s, false
}
