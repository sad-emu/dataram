package ramio

import (
	"context"
	"crypto/tls"
	"data_ram/ramstream"
	"fmt"

	quic "github.com/quic-go/quic-go"
)

// QUICStream implements a stream over QUIC with similar API to TCPStream.
type QUICStream struct {
	Address        string
	StreamType     string
	InternalStream ramstream.RamStream
	tlsConfig      *tls.Config
}

func NewQUICStream(address string, streamType string, internalStream ramstream.RamStream,
	tlsConfigIn *tls.Config) *QUICStream {
	return &QUICStream{
		Address:        address,
		StreamType:     streamType,
		InternalStream: internalStream,
		tlsConfig:      tlsConfigIn,
	}
}

func (q *QUICStream) Listen() error {
	if q.StreamType != ramstream.DROutputStream {
		return fmt.Errorf("Cannot listen on input stream")
	}
	listener, err := quic.ListenAddr(q.Address, q.tlsConfig, nil)
	if err != nil {
		return err
	}

	fmt.Println("QUIC Listening on", q.Address)
	for {
		sess, err := listener.Accept(context.Background())
		if err != nil {
			return err
		}
		go func(session *quic.Conn) {
			stream, err := session.AcceptStream(context.Background())
			if err != nil {
				return
			}
			buf := make([]byte, 4096)
			n, err := stream.Read(buf)
			if err == nil && n > 0 {
				outN, _ := q.InternalStream.Write(buf[:n])
				stream.Write([]byte(fmt.Sprintf("%d", outN)))
			}
		}(sess)
	}
}

func (q *QUICStream) Send(data []byte) (int, error) {
	sess, err := quic.DialAddr(context.Background(), q.Address, q.tlsConfig, nil)
	if err != nil {
		return 0, err
	}
	streamSend, err := sess.OpenStreamSync(context.Background())
	n, err := streamSend.Write(data)
	if err != nil {
		return n, err
	}
	resp := make([]byte, 16)
	respN, err := streamSend.Read(resp)
	if err != nil {
		return n, fmt.Errorf("Failed to read response: %w", err)
	}
	// Parse response as int
	var outN int
	fmt.Sscanf(string(resp[:respN]), "%d", &outN)
	if outN != n {
		return n, fmt.Errorf("Could not write over quic stream: sent %d, got response %d", n, outN)
	}
	return n, nil
}

func (q *QUICStream) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("Read should not be called on QUICStream")
}

func (q *QUICStream) Write(p []byte) (int, error) {
	if q.StreamType != ramstream.DROutputStream {
		return 0, fmt.Errorf("Cannot write to input stream")
	}
	if q.InternalStream != nil {
		return 0, fmt.Errorf("Internal stream should not be initialised for QUIC writing")
	}
	return q.Send(p)
}

func (q *QUICStream) Reset() error {
	q.InternalStream.Reset()
	return nil
}

func (q *QUICStream) Len() int {
	return q.InternalStream.Len()
}

func (q *QUICStream) Flush() error {
	return nil
}

var _ ramstream.RamStream = (*QUICStream)(nil)
