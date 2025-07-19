package ramformats

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"
)

// TODO this

func createTestFile(filename string, size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(filename, data, 0644)
	return data, err
}

func TestRamBundle_GetNextBundle(t *testing.T) {
	// Create 5 test files with random data
	fileCount := 5
	fileSize := 1024
	files := make([]string, fileCount)
	fileData := make([][]byte, fileCount)
	for i := 0; i < fileCount; i++ {
		filename := fmt.Sprintf("test_data/test_bundle_file_%d.bin", i)
		data, err := createTestFile(filename, fileSize)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		files[i] = filename
		fileData[i] = data
		defer os.Remove(filename)
	}

	// Create a new RamBundle
	chunkSize := int64(512)
	maxBundleCount := 4
	maxQueueSize := 10
	rb := NewRamBundle(chunkSize, maxBundleCount, maxQueueSize)

	// Add files to the bundle
	for i := 0; i < fileCount; i++ {
		rf := NewRamFileFromLocal(files[i], files[i])
		if rf == nil {
			t.Fatalf("Failed to create RamFile for %s", files[i])
		}
		if err := rb.AddFile(*rf); err != nil {
			t.Fatalf("Failed to add file to bundle: %v", err)
		}
	}

	// Call GetNextBundle until all bundles are sent
	bundleDatas := make([][]byte, 0)
	for {
		bundle, err := rb.GetNextBundle()
		if err != nil {
			// Only allow "GetNextBundle not implemented yet" error for now
			if err.Error() == "GetNextBundle not implemented yet" {
				break
			}
			t.Fatalf("GetNextBundle error: %v", err)
		}
		if bundle == nil {
			break
		}
		bundleDatas = append(bundleDatas, bundle)
	}

	// Verify bundle format and data
	if len(bundleDatas) == 0 {
		t.Fatal("No bundles were returned")
	}

	// Check metadata header and data header
	foundMeta := false
	foundData := false
	for _, bundle := range bundleDatas {
		if len(bundle) < 4 {
			t.Errorf("Bundle too short to contain header")
			continue
		}
		header := BytesToInt(bundle[0:4])
		if header == METADATA_HEADER {
			foundMeta = true
			// Optionally, check metadata format here
		} else if header == DATA_HEADER {
			foundData = true
			// Optionally, check data format here
		} else {
			t.Errorf("Unknown header in bundle: %d", header)
		}
	}

	if !foundMeta {
		t.Error("No metadata bundle found")
	}
	if !foundData {
		t.Error("No data bundle found")
	}
}
