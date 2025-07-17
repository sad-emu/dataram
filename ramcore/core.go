package ramcore

import (
	"data_ram/ramio"
)

// Core coordinates listeners and senders using the config.
type Core struct {
	Config   Config
	Listener ramio.Listener
	Sender   ramio.Sender
}

func NewCore(cfg Config) *Core {
	// TODO: Instantiate Listener and Sender based on config
	return &Core{Config: cfg}
}

// ...existing code...
