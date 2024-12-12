package ui

import "github.com/magodo/pipeform/internal/terraform/views"

type receiverMsg struct {
	msg views.Message
}

type receiverEOFMsg struct{}

type receiverErrorMsg struct {
	err error
}

func (m receiverErrorMsg) Error() string {
	return m.err.Error()
}
