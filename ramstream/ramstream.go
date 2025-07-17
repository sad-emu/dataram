package ramstream

import "io"

// Consts for stream types input and output.

const (
	DRInputStream  = "input"
	DROutputStream = "output"
)

// Stream is a generic interface for reading and writing.
type RamStream interface {
	io.Reader
	io.Writer
	Reset() error
	Len() int
	Flush() error
}
