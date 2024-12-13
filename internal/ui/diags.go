package ui

import "github.com/magodo/pipeform/internal/terraform/views/json"

type Diags struct {
	Warns []json.Diagnostic
	Errs  []json.Diagnostic
}
