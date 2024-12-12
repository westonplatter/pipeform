package reader

import (
	"bufio"
	"io"

	"github.com/magodo/pipeform/internal/terraform/views"
)

type Reader struct {
	scanner   *bufio.Scanner
	teeWriter io.Writer
}

func NewReader(r io.Reader, teeWriter io.Writer) Reader {
	return Reader{
		scanner:   bufio.NewScanner(r),
		teeWriter: teeWriter,
	}
}

// Next returns the message.
// Otherwise, it returns either the io.EOF error, or others.
func (r *Reader) Next() (views.Message, error) {
	if r.scanner.Scan() {
		line := r.scanner.Text()
		io.WriteString(r.teeWriter, line+"\n")
		return views.UnmarshalMessage([]byte(line))
	}
	if err := r.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}
