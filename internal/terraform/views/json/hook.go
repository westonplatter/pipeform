package json

type Hooker interface {
	isHooker()
}

type hooker struct{}

func (hooker) isHooker() {}

// OperationStart: triggered by Pre{Apply,EphemeralOp} hook
// msgType can be:
// - MessageApplyStart
// - MessageEphemeralOpStart
type OperationStart struct {
	hooker

	Resource ResourceAddr `json:"resource"`
	Action   ChangeAction `json:"action"`
	IDKey    string       `json:"id_key,omitempty"`
	IDValue  string       `json:"id_value,omitempty"`
}

// OperationProgress: currently triggered by a timer started on Pre{Apply,EphemeralOp}. In
// future, this might also be triggered by provider progress reporting.
// msgType can be:
// - MessageApplyProgress
// - MessageEphemeralOpProgress
type OperationProgress struct {
	hooker

	Resource ResourceAddr `json:"resource"`
	Action   ChangeAction `json:"action"`
	Elapsed  float64      `json:"elapsed_seconds"`
}

// OperationComplete: triggered by PostApply hook
// msgType can be:
// - MessageApplyComplete
// - MessageEphemeralOpComplete
type OperationComplete struct {
	hooker

	Resource ResourceAddr `json:"resource"`
	Action   ChangeAction `json:"action"`
	IDKey    string       `json:"id_key,omitempty"`
	IDValue  string       `json:"id_value,omitempty"`
	Elapsed  float64      `json:"elapsed_seconds"`
}

// OperationErrored: triggered by PostApply hook on failure. This will be followed
// by diagnostics when the apply finishes.
// msgType can be:
// - MessageApplyErrored
// - MessageEphemeralOpErrored
type OperationErrored struct {
	hooker

	Resource ResourceAddr `json:"resource"`
	Action   ChangeAction `json:"action"`
	Elapsed  float64      `json:"elapsed_seconds"`
}

// ProvisionStart: triggered by PreProvisionInstanceStep hook
// msgType can be:
// - MessageProvisionStart
type ProvisionStart struct {
	hooker

	Resource    ResourceAddr `json:"resource"`
	Provisioner string       `json:"provisioner"`
}

// ProvisionProgress: triggered by ProvisionOutput hook
// msgType can be:
// - MessageProvisionProgress
type ProvisionProgress struct {
	hooker

	Resource    ResourceAddr `json:"resource"`
	Provisioner string       `json:"provisioner"`
	Output      string       `json:"output"`
}

// ProvisionComplete: triggered by PostProvisionInstanceStep hook
// msgType can be:
// - MessageProvisionComplete
type ProvisionComplete struct {
	hooker

	Resource    ResourceAddr `json:"resource"`
	Provisioner string       `json:"provisioner"`
}

// ProvisionErrored: triggered by PostProvisionInstanceStep hook on failure.
// This will be followed by diagnostics when the apply finishes.
// msgType can be:
// - MessageProvisionErrored
type ProvisionErrored struct {
	hooker

	Resource    ResourceAddr `json:"resource"`
	Provisioner string       `json:"provisioner"`
}

// RefreshStart: triggered by PreRefresh hook
// msgType can be:
// - MessageRefreshStart
type RefreshStart struct {
	hooker

	Resource ResourceAddr `json:"resource"`
	IDKey    string       `json:"id_key,omitempty"`
	IDValue  string       `json:"id_value,omitempty"`
}

// RefreshComplete: triggered by PostRefresh hook
// msgType can be:
// - MessageRefreshComplete
type RefreshComplete struct {
	hooker

	Resource ResourceAddr `json:"resource"`
	IDKey    string       `json:"id_key,omitempty"`
	IDValue  string       `json:"id_value,omitempty"`
}
