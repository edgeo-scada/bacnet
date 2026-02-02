# BACnet/IP Driver

A comprehensive BACnet/IP client library and CLI tool for building automation systems.

## Features

- **Device Discovery**: Discover BACnet devices using Who-Is/I-Am
- **Property Read/Write**: Read and write object properties
- **COV Subscriptions**: Subscribe to Change of Value notifications
- **Multiple Output Formats**: Table, JSON, CSV, and raw output
- **Interactive Mode**: REPL for exploring BACnet devices
- **Cross-Platform**: Builds for Windows, Linux, and macOS

## Installation

### From Source

```bash
# Clone the repository
cd /path/to/edgeo/drivers/bacnet

# Build for current platform
make build

# Or build for all platforms
make build-all
```

### Binaries

Pre-built binaries are available in the `bin/` directory:
- `edgeo-bacnet-darwin-amd64` - macOS Intel
- `edgeo-bacnet-darwin-arm64` - macOS Apple Silicon
- `edgeo-bacnet-linux-amd64` - Linux x64
- `edgeo-bacnet-linux-arm64` - Linux ARM64
- `edgeo-bacnet-windows-amd64.exe` - Windows x64

## Quick Start

### Discover Devices

```bash
# Discover all devices on the network
edgeo-bacnet scan

# Discover devices with specific instance IDs
edgeo-bacnet scan --low 1 --high 100

# Discover with extended timeout
edgeo-bacnet scan --scan-timeout 10s
```

### Read Properties

```bash
# Read present value from analog input 1
edgeo-bacnet read -d 1234 -o analog-input:1 -p present-value

# Read using short names
edgeo-bacnet read -d 1234 -o ai:1 -p pv

# Read object name
edgeo-bacnet read -d 1234 -o device:1234 -p object-name

# Read with JSON output
edgeo-bacnet read -d 1234 -o ai:1 -p pv -o json
```

### Write Properties

```bash
# Write present value to analog output
edgeo-bacnet write -d 1234 -o analog-output:1 -p present-value -V 75.5

# Write with priority
edgeo-bacnet write -d 1234 -o binary-output:1 -p present-value -V true --priority 8

# Release a priority (write null)
edgeo-bacnet write -d 1234 -o analog-output:1 -p present-value -V null --priority 8
```

### Watch for Changes

```bash
# Poll present value every second
edgeo-bacnet watch -d 1234 -o analog-input:1 -p present-value --interval 1s

# Subscribe to COV notifications
edgeo-bacnet watch -d 1234 -o analog-input:1 --cov
```

### Device Information

```bash
# Get device info
edgeo-bacnet info -d 1234

# Dump all objects and properties
edgeo-bacnet dump -d 1234 -f backup.json -o json
```

### Interactive Mode

```bash
edgeo-bacnet interactive

bacnet> scan
bacnet> use 1234
bacnet[1234]> list
bacnet[1234]> read ai:1 pv
bacnet[1234]> write ao:1 pv 75.5
bacnet[1234]> exit
```

## Object Types

| Type | Short | ID |
|------|-------|-----|
| analog-input | ai | 0 |
| analog-output | ao | 1 |
| analog-value | av | 2 |
| binary-input | bi | 3 |
| binary-output | bo | 4 |
| binary-value | bv | 5 |
| device | dev | 8 |
| multi-state-input | msi | 13 |
| multi-state-output | mso | 14 |
| multi-state-value | msv | 19 |
| schedule | sch | 17 |
| trend-log | tl | 20 |

## Property Identifiers

| Property | Short | ID |
|----------|-------|-----|
| present-value | pv | 85 |
| object-name | name | 77 |
| description | desc | 28 |
| status-flags | sf | 111 |
| out-of-service | oos | 81 |
| units | - | 117 |
| priority-array | pa | 87 |
| relinquish-default | rd | 104 |

## Configuration

### Configuration File

Create `~/.edgeo-bacnet.yaml`:

```yaml
# Default device
device: 1234

# Network settings
timeout: 3s
retries: 3

# Output format
output: table

# BBMD settings (for foreign device registration)
bbmd: 192.168.1.1
bbmd-port: 47808
bbmd-ttl: 60s
```

### Environment Variables

All flags can be set via environment variables with the `BACNET_` prefix:

```bash
export BACNET_DEVICE=1234
export BACNET_TIMEOUT=5s
export BACNET_OUTPUT=json
```

## Library Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/edgeo/drivers/bacnet/bacnet"
)

func main() {
    // Create client
    client, err := bacnet.NewClient(
        bacnet.WithTimeout(3 * time.Second),
        bacnet.WithRetries(3),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Connect
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Discover devices
    devices, err := client.WhoIs(ctx, bacnet.WithDiscoveryTimeout(5*time.Second))
    if err != nil {
        log.Fatal(err)
    }

    for _, dev := range devices {
        fmt.Printf("Found device: %d\n", dev.ObjectID.Instance)
    }

    // Read property
    value, err := client.ReadProperty(ctx, 1234,
        bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
        bacnet.PropertyPresentValue,
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Value: %v\n", value)

    // Write property
    err = client.WriteProperty(ctx, 1234,
        bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogOutput, 1),
        bacnet.PropertyPresentValue,
        75.5,
        bacnet.WithPriority(8),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

## Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint
```

## Project Structure

```
bacnet/
├── bacnet/                    # Library package
│   ├── client.go              # Main client implementation
│   ├── types.go               # BACnet types and constants
│   ├── options.go             # Functional options
│   ├── errors.go              # Error types
│   ├── protocol.go            # Protocol encoding/decoding
│   ├── metrics.go             # Metrics collection
│   └── internal/
│       └── transport/
│           └── udp.go         # UDP transport
├── cmd/
│   └── edgeo-bacnet/          # CLI application
│       ├── main.go
│       ├── root.go
│       ├── scan.go
│       ├── read.go
│       ├── write.go
│       ├── watch.go
│       ├── dump.go
│       ├── info.go
│       ├── interactive.go
│       └── output.go
├── bin/                       # Built binaries
├── go.mod
├── go.work
├── Makefile
└── README.md
```

## License

MIT License - see LICENSE file for details.
