package main

const (
	TCP_S  = "TCP"
	QUIC_S = "QUIC"
)

// FileMetadata holds metadata about the file to transfer
type FileMetadata struct {
	Uuid         string                 `json:"uuid"` // Unique identifier for the file transfer
	Filename     string                 `json:"filename"`
	FlowName     string                 `json:"flow_name"`
	Size         int64                  `json:"size"`
	ChunkSize    int                    `json:"chunk_size"`
	Availability map[int]bool           `json:"availability"` // Map of chunk index to availability. Not all chunks may be available
	Metadata     map[string]interface{} `json:"metadata"`
	Transport    string                 `json:"transport"`           // "TCP" or "QUIC"
	QuicAddr     string                 `json:"quic_addr,omitempty"` // Only set if using QUIC
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

// ProgressManager tracks progress for multiple file transfers by UUID
// Used by the receiver/listener to support resumable/partial transfers
type ProgressManager struct {
	progressMap map[string]*Progress
}

func NewProgressManager() *ProgressManager {
	return &ProgressManager{progressMap: make(map[string]*Progress)}
}

func (pm *ProgressManager) GetOrCreate(uuid string, totalChunks int) *Progress {
	if prog, ok := pm.progressMap[uuid]; ok {
		return prog
	}
	prog := &Progress{
		TotalChunks: totalChunks,
		Received:    make(map[int]bool),
		Sent:        make(map[int]bool),
	}
	pm.progressMap[uuid] = prog
	return prog
}

func (pm *ProgressManager) Remove(uuid string) {
	delete(pm.progressMap, uuid)
}
