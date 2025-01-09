package ui

import (
	"strings"

	"github.com/magodo/pipeform/internal/terraform/views/json"
)

type Diags []json.Diagnostic

func (diags Diags) HasError() bool {
	for _, diag := range diags {
		if strings.EqualFold(diag.Severity, "error") {
			return true
		}
	}
	return false
}
