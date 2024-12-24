package clipboard

type Clipboard interface {
	Enabled() bool
	Write([]byte)
}
