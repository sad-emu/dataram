package ramformats

import (
	"os"
	"testing"
)

func TestNewRamFileFromLocal(t *testing.T) {
	// Create a temporary file
	os.MkdirAll("test_data", 0755)
	tmpFile := "test_data/test_localstream_len.txt"
	defer os.Remove(tmpFile)

	content := []byte("hello world")
	err := os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	rf := NewRamFileFromLocal(tmpFile, "testfile.txt")
	if rf.LocalPath != tmpFile {
		t.Errorf("LocalPath mismatch: got %s, want %s", rf.LocalPath, tmpFile)
	}
	if rf.MetaData[DRFileNameKey] != "testfile.txt" {
		t.Errorf("filename metadata mismatch: got %s", rf.MetaData[DRFileNameKey])
	}
	if rf.MetaData[DRFileSizeKey] != "11" {
		t.Errorf("fileSize metadata mismatch: got %s, want 11", rf.MetaData[DRFileSizeKey])
	}
	if rf.UUID == "" {
		t.Error("UUID should not be empty")
	}
}

func TestNewRamFileFromMeta(t *testing.T) {
	meta := map[string]string{
		DRFileNameKey: "meta.txt",
		DRFileSizeKey: "123",
		DRUUIDKey:     "test-uuid",
	}
	rf := NewRamFileFromMeta(meta)
	if rf.MetaData[DRFileNameKey] != "meta.txt" {
		t.Errorf("filename metadata mismatch: got %s", rf.MetaData[DRFileNameKey])
	}
	if rf.MetaData[DRFileSizeKey] != "123" {
		t.Errorf("fileSize metadata mismatch: got %s", rf.MetaData[DRFileSizeKey])
	}
	if rf.UUID != "test-uuid" {
		t.Errorf("UUID mismatch: got %s, want test-uuid", rf.UUID)
	}
	if rf.RamFileType != DROutputFile {
		t.Errorf("RamFileType mismatch: got %s, want %s", rf.RamFileType, DROutputFile)
	}
}

func TestRamFileMetaRoundTrip(t *testing.T) {
	// Create a file and RamFile
	os.MkdirAll("test_data", 0755)
	tmpFile := "test_data/test_localstream_len.txt"
	defer os.Remove(tmpFile)
	content := []byte("roundtrip content")
	err := os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	rf1 := NewRamFileFromLocal(tmpFile, "roundtrip.txt")
	meta := rf1.MetaData

	rf2 := NewRamFileFromMeta(meta)
	if rf2.MetaData[DRFileNameKey] != "roundtrip.txt" {
		t.Errorf("filename metadata mismatch: got %s", rf2.MetaData[DRFileNameKey])
	}
	if rf2.MetaData[DRFileSizeKey] != "17" {
		t.Errorf("fileSize metadata mismatch: got %s, want 17", rf2.MetaData[DRFileSizeKey])
	}
	if rf2.UUID != rf1.UUID {
		t.Errorf("UUID mismatch: got %s, want %s", rf2.UUID, rf1.UUID)
	}
}

func TestNewRamFileNonExistFromLocal(t *testing.T) {
	// Create a temporary file
	ramFile := NewRamFileFromLocal("non_existent_file.txt", "nonexistent.txt")
	if ramFile != nil {
		t.Error("Expected nil RamFile for non-existent file, got non-nil")
	}
}
