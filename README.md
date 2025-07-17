# Data Transport Go Project

This project is a Go application designed to transport data using a flexible interface-based architecture. It includes:
- `ramio` package: `Listener` interface (with initial `TCPListener` implementation, extensible to SFTP and HTTP), `Sender` interface (with initial `TCPSender` implementation, extensible to more types)
- `ramcore` package: `Core` struct to coordinate the application, `Config` struct for configuration

## Getting Started

1. Ensure you have Go installed (https://golang.org/dl/)
2. Run `go mod tidy` to ensure dependencies are up to date
3. Build and run the application as needed

## Project Structure
- `ramio/listener.go` — Listener interface and implementations
- `ramio/sender.go` — Sender interface and implementations
- `ramcore/core.go` — Core struct
- `ramcore/config.go` — Config struct

## Extending
Add new listener or sender types by implementing the respective interfaces in the `ramio` package.
