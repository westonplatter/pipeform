package ui

import (
	"time"

	"github.com/magodo/pipeform/internal/terraform/views/json"
)

type ResourceState string

const (
	// Initial state
	ResourceStateIdle ResourceState = "idle"
	// Once received one OperationStart hook message
	ResourceStateStart ResourceState = "start"
	// Once received one or more OperationProgress hook message
	ResourceStateProgress ResourceState = "progress"
	// Once received one OperationComplete hook message
	ResourceStateComplete ResourceState = "complete"
	// Once received one OperationErrored hook message
	ResourceStateErrored ResourceState = "error"

	// TODO: Support refresh? (refresh is an independent lifecycle than the resource apply lifecycle)
	// TODO: Support provision? (provision is a intermidiate stage in the resource apply lifecycle)
)

type ResourceInfo struct {
	ResourceAddr string
	Action       json.ChangeAction
	State        ResourceState
	StartTime    time.Time
	EndTime      time.Time
	Diag         *json.Diagnostic
}

// ResourceInfos records the operation information for each resource's action.
// The first key is the ResourceAddr.
// The second key is the resource Action.
// It can happen that one single resource have more than one actions conducted in one apply,
// e.g., a resource being re-created (remove + create).
type ResourceInfos map[string]map[json.ChangeAction]*ResourceInfo
