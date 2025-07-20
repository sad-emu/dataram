package ramformats

import (
	"sync"
)

// Define const header bytes

// The purpose of RamBundle is to take chunks of ramexportbundles
// and piece them back into ramfiles

// The bundles have two types of output
// 1. A Metadata map that contains information about the files in the bundle
// 2. A data block that contains the actual data of the files in the bundle
// Multiple data blocks may result from a single bundle. Only a single metadata map will be created for each set of bundles.

// State will need to be saved and stored for this class for restarts

type RamImportBundle struct {
	processingDirectory string
	filePartsQueue      []byte
	processBundles      []RamFile
	completedFiles      []RamFile

	maxQueueSize   int // Maximum size of the queue
	chunkSize      int64
	maxBundleCount int
	mu             sync.Mutex // Mutex to protect concurrent access
}
