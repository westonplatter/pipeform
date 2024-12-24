//go:build !cgo

package clipboard

type t struct{}

func (c *t) Enabled() bool {
	return false
}

func (c *t) Write(b []byte) {
	return
}

func NewClipboard() Clipboard {
	return &t{}
}
