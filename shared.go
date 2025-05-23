package main

// FileMetadata holds metadata about the file to transfer
type FileMetadata struct {
	Filename  string                 `json:"filename"`
	FlowName  string                 `json:"flow_name"`
	Size      int64                  `json:"size"`
	ChunkSize int                    `json:"chunk_size"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Progress tracks transfer progress
type Progress struct {
	TotalChunks int
	Received    map[int]bool
	Sent        map[int]bool
}
