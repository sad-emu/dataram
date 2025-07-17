package ramio

import (
	"data_ram/ramstream"
)

// DummyStream for testing purposes
type DummyStream struct {
	StreamType string
	Position   int64
	Data       []byte
}

// Constructor for DummyStream
func NewDummyStream(streamType string) *DummyStream {
	return &DummyStream{
		StreamType: streamType,
		Position:   0,
		Data:       make([]byte, 0),
	}
}

func (d *DummyStream) Read(p []byte) (int, error) {
	if d.StreamType != ramstream.DRInputStream {
		return 0, nil
	}
	if d.Position >= int64(len(d.Data)) {
		return 0, nil
	}
	n := copy(p, d.Data[d.Position:])
	d.Position += int64(n)
	return n, nil
}

func (d *DummyStream) Write(p []byte) (int, error) {
	if d.StreamType != ramstream.DROutputStream {
		return 0, nil
	}
	d.Data = append(d.Data, p...)
	return len(p), nil
}

func (d *DummyStream) Reset() error {
	d.Position = 0
	return nil
}

func (d *DummyStream) Flush() error {
	return nil
}

func (d *DummyStream) Len() int {
	return len(d.Data)
}

var _ ramstream.RamStream = (*DummyStream)(nil)
