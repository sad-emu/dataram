package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"

	quic "github.com/quic-go/quic-go"
)

func runQuicSender(addr string, meta FileMetadata, stream Stream) {

	// Start QUIC server for data transfer
	tlsConf := &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"dataram"}}
	listener, err := quic.ListenAddr(meta.QuicAddr, tlsConf, nil)
	if err != nil {
		fmt.Println("QUIC listen error:", err)
		return
	}
	defer listener.Close()

	// Connect to listener via TCP for metadata
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(meta); err != nil {
		fmt.Println("Metadata send error:", err)
		return
	}
	// Wait for QUIC session from receiver
	session, err := listener.Accept(context.Background())
	if err != nil {
		fmt.Println("QUIC accept error:", err)
		return
	}
	defer session.CloseWithError(0, "")
	progress := &Progress{TotalChunks: int((meta.Size + int64(meta.ChunkSize) - 1) / int64(meta.ChunkSize)), Sent: make(map[int]bool)}
	for {
		streamRecv, err := session.AcceptStream(context.Background())
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
		streamSend, err := session.OpenStreamSync(context.Background())
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
		fmt.Println("QUIC session close error:", err)
		return
	}
	// Close the TCP connection
	if err := conn.Close(); err != nil {
		fmt.Println("TCP connection close error:", err)
		return
	}
}

func runTcpSender(addr string, meta FileMetadata, stream Stream) {
	// Default: TCP
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(meta); err != nil {
		fmt.Println("Metadata send error:", err)
		return
	}
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	progress := &Progress{TotalChunks: int((meta.Size + int64(meta.ChunkSize) - 1) / int64(meta.ChunkSize)), Sent: make(map[int]bool)}
	for {
		var req map[string]int
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Request decode error:", err)
			return
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
			fmt.Println("Chunk send error:", err)
			return
		}
		progress.Sent[idx] = true
		fmt.Printf("Sent chunk %d/%d\n", idx+1, progress.TotalChunks)
	}
	fmt.Println("File send complete.")
	// Close the TCP connection
	if err := conn.Close(); err != nil {
		fmt.Println("TCP connection close error:", err)
		return
	}
}

// Sender (Client)
func runSender(addr string, meta FileMetadata, stream Stream) {
	if meta.Transport == QUIC_S {
		runQuicSender(addr, meta, stream)
	} else if meta.Transport == TCP_S {
		runTcpSender(addr, meta, stream)
	}
}
