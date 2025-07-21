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

func NewRamImportBundle(chunkSize int64, maxBundleCount int, maxQueueSize int, processingDir string) *RamImportBundle {
	return &RamImportBundle{
		processBundles:      make([]RamFile, 0),
		completedFiles:      make([]RamFile, 0),
		filePartsQueue:      make([]byte, 0),
		chunkSize:           chunkSize,
		maxBundleCount:      maxBundleCount,
		maxQueueSize:        maxQueueSize,
		processingDirectory: processingDir,
	}
}

func (rb *RamImportBundle) PopFile(rf RamFile) *RamFile {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if len(rb.completedFiles) != 0 {
		rf, _ := PopFront(&rb.completedFiles)
		return &rf
	} else {
		return nil
	}
}

func (rb *RamImportBundle) ProcessNextExportBundle(dataIn []byte) error {

	// Verify data has a valid header

	// if it's a metadata bundle
	// create RamFiles and add into process bundles
	// Check to see if they exist first with uuid checks, if they exist just update metadata

	// if it's a data bundle
	// Create ramfiles and add into process bundles
	// Check to see if they exist first with uuid checks
	// Write the data as specified in the data metadata bundle

	// If files are completed move them to the completedfiles list

}
