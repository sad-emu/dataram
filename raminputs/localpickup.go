// TOOD make interface that this implements
package raminputs

import (
	"data_ram/ramformats"
	"fmt"
	"os"
	"regexp"
)

type LocalPickup struct {
	PickupPath      string
	PickupRegex     string
	IgnoreDotFiles  bool
	FilesInProgress map[string]ramformats.RamFile
	FilesInQueue    map[string]ramformats.RamFile
}

func NewLocalPickup(pickupPath, pickupRegex string, ignoreDotFiles bool) *LocalPickup {
	return &LocalPickup{
		PickupPath:      pickupPath,
		PickupRegex:     pickupRegex,
		IgnoreDotFiles:  ignoreDotFiles,
		FilesInProgress: make(map[string]ramformats.RamFile),
		FilesInQueue:    make(map[string]ramformats.RamFile),
	}
}

func (lp *LocalPickup) Init() error {
	// Check to see if pickup path exists
	if lp.PickupPath == "" {
		return fmt.Errorf("Pickup path is not set")
	}
	// Check to see if pickup path is a directory
	info, err := os.Stat(lp.PickupPath)
	if err != nil {
		return fmt.Errorf("Error accessing pickup path: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("Pickup path is not a directory")
	}
	// Check if the directory is writeable using a random uuid filename
	testFileName := lp.PickupPath + "/rampickuptest_" + ramformats.GenerateUUID()
	if err := os.WriteFile(lp.PickupPath+"/"+testFileName, []byte("test"), 0644); err != nil {
		return fmt.Errorf("Pickup path is not writeable check file failed: %v", err)
	}
	// Remove the test file
	if err := os.Remove(lp.PickupPath + lp.PickupPath + "/" + testFileName); err != nil {
		return fmt.Errorf("Error removing check file from pickup path: %v", err)
	}
	// Check regex is valid or empty
	if lp.PickupRegex != "" {
		if _, err := regexp.Compile(lp.PickupRegex); err != nil {
			return fmt.Errorf("Pickup regex is not valid: %v", err)
		}
	}
	return nil
}

func (lp *LocalPickup) Pulse() error {

	return nil
}

func (lp *LocalPickup) GetFile() (*ramformats.RamFile, error) {
	if len(lp.FilesInQueue) > 0 {
		// Pop the first file from the queue
		for uuid, rf := range lp.FilesInQueue {
			delete(lp.FilesInQueue, rf.UUID)
			lp.FilesInProgress[uuid] = rf
			return &rf, nil
		}
	}
	// Return error if no files are available
	return nil, fmt.Errorf("No files available in queue")
}

func (lp *LocalPickup) ReadData(rf *ramformats.RamFile, len int) ([]byte, error) {
	// Check the ramfile is in the in-progress map
	if _, exists := lp.FilesInProgress[rf.UUID]; !exists {
		return nil, fmt.Errorf("RamFile %s is not in progress", rf.UUID)
	}
	// Read the data from the local file to the len
	data, err := os.ReadFile(rf.LocalPath)
	if err != nil {
	return nil, nil
}
