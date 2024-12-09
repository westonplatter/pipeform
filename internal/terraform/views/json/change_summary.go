package json

type Operation string

const (
	OperationApplied   Operation = "apply"
	OperationDestroyed Operation = "destroy"
	OperationPlanned   Operation = "plan"
)

type ChangeSummary struct {
	Add       int       `json:"add"`
	Change    int       `json:"change"`
	Import    int       `json:"import"`
	Remove    int       `json:"remove"`
	Operation Operation `json:"operation"`
}
