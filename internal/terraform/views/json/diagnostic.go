package json

const (
	DiagnosticSeverityUnknown = "unknown"
	DiagnosticSeverityError   = "error"
	DiagnosticSeverityWarning = "warning"
)

type Diagnostic struct {
	Severity string             `json:"severity"`
	Summary  string             `json:"summary"`
	Detail   string             `json:"detail"`
	Address  string             `json:"address,omitempty"`
	Range    *DiagnosticRange   `json:"range,omitempty"`
	Snippet  *DiagnosticSnippet `json:"snippet,omitempty"`
}

type Pos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Byte   int `json:"byte"`
}

type DiagnosticRange struct {
	Filename string `json:"filename"`
	Start    Pos    `json:"start"`
	End      Pos    `json:"end"`
}

type DiagnosticSnippet struct {
	Context              *string                     `json:"context"`
	Code                 string                      `json:"code"`
	StartLine            int                         `json:"start_line"`
	HighlightStartOffset int                         `json:"highlight_start_offset"`
	HighlightEndOffset   int                         `json:"highlight_end_offset"`
	Values               []DiagnosticExpressionValue `json:"values"`
	FunctionCall         *DiagnosticFunctionCall     `json:"function_call,omitempty"`
}

type DiagnosticExpressionValue struct {
	Traversal string `json:"traversal"`
	Statement string `json:"statement"`
}

type DiagnosticFunctionCall struct {
	CalledAs  string    `json:"called_as"`
	Signature *Function `json:"signature,omitempty"`
}
