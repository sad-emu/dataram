package main

const (
	TCP_S  = "TCP"
	QUIC_S = "QUIC"
)

// FileMetadata holds metadata about the file to transfer
type FileMetadata struct {
	Filename  string                 `json:"filename"`
	FlowName  string                 `json:"flow_name"`
	Size      int64                  `json:"size"`
	ChunkSize int                    `json:"chunk_size"`
	Metadata  map[string]interface{} `json:"metadata"`
	Transport string                 `json:"transport"`           // "TCP" or "QUIC"
	QuicAddr  string                 `json:"quic_addr,omitempty"` // Only set if using QUIC
}

// Progress tracks transfer progress
type Progress struct {
	TotalChunks int
	Received    map[int]bool
	Sent        map[int]bool
}

type ChunkResult struct {
	Index int
	Data  []byte
	Ok    bool
}

// ChunkResultList represents a variable list of ChunkResult
// Useful for passing or collecting multiple chunk results
// e.g. for batch operations or result aggregation
type ChunkResultList []ChunkResult
