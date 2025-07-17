package ramio

import (
	"data_ram/ramstream"
	"fmt"
	"net"
)

// TCPListener implements Listener for TCP connections.
type TCPStream struct {
	Address        string
	listener       net.Listener
	StreamType     string
	InternalStream ramstream.RamStream
}

func (t *TCPStream) Listen() error {
	// Check stream type
	if t.StreamType != ramstream.DROutputStream {
		return fmt.Errorf("Cannot listen on input stream")
	}

	ln, err := net.Listen("tcp", t.Address)
	if err != nil {
		return err
	}
	t.listener = ln
	fmt.Println("Listening on", t.Address)
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 4096)
			n, err := c.Read(buf)
			if err == nil && n > 0 {
				t.InternalStream.Write(buf[:n])
			}
		}(conn)
	}
}

func (t *TCPStream) Send(data []byte) (int, error) {
	conn, err := net.Dial("tcp", t.Address)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	n, err := conn.Write(data)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (t *TCPStream) Read(p []byte) (int, error) {
	// This should never be called - this stream is fed by the listener
	return 0, fmt.Errorf("Read should not be called on TCPStream")
}

func (t *TCPStream) Write(p []byte) (int, error) {
	// Check to see we are an output stream
	if t.StreamType != ramstream.DROutputStream {
		return 0, fmt.Errorf("Cannot write to input stream")
	}
	if t.InternalStream == nil {
		return 0, fmt.Errorf("Internal stream is not initialized")
	}
	return t.Send(p)
}

func (t *TCPStream) Reset() error {
	t.InternalStream.Reset()
	return nil
}

func (t *TCPStream) Len() int {
	return t.InternalStream.Len()
}

func (t *TCPStream) Flush() error {
	return nil
}

var _ ramstream.RamStream = (*TCPStream)(nil)
