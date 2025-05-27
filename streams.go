package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TestStream: In-memory stream with absolute-position seeking
type TestStream struct {
	buffer []byte
	pos    int64
}

func (ts *TestStream) Write(p []byte) (int, error) {
	endPos := ts.pos + int64(len(p))
	if endPos > int64(len(ts.buffer)) {
		newBuf := make([]byte, endPos)
		copy(newBuf, ts.buffer)
		ts.buffer = newBuf
	}
	copy(ts.buffer[ts.pos:], p)
	ts.pos += int64(len(p))
	return len(p), nil
}

func (ts *TestStream) Read(p []byte) (int, error) {
	if ts.pos >= int64(len(ts.buffer)) {
		return 0, io.EOF
	}
	n := copy(p, ts.buffer[ts.pos:])
	ts.pos += int64(n)
	return n, nil
}

func (ts *TestStream) SeekAbsolute(pos int64) error {
	if pos < 0 {
		return fmt.Errorf("negative position")
	}
	ts.pos = pos
	return nil
}

// FileStream: Wraps *os.File for absolute-position seek only
type FileStream struct {
	file *os.File
}

func NewFileStream(filename string) (*FileStream, error) {
	dir := GlobalSettings.StorageDir
	if dir == "" {
		dir = "."
	}
	fullPath := filepath.Join(dir, filename)
	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &FileStream{file: f}, nil
}

func (fs *FileStream) Write(p []byte) (int, error) {
	return fs.file.Write(p)
}

func (fs *FileStream) Read(p []byte) (int, error) {
	return fs.file.Read(p)
}

func (fs *FileStream) SeekAbsolute(pos int64) error {
	_, err := fs.file.Seek(pos, io.SeekStart)
	return err
}

func (fs *FileStream) Close() error {
	return fs.file.Close()
}

// Define our own interface with absolute-position seeking
type Stream interface {
	io.Reader
	io.Writer
	SeekAbsolute(pos int64) error
}
