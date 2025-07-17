package ramio

import (
	"data_ram/ramstream"
	"fmt"
	"io"
	"os"
)

// TCPListener implements Listener for TCP connections.
type LocalStream struct {
	StreamType    string
	LocalFilePath string
	Position      int64
	FileHandle    *os.File
}

// Constructor for localstream that sets the stream type and local file path.
func NewLocalStream(streamType, localFilePath string) *LocalStream {
	return &LocalStream{
		StreamType:    streamType,
		LocalFilePath: localFilePath,
		Position:      0,
	}
}

func (t *LocalStream) Read(p []byte) (int, error) {
	// Check to see we are an input stream
	if t.StreamType != ramstream.DRInputStream {
		return 0, fmt.Errorf("Cannot read from output stream")
	}
	if t.FileHandle == nil {
		var err error
		t.FileHandle, err = os.Open(t.LocalFilePath)
		if err != nil {
			return 0, fmt.Errorf("failed to open file: %w", err)
		}
	}
	n, err := t.FileHandle.ReadAt(p, t.Position)
	if err != nil {
		if err == io.EOF {
			return n, nil // Return number of bytes read before EOF
		}
		return n, fmt.Errorf("failed to read from file: %w", err)
	}
	t.Position += int64(n)
	return n, nil
}

func (t *LocalStream) Write(p []byte) (int, error) {
	// Check to see we are an output stream
	if t.StreamType != ramstream.DROutputStream {
		return 0, fmt.Errorf("Cannot write to input stream")
	}
	if t.FileHandle == nil {
		var err error
		// Try opening the file for writing
		t.FileHandle, err = os.OpenFile(t.LocalFilePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return 0, fmt.Errorf("failed to open or create file: %w", err)
		}
	}
	n, err := t.FileHandle.WriteAt(p, t.Position)
	if err != nil {
		return n, fmt.Errorf("failed to write to file: %w", err)
	}
	t.Position += int64(n)
	return n, nil
}

func (t *LocalStream) Reset() error {
	t.Position = 0
	t.FileHandle = nil // Close the file handle
	return nil
}

func (t *LocalStream) Flush() error {
	return nil
}

func (t *LocalStream) Len() int {
	return int(t.Position)
}

var _ ramstream.RamStream = (*LocalStream)(nil)
