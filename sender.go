package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	quic "github.com/quic-go/quic-go"
)

func runQuicSender(addr string, meta FileMetadata, stream Stream, timeoutSeconds int, tlsConf *tls.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Start QUIC server for data transfer
	listener, err := quic.ListenAddr(meta.QuicAddr, tlsConf, nil)
	if err != nil {
		fmt.Println("Sender QUIC listen error:", err)
		return err
	}
	defer listener.Close()

	// Connect to listener via TCP for metadata
	conn, err := tls.Dial("tcp", addr, tlsConf)
	if err != nil {
		fmt.Println("Sender TCP Dial error:", err)
		return err
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(meta); err != nil {
		fmt.Println("Sender Metadata send error:", err)
		return err
	}
	// Wait for QUIC session from receiver
	session, err := listener.Accept(ctx)
	if err != nil {
		fmt.Println("Sender QUIC accept error:", err)
		return err
	}
	defer session.CloseWithError(0, "")
	progress := &Progress{TotalChunks: int((meta.Size + int64(meta.ChunkSize) - 1) / int64(meta.ChunkSize)), Sent: make(map[int]bool)}
	for {
		streamRecv, err := session.AcceptStream(ctx)
		if err != nil {
			break
		}
		var req map[string]int
		if err := json.NewDecoder(streamRecv).Decode(&req); err != nil {
			break
		}
		idx := req["request_chunk"]
		streamSend, err := session.OpenStreamSync(ctx)
		if err != nil {
			break
		}
		if meta.Availability != nil && !meta.Availability[idx] {
			// Not available, respond with Ok: false
			json.NewEncoder(streamSend).Encode(struct {
				ChunkIndex int    `json:"chunk_index"`
				Data       []byte `json:"data"`
				Ok         bool   `json:"ok"`
			}{ChunkIndex: idx, Data: nil, Ok: false})
			streamSend.Close()
			continue
		}
		buf := make([]byte, meta.ChunkSize)
		stream.SeekAbsolute(int64(idx * meta.ChunkSize))
		n, _ := stream.Read(buf)
		json.NewEncoder(streamSend).Encode(struct {
			ChunkIndex int    `json:"chunk_index"`
			Data       []byte `json:"data"`
			Ok         bool   `json:"ok"`
		}{ChunkIndex: idx, Data: buf[:n], Ok: true})
		streamSend.Close()
		progress.Sent[idx] = true
		fmt.Printf("Sent chunk %d/%d\n", idx+1, progress.TotalChunks)
	}
	fmt.Println("File send complete.")
	// Close the QUIC session
	if err := session.CloseWithError(0, ""); err != nil {
		fmt.Println("Sender QUIC session close error:", err)
		return err
	}
	if err := conn.Close(); err != nil {
		fmt.Println("Sender TCP connection close error:", err)
		return err
	}
	return nil
}

func runTcpSender(addr string, meta FileMetadata, stream Stream, timeoutSeconds int, tlsConf *tls.Config) error {
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: time.Duration(timeoutSeconds) * time.Second}, "tcp", addr, tlsConf)
	if err != nil {
		fmt.Println("Sender TCP ONLY Dial error:", err)
		return err
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(meta); err != nil {
		fmt.Println("Sender TCP Metadata send error:", err)
		return err
	}
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	progress := &Progress{TotalChunks: int((meta.Size + int64(meta.ChunkSize) - 1) / int64(meta.ChunkSize)), Sent: make(map[int]bool)}
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("Sender TCP timed out waiting for chunk requests")
		}
		conn.SetReadDeadline(deadline)
		var req map[string]int
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Sender TCP Request decode error:", err)
			return err
		}
		idx := req["request_chunk"]
		if meta.Availability != nil && !meta.Availability[idx] {
			// Not available, respond with Ok: false
			enc.Encode(struct {
				ChunkIndex int    `json:"chunk_index"`
				Data       []byte `json:"data"`
				Ok         bool   `json:"ok"`
			}{ChunkIndex: idx, Data: nil, Ok: false})
			continue
		}
		buf := make([]byte, meta.ChunkSize)
		stream.SeekAbsolute(int64(idx * meta.ChunkSize))
		n, _ := stream.Read(buf)
		enc.Encode(struct {
			ChunkIndex int    `json:"chunk_index"`
			Data       []byte `json:"data"`
			Ok         bool   `json:"ok"`
		}{ChunkIndex: idx, Data: buf[:n], Ok: true})
		progress.Sent[idx] = true
		fmt.Printf("Sent chunk %d/%d\n", idx+1, progress.TotalChunks)
	}
	fmt.Println("File send complete over TCP.")
	if err := conn.Close(); err != nil {
		fmt.Println("TCP ONLY Sender connection close error:", err)
		return err
	}
	return nil
}

// Sender (Client)
func runSender(addr string, meta FileMetadata, stream Stream, timeoutSeconds int, tlsConf *tls.Config) error {
	if meta.Transport == QUIC_S {
		return runQuicSender(addr, meta, stream, timeoutSeconds, tlsConf)
	} else if meta.Transport == TCP_S {
		return runTcpSender(addr, meta, stream, timeoutSeconds, tlsConf)
	} else {
		return fmt.Errorf("unsupported transport: %s", meta.Transport)
	}
}
