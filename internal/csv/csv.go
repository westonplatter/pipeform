package csv

import (
	"strings"

	"github.com/magodo/pipeform/internal/state"
)

type Input struct {
	RefreshInfos state.ResourceOperationInfos
	ApplyInfos   state.ResourceOperationInfos
}

func ToCsv(input Input) []byte {
	out := []string{
		strings.Join([]string{
			"Start Timestamp",
			"End Timestamp",
			"Stage",
			"Action",
			"Module",
			"Resource Type",
			"Resource Name",
			"Resource Key",
			"Status",
			"Duration (sec)",
		}, ","),
	}
	out = append(out, input.RefreshInfos.ToCsv("refresh")...)
	out = append(out, input.ApplyInfos.ToCsv("apply")...)
	return []byte(strings.Join(out, "\n"))
}
