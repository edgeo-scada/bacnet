# Getting Started

This guide covers the basics of using the BACnet/IP driver for device communication.

## Prerequisites

- Go 1.21 or later
- Network access to BACnet/IP devices (UDP port 47808)
- Basic understanding of BACnet concepts (objects, properties, device IDs)

## Installation

```bash
go get github.com/edgeo-scada/bacnet
```

## Basic Client Usage

### Creating a Client

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/edgeo-scada/bacnet/bacnet"
)

func main() {
    // Create client with default options
    client, err := bacnet.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    // Or with custom configuration
    client, err = bacnet.NewClient(
        bacnet.WithTimeout(3 * time.Second),
        bacnet.WithRetries(3),
        bacnet.WithRetryDelay(500 * time.Millisecond),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Connect to the BACnet network
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Client is now ready to use
}
```

### Discovering Devices

```go
// Discover all devices on the network
devices, err := client.WhoIs(ctx)
if err != nil {
    log.Fatal(err)
}

for _, dev := range devices {
    fmt.Printf("Device %d at %v (Vendor: %d)\n",
        dev.ObjectID.Instance,
        dev.Address,
        dev.VendorID,
    )
}

// Discover devices in a specific range
devices, err = client.WhoIs(ctx,
    bacnet.WithDeviceRange(1000, 2000),
    bacnet.WithDiscoveryTimeout(5 * time.Second),
)
```

### Reading Properties

```go
deviceID := uint32(1234)

// Read present value from an analog input
value, err := client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
    bacnet.PropertyPresentValue,
)
if err != nil {
    log.Printf("Read failed: %v", err)
} else {
    fmt.Printf("Temperature: %v\n", value)
}

// Read with array index
value, err = client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID),
    bacnet.PropertyObjectList,
    bacnet.WithArrayIndex(0), // Get array length
)
```

### Writing Properties

```go
deviceID := uint32(1234)

// Write a setpoint value
err := client.WriteProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogOutput, 1),
    bacnet.PropertyPresentValue,
    72.5, // Temperature setpoint
)
if err != nil {
    log.Printf("Write failed: %v", err)
}

// Write with priority (1-16, where 1 is highest)
err = client.WriteProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeBinaryOutput, 1),
    bacnet.PropertyPresentValue,
    true,
    bacnet.WithPriority(8), // Manual operator priority
)
```

### Reading Multiple Properties

```go
// Read multiple properties in one request
results, err := client.ReadPropertyMultiple(ctx, deviceID, []bacnet.ReadPropertyRequest{
    {
        ObjectID:   bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
        PropertyID: bacnet.PropertyPresentValue,
    },
    {
        ObjectID:   bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
        PropertyID: bacnet.PropertyUnits,
    },
    {
        ObjectID:   bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 2),
        PropertyID: bacnet.PropertyPresentValue,
    },
})
if err != nil {
    log.Printf("ReadPropertyMultiple failed: %v", err)
}

for _, result := range results {
    fmt.Printf("%s.%s = %v\n",
        result.ObjectID,
        result.PropertyID,
        result.Value,
    )
}
```

## Object Type Aliases

Use shorthand aliases for common object types:

| Alias | Object Type |
|-------|-------------|
| ai | analog-input |
| ao | analog-output |
| av | analog-value |
| bi | binary-input |
| bo | binary-output |
| bv | binary-value |
| msi | multi-state-input |
| mso | multi-state-output |
| msv | multi-state-value |
| dev | device |
| sch | schedule |
| tl | trend-log |

## Property Aliases

Use shorthand aliases for common properties:

| Alias | Property |
|-------|----------|
| pv | present-value |
| name | object-name |
| oos | out-of-service |
| pa | priority-array |
| rd | relinquish-default |
| sf | status-flags |
| desc | description |

## CLI Quick Start

Install the CLI tool:

```bash
go install github.com/edgeo-scada/bacnet/cmd/edgeo-bacnet@latest
```

Basic commands:

```bash
# Discover devices
edgeo-bacnet scan

# Read a property
edgeo-bacnet read -d 1234 -o ai:1 -p pv

# Write a value
edgeo-bacnet write -d 1234 -o ao:1 -p pv -v 72.5

# Watch for changes
edgeo-bacnet watch -d 1234 -o ai:1

# Get device info
edgeo-bacnet info -d 1234
```

## Error Handling

```go
value, err := client.ReadProperty(ctx, deviceID, objectID, propertyID)
if err != nil {
    if bacnet.IsTimeout(err) {
        log.Println("Device not responding")
    } else if bacnet.IsDeviceNotFound(err) {
        log.Println("Device not found on network")
    } else if bacnet.IsPropertyNotFound(err) {
        log.Println("Property does not exist")
    } else if bacnet.IsAccessDenied(err) {
        log.Println("Access denied - check permissions")
    } else {
        log.Printf("Unexpected error: %v", err)
    }
    return
}
```

## Next Steps

- [Client API Reference](client.md) - Complete client operations
- [Configuration Options](options.md) - All available options
- [COV Subscriptions](cov.md) - Real-time value notifications
- [CLI Reference](cli.md) - Command-line tool usage
