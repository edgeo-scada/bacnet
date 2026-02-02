# BACnet/IP Driver

A comprehensive BACnet/IP client implementation for Go, designed for building automation and control systems.

**Version:** 1.0.0

## Features

- **BACnet/IP Protocol**: Full UDP-based BACnet/IP implementation with BVLC support
- **Device Discovery**: Who-Is/I-Am broadcast mechanism with range filtering
- **Property Operations**: ReadProperty, WriteProperty, ReadPropertyMultiple
- **COV Subscriptions**: Change of Value notifications with confirmed/unconfirmed modes
- **BBMD Support**: Foreign device registration for cross-subnet communication
- **Metrics Collection**: Comprehensive metrics for monitoring and observability
- **CLI Tool**: Full-featured command-line interface for device interaction
- **Structured Logging**: Integration with Go's `slog` package

## Installation

```bash
go get github.com/edgeo/drivers/bacnet
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/edgeo/drivers/bacnet/bacnet"
)

func main() {
    // Create client
    client, err := bacnet.NewClient(
        bacnet.WithTimeout(3 * time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Connect to network
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Discover devices
    devices, err := client.WhoIs(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d devices\n", len(devices))

    // Read a property
    if len(devices) > 0 {
        deviceID := devices[0].ObjectID.Instance
        value, err := client.ReadProperty(ctx, deviceID,
            bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
            bacnet.PropertyPresentValue,
        )
        if err != nil {
            log.Printf("Read error: %v", err)
        } else {
            fmt.Printf("Value: %v\n", value)
        }
    }
}
```

## Documentation

- [Getting Started](getting-started.md) - Installation and first steps
- [Client API](client.md) - Client operations reference
- [Configuration Options](options.md) - All configuration options
- [Device Discovery](discovery.md) - Who-Is/I-Am and device management
- [COV Subscriptions](cov.md) - Change of Value notifications
- [Error Handling](errors.md) - Error types and handling patterns
- [Metrics](metrics.md) - Metrics collection and monitoring
- [CLI Reference](cli.md) - Command-line tool documentation
- [Changelog](changelog.md) - Version history

## Examples

- [Basic Client](examples/basic-client.md) - Complete client usage example
- [Device Discovery](examples/discovery.md) - Comprehensive discovery example

## Protocol Support

### Services Implemented

| Service | Type | Description |
|---------|------|-------------|
| Who-Is | Unconfirmed | Device discovery broadcast |
| I-Am | Unconfirmed | Device announcement |
| ReadProperty | Confirmed | Read single property |
| WriteProperty | Confirmed | Write single property |
| ReadPropertyMultiple | Confirmed | Read multiple properties |
| SubscribeCOV | Confirmed | COV subscription |
| UnconfirmedCOVNotification | Unconfirmed | COV notification receipt |

### Object Types

The driver supports all standard BACnet object types including:

- Analog Input/Output/Value
- Binary Input/Output/Value
- Multi-State Input/Output/Value
- Device, Schedule, Calendar
- Trend Log, Notification Class
- And more (60+ object types)

### Properties

Common properties with shorthand aliases:

| Property | Alias | ID |
|----------|-------|-----|
| present-value | pv | 85 |
| object-name | name | 77 |
| out-of-service | oos | 81 |
| priority-array | pa | 87 |
| relinquish-default | rd | 104 |
| units | - | 117 |
| status-flags | sf | 111 |

## Requirements

- Go 1.21 or later
- Network access to BACnet/IP devices (UDP port 47808)

## License

MIT License
