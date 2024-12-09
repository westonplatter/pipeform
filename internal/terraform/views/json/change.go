package json

type ResourceInstanceChange struct {
	Resource         ResourceAddr  `json:"resource"`
	PreviousResource *ResourceAddr `json:"previous_resource,omitempty"`
	Action           ChangeAction  `json:"action"`
	Reason           ChangeReason  `json:"reason,omitempty"`
	Importing        *Importing    `json:"importing,omitempty"`
	GeneratedConfig  string        `json:"generated_config,omitempty"`
}

type ChangeAction string

const (
	ActionNoOp    ChangeAction = "noop"
	ActionMove    ChangeAction = "move"
	ActionForget  ChangeAction = "remove"
	ActionCreate  ChangeAction = "create"
	ActionRead    ChangeAction = "read"
	ActionUpdate  ChangeAction = "update"
	ActionReplace ChangeAction = "replace"
	ActionDelete  ChangeAction = "delete"
	ActionImport  ChangeAction = "import"

	// While ephemeral resources do not represent a change
	// or participate in the plan in the same way as the above
	// we declare them here for convenience in helper functions.
	ActionOpen  ChangeAction = "open"
	ActionRenew ChangeAction = "renew"
	ActionClose ChangeAction = "close"
)

type ChangeReason string

const (
	ReasonNone               ChangeReason = ""
	ReasonTainted            ChangeReason = "tainted"
	ReasonRequested          ChangeReason = "requested"
	ReasonReplaceTriggeredBy ChangeReason = "replace_triggered_by"
	ReasonCannotUpdate       ChangeReason = "cannot_update"
	ReasonUnknown            ChangeReason = "unknown"

	ReasonDeleteBecauseNoResourceConfig ChangeReason = "delete_because_no_resource_config"
	ReasonDeleteBecauseWrongRepetition  ChangeReason = "delete_because_wrong_repetition"
	ReasonDeleteBecauseCountIndex       ChangeReason = "delete_because_count_index"
	ReasonDeleteBecauseEachKey          ChangeReason = "delete_because_each_key"
	ReasonDeleteBecauseNoModule         ChangeReason = "delete_because_no_module"
	ReasonDeleteBecauseNoMoveTarget     ChangeReason = "delete_because_no_move_target"
	ReasonReadBecauseConfigUnknown      ChangeReason = "read_because_config_unknown"
	ReasonReadBecauseDependencyPending  ChangeReason = "read_because_dependency_pending"
	ReasonReadBecauseCheckNested        ChangeReason = "read_because_check_nested"
)
