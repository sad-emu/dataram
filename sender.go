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

func runQuicSender(addr string, meta FileMetadata, stream Stream, timeoutSeconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Start QUIC server for data transfer
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{loadSenderCert()},
		NextProtos:         []string{"dataram"},
		ServerName:         "localhost",
	}
	listener, err := quic.ListenAddr(meta.QuicAddr, tlsConf, nil)
	if err != nil {
		fmt.Println("Sender QUIC listen error:", err)
		return err
	}
	defer listener.Close()

	// Connect to listener via TCP for metadata
	conn, err := net.Dial("tcp", addr)
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
		buf := make([]byte, meta.ChunkSize)
		stream.SeekAbsolute(int64(idx * meta.ChunkSize))
		n, _ := stream.Read(buf)
		chunk := struct {
			ChunkIndex int    `json:"chunk_index"`
			Data       []byte `json:"data"`
		}{ChunkIndex: idx, Data: buf[:n]}
		streamSend, err := session.OpenStreamSync(ctx)
		if err != nil {
			break
		}
		json.NewEncoder(streamSend).Encode(chunk)
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
	// Close the TCP connection
	if err := conn.Close(); err != nil {
		fmt.Println("Sender TCP connection close error:", err)
		return err
	}
	return nil
}

func runTcpSender(addr string, meta FileMetadata, stream Stream, timeoutSeconds int) error {
	conn, err := net.DialTimeout("tcp", addr, time.Duration(timeoutSeconds)*time.Second)
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
		buf := make([]byte, meta.ChunkSize)
		stream.SeekAbsolute(int64(idx * meta.ChunkSize))
		n, _ := stream.Read(buf)
		chunk := struct {
			ChunkIndex int    `json:"chunk_index"`
			Data       []byte `json:"data"`
		}{ChunkIndex: idx, Data: buf[:n]}
		if err := enc.Encode(chunk); err != nil {
			fmt.Println("Sender TCP Chunk send error:", err)
			return err
		}
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
func runSender(addr string, meta FileMetadata, stream Stream, timeoutSeconds int) error {
	if meta.Transport == QUIC_S {
		return runQuicSender(addr, meta, stream, timeoutSeconds)
	} else if meta.Transport == TCP_S {
		return runTcpSender(addr, meta, stream, timeoutSeconds)
	} else {
		return fmt.Errorf("unsupported transport: %s", meta.Transport)
	}
}

// Helper to load the server certificate
func loadSenderCert() tls.Certificate {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		panic("failed to load server TLS cert: " + err.Error())
	}
	return cert
}
