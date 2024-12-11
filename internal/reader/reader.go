package reader

import (
	"bufio"
	"io"

	"github.com/magodo/pipeform/internal/terraform/views"
)

type Reader struct {
	scanner *bufio.Scanner
}

func NewReader(r io.Reader) Reader {
	return Reader{
		scanner: bufio.NewScanner(r),
	}
}

// Next returns the message.
// Otherwise, it returns either the io.EOF error, or others.
func (r *Reader) Next() (views.Message, error) {
	if r.scanner.Scan() {
		return views.UnmarshalMessage([]byte(r.scanner.Text()))
	}
	if err := r.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}
