package main

import (
	"data_ram/ramcore"
)

func main() {
	// Entry point for the data transport application
	// Example: Load config, create core, start listener/sender
	_ = ramcore.NewCore(ramcore.Config{})

}
