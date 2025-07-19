package ramio

import (
	"context"
	"data_ram/ramstream"
	"fmt"
	"net"
)

// Define const header bytes
const (
	TCP_KEEPALIVE = 1 // Size of the header for int64 response
	TCP_DATA      = 2
	TCP_CLOSE     = 3
)

// TCPListener implements Listener for TCP connections.
type TCPStream struct {
	Address        string
	listener       net.Listener
	StreamType     string
	InternalStream ramstream.RamStream
	tcpCon         net.Conn
	cancelListen   context.CancelFunc // Added to cancel goroutines
}

// Constructor for TCPStream that sets the address and stream type.
func NewTCPStream(address string, streamType string, internalStream ramstream.RamStream) *TCPStream {
	return &TCPStream{
		Address:        address,
		StreamType:     streamType,
		InternalStream: internalStream,
	}
}

func (t *TCPStream) handleListen(ctx context.Context, c net.Conn, bufferSize int) {
	defer c.Close()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("TCP handler cancelled")
			return
		default:
			respN, err := readInt64(c)
			if err != nil {
				fmt.Printf("TCP read failed: %v\n", err)
				return
			}
			switch respN {
			case TCP_CLOSE:
				return
			case TCP_KEEPALIVE:
				writeInt64(c, int64(TCP_KEEPALIVE))
			case TCP_DATA:
				buf := make([]byte, bufferSize)
				n, err := c.Read(buf)
				if err == nil && n > 0 {
					outN, _ := t.InternalStream.Write(buf[:n])
					writeInt64(c, int64(outN))
				} else {
					fmt.Printf("TCP failed to read from stream; closing\n")
					return
				}
			default:
				fmt.Printf("Unknown TCP header type: %d\n", respN)
				return
			}
		}
	}
}

func (t *TCPStream) Listen(bufferSize int) error {
	// Check stream type
	if t.StreamType != ramstream.DROutputStream {
		return fmt.Errorf("Cannot listen on input stream")
	}

	ln, err := net.Listen("tcp", t.Address)
	if err != nil {
		fmt.Println("Failed to listen on", t.Address)
		return err
	}
	t.listener = ln
	fmt.Println("Listening on", t.Address)

	ctx, cancel := context.WithCancel(context.Background())
	t.cancelListen = cancel // Store cancel func to use in Flush

	var handleListenRunning bool

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err // this will happen when listener is closed
		}
		if handleListenRunning {
			fmt.Println("Stopping new connection attempted while another is running on address", t.Address)
			conn.Close()
			continue
		}
		handleListenRunning = true
		go func() {
			t.handleListen(ctx, conn, bufferSize)
			handleListenRunning = false
		}()
	}
}

// Update Send to wait for response
func (t *TCPStream) Send(data []byte) (int, error) {
	if t.tcpCon != nil && t.tcpCon.RemoteAddr() != nil {
		err := writeInt64(t.tcpCon, int64(TCP_KEEPALIVE))
		resp, err := readInt64(t.tcpCon)
		if err != nil || resp != int64(TCP_KEEPALIVE) {
			t.tcpCon.Close()
			t.tcpCon = nil
			fmt.Printf("TCP keepalive failed: %v\n", err)
		}
	}

	if t.tcpCon == nil {
		con, err := net.Dial("tcp", t.Address)
		if err != nil {
			fmt.Printf("TCP dial failed: %v\n", err)
			return 0, err
		}
		t.tcpCon = con
	}
	err := writeInt64(t.tcpCon, int64(TCP_DATA))
	if err != nil {
		return 0, err
	}
	n, err := t.tcpCon.Write(data)
	if err != nil {
		return n, err
	}
	// Wait for response
	respN, err := readInt64(t.tcpCon)
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
	if t.InternalStream != nil {
		return 0, fmt.Errorf("Internal stream should not be initialised for TCP writing") // This should not be set for writing
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
	if t.listener != nil {
		t.listener.Close()
		t.listener = nil
	}
	if t.cancelListen != nil {
		t.cancelListen()
		t.cancelListen = nil
	}
	return nil
}

var _ ramstream.RamStream = (*TCPStream)(nil)
