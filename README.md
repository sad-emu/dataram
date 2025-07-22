# Data Transport Go Project

This project is a Go application designed to transport data using a flexible interface-based architecture.
The goal is to ram data through a network as quickly as possible.

## Getting Started

1. Ensure you have Go installed (https://golang.org/dl/)
2. Run `go mod tidy` to ensure dependencies are up to date
3. Build and run the application as needed

## Project Structure
- `ramio` — Connection interfaces and logic e.g. TCP, QUIC, TCP-TLS is TODO
- `ramformats` — Objects and formats used for transport
- `ramcore` - TODO
- `tools` - Misc scripts. e.g. keygen.sh will generate TLS certs for QUIC/TCP-TLS

## Extending
Add new listener or sender types by implementing the respective interfaces in the `ramio` package.
