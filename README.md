# BACnet/IP Client Library

A pure Go BACnet/IP client library for building automation and control systems.

## Features

- **BACnet/IP Protocol**: Full support for BACnet/IP over UDP with BVLC
- **Device Discovery**: Who-Is/I-Am broadcasts for automatic device discovery
- **Property Services**: ReadProperty, WriteProperty, ReadPropertyMultiple
- **COV Subscriptions**: Subscribe to Change of Value notifications
- **Foreign Device Registration**: BBMD support for cross-subnet communication
- **Priority Writing**: Full support for BACnet priority array (1-16)
- **Metrics**: Built-in metrics for monitoring request statistics
- **Functional Options**: Clean, extensible configuration pattern

## Installation

```bash
go get github.com/edgeo-scada/bacnet
```

## Quick Start

### Device Discovery

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/edgeo-scada/bacnet/bacnet"
)

func main() {
    client, err := bacnet.NewClient(
        bacnet.WithTimeout(3 * time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
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
        fmt.Printf("Device %d at %v\n", dev.ObjectID.Instance, dev.Address)
    }
}
```

### Read Property

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/edgeo-scada/bacnet/bacnet"
)

func main() {
    client, err := bacnet.NewClient(
        bacnet.WithTimeout(3 * time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Read present value from analog input 1 on device 1234
    value, err := client.ReadProperty(ctx, 1234,
        bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
        bacnet.PropertyPresentValue,
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Value: %v\n", value)
}
```

### Write Property

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/edgeo-scada/bacnet/bacnet"
)

func main() {
    client, err := bacnet.NewClient(
        bacnet.WithTimeout(3 * time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Write 75.5 to analog output 1 at priority 8
    err = client.WriteProperty(ctx, 1234,
        bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogOutput, 1),
        bacnet.PropertyPresentValue,
        float32(75.5),
        bacnet.WithPriority(8),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

## Configuration Options

### Client Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithDeviceID(id)` | Local device ID | 0xFFFFFFFF |
| `WithLocalAddress(addr)` | Local address to bind to | Auto |
| `WithNetworkNumber(net)` | BACnet network number | 0 |
| `WithTimeout(duration)` | Request timeout | 3s |
| `WithRetries(n)` | Number of retries | 3 |
| `WithRetryDelay(duration)` | Delay between retries | 500ms |
| `WithLogger(logger)` | Custom slog logger | slog.Default() |

### BBMD Options

| Option | Description |
|--------|-------------|
| `WithBBMD(addr, port, ttl)` | Register as foreign device with BBMD |

### APDU Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithMaxAPDULength(length)` | Maximum APDU length | 1476 |
| `WithSegmentation(seg)` | Segmentation capability | None |
| `WithProposedWindowSize(size)` | Segmentation window size | 1 |

### Discovery Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithDeviceRange(low, high)` | Device instance range for Who-Is | All |
| `WithDiscoveryTimeout(duration)` | Discovery timeout | 5s |
| `WithTargetNetwork(net)` | Target network for discovery | Local |

### Read Options

| Option | Description |
|--------|-------------|
| `WithArrayIndex(index)` | Read specific array element |

### Write Options

| Option | Description |
|--------|-------------|
| `WithPriority(priority)` | Write priority (1-16) |
| `WithWriteArrayIndex(index)` | Write to specific array element |

### COV Subscription Options

| Option | Description |
|--------|-------------|
| `WithSubscriptionLifetime(seconds)` | Subscription lifetime |
| `WithCOVIncrement(increment)` | COV increment for analog values |
| `WithConfirmedNotifications(bool)` | Request confirmed notifications |

## COV Subscriptions

```go
// Subscribe to COV notifications
handler := func(deviceID uint32, objectID bacnet.ObjectIdentifier, values []bacnet.PropertyValue) {
    for _, pv := range values {
        fmt.Printf("COV: %s.%s = %v\n", objectID, pv.PropertyID, pv.Value)
    }
}

subID, err := client.SubscribeCOV(ctx, 1234,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
    handler,
    bacnet.WithSubscriptionLifetime(300),
)
if err != nil {
    log.Fatal(err)
}

// Later, unsubscribe
err = client.UnsubscribeCOV(ctx, 1234,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
    subID,
)
```

## CLI Tool (edgeo-bacnet)

A complete BACnet/IP command-line client for testing, debugging, and monitoring.

### Installation

```bash
go build -o edgeo-bacnet ./cmd/edgeo-bacnet
```

### Commands

| Command | Description |
|---------|-------------|
| `scan` | Discover BACnet devices on the network |
| `read` | Read a property from an object |
| `write` | Write a property to an object |
| `watch` | Monitor a property for changes |
| `dump` | Dump all objects and properties from a device |
| `info` | Display device information |
| `interactive` | Interactive REPL shell |
| `version` | Print version information |

### Global Flags

```
-H, --host string        Target device IP address
-p, --port int           BACnet/IP port (default 47808)
-d, --device uint32      Target device instance ID
-t, --timeout duration   Request timeout (default 3s)
    --retries int        Number of retries (default 3)
-o, --output string      Output format: table, json, csv, raw (default "table")
-v, --verbose            Verbose output
    --local string       Local address to bind to
    --bbmd string        BBMD address for foreign device registration
    --bbmd-port int      BBMD port (default 47808)
    --bbmd-ttl duration  BBMD registration TTL (default 60s)
    --config string      Config file (default ~/.edgeo-bacnet.yaml)
```

### Scan Examples

```bash
# Discover all devices
edgeo-bacnet scan

# Discover devices with instance IDs 1-100
edgeo-bacnet scan --low 1 --high 100

# Discover with extended timeout
edgeo-bacnet scan --scan-timeout 10s

# Output as JSON
edgeo-bacnet scan -o json
```

### Read Examples

```bash
# Read present value from analog input 1
edgeo-bacnet read -d 1234 -O analog-input:1 -P present-value

# Read using short names
edgeo-bacnet read -d 1234 -O ai:1 -P pv

# Read object name
edgeo-bacnet read -d 1234 -O device:1234 -P object-name

# Read array element
edgeo-bacnet read -d 1234 -O device:1234 -P object-list --index 1

# JSON output
edgeo-bacnet read -d 1234 -O ai:1 -P pv -o json
```

### Write Examples

```bash
# Write present value to analog output
edgeo-bacnet write -d 1234 -O analog-output:1 -P present-value -V 75.5

# Write with priority
edgeo-bacnet write -d 1234 -O binary-output:1 -P present-value -V true --priority 8

# Release a priority (write null)
edgeo-bacnet write -d 1234 -O analog-output:1 -P present-value -V null --priority 8

# Write object name
edgeo-bacnet write -d 1234 -O analog-value:1 -P object-name -V "Temperature Setpoint"
```

### Watch Examples

```bash
# Poll present value every second
edgeo-bacnet watch -d 1234 -O analog-input:1 -P present-value --interval 1s

# Subscribe to COV notifications
edgeo-bacnet watch -d 1234 -O analog-input:1 --cov

# COV with custom lifetime
edgeo-bacnet watch -d 1234 -O analog-input:1 --cov --cov-lifetime 300

# Log to CSV
edgeo-bacnet watch -d 1234 -O ai:1 -o csv > log.csv
```

### Dump Examples

```bash
# Dump all objects to stdout
edgeo-bacnet dump -d 1234

# Dump to a JSON file
edgeo-bacnet dump -d 1234 -f device_backup.json -o json

# Dump specific object types
edgeo-bacnet dump -d 1234 --objects analog-input,analog-output

# Dump specific properties
edgeo-bacnet dump -d 1234 --props present-value,object-name,description
```

### Interactive Mode

```bash
edgeo-bacnet interactive
```

Commands in interactive mode:
- `scan` - Discover devices
- `use <device-id>` - Select a device
- `list` - List objects on current device
- `read <object> <property>` - Read a property
- `write <object> <property> <value>` - Write a property
- `info` - Show device info
- `metrics` - Show client metrics
- `help` - Show help
- `exit` - Exit interactive mode

### Configuration File

Create `~/.edgeo-bacnet.yaml`:

```yaml
device: 1234
timeout: 3s
retries: 3
output: table
bbmd: 192.168.1.1
bbmd-port: 47808
bbmd-ttl: 60s
```

## Metrics

```go
metrics := client.Metrics()
snapshot := metrics.Snapshot()

fmt.Printf("Requests sent: %d\n", snapshot.RequestsSent)
fmt.Printf("Requests succeeded: %d\n", snapshot.RequestsSucceeded)
fmt.Printf("Requests failed: %d\n", snapshot.RequestsFailed)
fmt.Printf("Devices discovered: %d\n", snapshot.DevicesDiscovered)
fmt.Printf("Avg latency: %v\n", snapshot.LatencyStats.Avg)
fmt.Printf("Uptime: %v\n", snapshot.Uptime)
```

## API Reference

### Client Methods

| Method | Description |
|--------|-------------|
| `Connect(ctx)` | Open BACnet client connection |
| `Close()` | Close the connection |
| `State()` | Get connection state |
| `WhoIs(ctx, opts...)` | Discover devices |
| `GetDevice(deviceID)` | Get discovered device info |
| `ReadProperty(ctx, deviceID, objectID, propertyID, opts...)` | Read a property |
| `WriteProperty(ctx, deviceID, objectID, propertyID, value, opts...)` | Write a property |
| `ReadPropertyMultiple(ctx, deviceID, requests)` | Read multiple properties |
| `SubscribeCOV(ctx, deviceID, objectID, handler, opts...)` | Subscribe to COV |
| `UnsubscribeCOV(ctx, deviceID, objectID, subID)` | Unsubscribe from COV |
| `GetObjectList(ctx, deviceID)` | Get list of objects from device |
| `Metrics()` | Get metrics |

### Object Types

| Type | Short | ID |
|------|-------|-----|
| analog-input | ai | 0 |
| analog-output | ao | 1 |
| analog-value | av | 2 |
| binary-input | bi | 3 |
| binary-output | bo | 4 |
| binary-value | bv | 5 |
| calendar | cal | 6 |
| device | dev | 8 |
| file | - | 10 |
| loop | - | 12 |
| multi-state-input | msi | 13 |
| multi-state-output | mso | 14 |
| notification-class | nc | 15 |
| program | prg | 16 |
| schedule | sch | 17 |
| multi-state-value | msv | 19 |
| trend-log | tl | 20 |

### Property Identifiers

| Property | Short | ID |
|----------|-------|-----|
| object-identifier | oid | 75 |
| object-name | name | 77 |
| object-type | type | 79 |
| present-value | pv | 85 |
| description | desc | 28 |
| status-flags | sf | 111 |
| event-state | - | 36 |
| reliability | - | 103 |
| out-of-service | oos | 81 |
| units | - | 117 |
| priority-array | pa | 87 |
| relinquish-default | rd | 104 |
| cov-increment | - | 22 |
| high-limit | - | 45 |
| low-limit | - | 59 |
| vendor-name | - | 121 |
| vendor-identifier | - | 120 |
| model-name | - | 70 |
| firmware-revision | - | 44 |
| object-list | - | 76 |

### Priority Levels

| Level | Usage |
|-------|-------|
| 1 | Manual-Life Safety |
| 2 | Automatic-Life Safety |
| 3-4 | Reserved |
| 5 | Critical Equipment Control |
| 6 | Minimum On/Off |
| 7 | Reserved |
| 8 | Manual Operator |
| 9-15 | Various automated functions |
| 16 | Default/Relinquish |

### Error Handling

```go
value, err := client.ReadProperty(ctx, deviceID, objectID, propertyID)
if err != nil {
    if bacnet.IsTimeout(err) {
        // Request timed out
    } else if bacnet.IsDeviceNotFound(err) {
        // Device not found
    } else if bacnet.IsPropertyNotFound(err) {
        // Property not found
    } else if bacnet.IsAccessDenied(err) {
        // Read/write access denied
    } else {
        // Other error
        var bacnetErr *bacnet.BACnetError
        if errors.As(err, &bacnetErr) {
            fmt.Printf("BACnet error: class=%s, code=%s\n",
                bacnetErr.Class, bacnetErr.Code)
        }
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
│   ├── examples/
│   │   └── basic/main.go      # Basic usage example
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

MIT License
