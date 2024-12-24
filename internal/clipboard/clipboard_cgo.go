//go:build cgo

package clipboard

import cb "golang.design/x/clipboard"

type t struct{}

func (c *t) Enabled() bool {
	return cb.Init() == nil
}

func (c *t) Write(b []byte) {
	cb.Write(cb.FmtText, b)
}

func NewClipboard() Clipboard {
	return &t{}
}
