package ramio

import (
	"data_ram/ramstream"
	"fmt"
	"net"
)

// Helper to read a big-endian int64 from a connection
func readInt64(conn net.Conn) (int64, error) {
	buf := make([]byte, 8)
	_, err := conn.Read(buf)
	if err != nil {
		return 0, err
	}
	val := int64(buf[0])<<56 | int64(buf[1])<<48 | int64(buf[2])<<40 | int64(buf[3])<<32 |
		int64(buf[4])<<24 | int64(buf[5])<<16 | int64(buf[6])<<8 | int64(buf[7])
	return val, nil
}

// Helper to write a big-endian int64 to a connection
func writeInt64(conn net.Conn, val int64) error {
	buf := []byte{
		byte(val >> 56), byte(val >> 48), byte(val >> 40), byte(val >> 32),
		byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
	}
	_, err := conn.Write(buf)
	return err
}

// TCPListener implements Listener for TCP connections.
type TCPStream struct {
	Address        string
	listener       net.Listener
	StreamType     string
	InternalStream ramstream.RamStream
}

// Constructor for TCPStream that sets the address and stream type.
func NewTCPStream(address string, streamType string, internalStream ramstream.RamStream) *TCPStream {
	return &TCPStream{
		Address:        address,
		StreamType:     streamType,
		InternalStream: internalStream,
	}
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
				outN, _ := t.InternalStream.Write(buf[:n])
				writeInt64(c, int64(outN))
			}
		}(conn)
	}
}

// Update Listen to respond with bytes written
// (Replace the goroutine in Listen)
// In Listen, after writing to InternalStream:
// writeInt64(c, int64(outN))

// Update Send to wait for response
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
	// Wait for response
	respN, err := readInt64(conn)
	if err != nil {
		return 0, fmt.Errorf("Failed to read response: %w", err)
	}
	if int(respN) != n {
		return 0, fmt.Errorf("Could not write over tcp stream: sent %d, got response %d", n, respN)
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
