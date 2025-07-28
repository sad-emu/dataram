package ramformats

import (
	"fmt"
	"os"
	"testing"
)

func TestRamImportBundle_RoundTrip(t *testing.T) {
	// Create test files
	fileCount := 3
	fileSize := 512
	files := make([]string, fileCount)
	fileData := make([][]byte, fileCount)
	for i := 0; i < fileCount; i++ {
		filename := "test_data/import_bundle_file_" + "A" + fmt.Sprintf("%d", i) + ".bin"
		data := make([]byte, fileSize)
		for j := range data {
			data[j] = byte(i + j)
		}
		if err := os.WriteFile(filename, data, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		files[i] = filename
		fileData[i] = data
		defer os.Remove(filename)
	}

	// Export: create bundle
	chunkSize := int64(256)
	maxBundleCount := 2
	maxQueueSize := 5
	exp := NewRamExportBundle(chunkSize, maxBundleCount, maxQueueSize)
	for i := 0; i < fileCount; i++ {
		rf := NewRamFileFromLocal(files[i], files[i])
		if rf == nil {
			t.Fatalf("Failed to create RamFile for %s", files[i])
		}
		if err := exp.PushFile(*rf); err != nil {
			t.Fatalf("Failed to add file to export bundle: %v", err)
		}
	}

	bundles := make([][]byte, 0)
	for {
		bundle, err := exp.GetNextExportBundle()
		if err != nil {
			break
		}
		if bundle == nil {
			break
		}
		bundles = append(bundles, bundle)
	}

	// Import: reconstruct files from bundle
	imp := NewRamImportBundle(10, "test_data/")
	for _, bundle := range bundles {
		if err := imp.ProcessNextExportBundle(bundle); err != nil {
			t.Fatalf("Failed to push bundle to import: %v", err)
		}
	}

	// TODO test for status?

	// if err := imp.ReconstructFiles("test_data/imported"); err != nil {
	// 	t.Fatalf("Failed to reconstruct files: %v", err)
	// }

	// Verify reconstructed files
	filesChecked := 0

	for {
		importedFile := imp.PopFile()
		if importedFile == nil {
			if filesChecked < fileCount {
				t.Errorf("Not all files were processed, expected %d, got %d", fileCount, filesChecked)
			}
			break // No more files to process
		} else {
			filesChecked++
		}
		importedData, err1 := os.ReadFile(importedFile.LocalPath)
		originalData, err2 := os.ReadFile(importedFile.MetaData[DRFileNameKey])
		if err1 != nil {
			t.Fatalf("Failed to read imported file: %v", err1)
		}
		if err2 != nil {
			t.Fatalf("Failed to read original file: %v", err2)
		}
		if len(importedData) != fileSize {
			t.Errorf("Imported file size mismatch: got %d, want %d", len(importedData), fileSize)
		}
		for j := range importedData {
			if importedData[j] != originalData[j] {
				t.Errorf("Data mismatch in file %s at byte %d", importedFile, j)
				break
			}
		}
		os.Remove(importedFile.LocalPath)
	}
}
