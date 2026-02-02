# Basic Client Example

This example demonstrates fundamental BACnet client operations including discovery, reading, and writing properties.

## Complete Example

```go
// Package main demonstrates basic BACnet client usage
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "log/slog"
    "os"
    "time"

    "github.com/edgeo/drivers/bacnet/bacnet"
)

func main() {
    // Command line flags
    deviceID := flag.Uint("device", 0, "Target device ID (0 = auto-discover)")
    timeout := flag.Duration("timeout", 3*time.Second, "Request timeout")
    verbose := flag.Bool("verbose", false, "Enable verbose logging")
    flag.Parse()

    // Setup logging
    logLevel := slog.LevelInfo
    if *verbose {
        logLevel = slog.LevelDebug
    }
    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: logLevel,
    }))

    // Create BACnet client
    client, err := bacnet.NewClient(
        bacnet.WithTimeout(*timeout),
        bacnet.WithRetries(3),
        bacnet.WithRetryDelay(500*time.Millisecond),
        bacnet.WithLogger(logger),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    ctx := context.Background()

    // Connect to network
    fmt.Println("Connecting to BACnet network...")
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Close()

    fmt.Println("Connected successfully")

    // Discover devices
    fmt.Println("\nDiscovering devices...")
    devices, err := client.WhoIs(ctx, bacnet.WithDiscoveryTimeout(5*time.Second))
    if err != nil {
        log.Fatalf("Discovery failed: %v", err)
    }

    fmt.Printf("Found %d device(s):\n", len(devices))
    for _, dev := range devices {
        fmt.Printf("  Device %d (Vendor ID: %d, Max APDU: %d)\n",
            dev.ObjectID.Instance,
            dev.VendorID,
            dev.MaxAPDULength,
        )
    }

    if len(devices) == 0 {
        fmt.Println("\nNo devices found on network")
        return
    }

    // Select target device
    var targetID uint32
    if *deviceID > 0 {
        targetID = uint32(*deviceID)
    } else {
        targetID = devices[0].ObjectID.Instance
    }
    fmt.Printf("\nUsing device %d\n", targetID)

    // Read device properties
    fmt.Println("\n--- Device Properties ---")
    deviceObj := bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, targetID)

    readDeviceProperty := func(prop bacnet.PropertyIdentifier, name string) {
        value, err := client.ReadProperty(ctx, targetID, deviceObj, prop)
        if err != nil {
            fmt.Printf("  %s: (error: %v)\n", name, err)
        } else {
            fmt.Printf("  %s: %v\n", name, value)
        }
    }

    readDeviceProperty(bacnet.PropertyObjectName, "Object Name")
    readDeviceProperty(bacnet.PropertyVendorName, "Vendor Name")
    readDeviceProperty(bacnet.PropertyModelName, "Model Name")
    readDeviceProperty(bacnet.PropertyFirmwareRevision, "Firmware")
    readDeviceProperty(bacnet.PropertyDescription, "Description")
    readDeviceProperty(bacnet.PropertyLocation, "Location")

    // Get object list
    fmt.Println("\n--- Object List ---")
    objects, err := client.GetObjectList(ctx, targetID)
    if err != nil {
        log.Printf("Failed to get object list: %v", err)
    } else {
        // Count by type
        byType := make(map[bacnet.ObjectType]int)
        for _, obj := range objects {
            byType[obj.Type]++
        }
        for objType, count := range byType {
            fmt.Printf("  %s: %d\n", objType, count)
        }
    }

    // Read analog inputs
    fmt.Println("\n--- Analog Inputs ---")
    for i := uint32(1); i <= 5; i++ {
        obj := bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, i)

        // Try to read the value
        value, err := client.ReadProperty(ctx, targetID, obj, bacnet.PropertyPresentValue)
        if err != nil {
            if bacnet.IsPropertyNotFound(err) || bacnet.IsDeviceNotFound(err) {
                continue // Object doesn't exist
            }
            fmt.Printf("  AI:%d - Error: %v\n", i, err)
            continue
        }

        // Read name and units
        name, _ := client.ReadProperty(ctx, targetID, obj, bacnet.PropertyObjectName)
        units, _ := client.ReadProperty(ctx, targetID, obj, bacnet.PropertyUnits)

        fmt.Printf("  AI:%d - %v = %v %v\n", i, name, value, units)
    }

    // Read binary inputs
    fmt.Println("\n--- Binary Inputs ---")
    for i := uint32(1); i <= 5; i++ {
        obj := bacnet.NewObjectIdentifier(bacnet.ObjectTypeBinaryInput, i)

        value, err := client.ReadProperty(ctx, targetID, obj, bacnet.PropertyPresentValue)
        if err != nil {
            if bacnet.IsPropertyNotFound(err) || bacnet.IsDeviceNotFound(err) {
                continue
            }
            fmt.Printf("  BI:%d - Error: %v\n", i, err)
            continue
        }

        name, _ := client.ReadProperty(ctx, targetID, obj, bacnet.PropertyObjectName)

        state := "inactive"
        if v, ok := value.(bool); ok && v {
            state = "active"
        }
        fmt.Printf("  BI:%d - %v = %s\n", i, name, state)
    }

    // Example: Read multiple properties at once
    fmt.Println("\n--- ReadPropertyMultiple Example ---")
    aiObj := bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1)
    results, err := client.ReadPropertyMultiple(ctx, targetID, []bacnet.ReadPropertyRequest{
        {ObjectID: aiObj, PropertyID: bacnet.PropertyObjectName},
        {ObjectID: aiObj, PropertyID: bacnet.PropertyPresentValue},
        {ObjectID: aiObj, PropertyID: bacnet.PropertyUnits},
        {ObjectID: aiObj, PropertyID: bacnet.PropertyStatusFlags},
    })
    if err != nil {
        fmt.Printf("ReadPropertyMultiple error: %v\n", err)
    } else {
        for _, pv := range results {
            fmt.Printf("  %s = %v\n", pv.PropertyID, pv.Value)
        }
    }

    // Example: Write a value (commented out for safety)
    /*
    fmt.Println("\n--- Write Example ---")
    aoObj := bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogOutput, 1)
    err = client.WriteProperty(ctx, targetID, aoObj,
        bacnet.PropertyPresentValue,
        72.5,
        bacnet.WithPriority(8), // Manual operator priority
    )
    if err != nil {
        fmt.Printf("Write failed: %v\n", err)
    } else {
        fmt.Println("Write successful")
    }
    */

    // Print metrics summary
    fmt.Println("\n--- Metrics ---")
    metrics := client.Metrics().Snapshot()
    fmt.Printf("  Uptime: %v\n", metrics.Uptime)
    fmt.Printf("  Requests sent: %d\n", metrics.RequestsSent)
    fmt.Printf("  Requests succeeded: %d\n", metrics.RequestsSucceeded)
    fmt.Printf("  Requests failed: %d\n", metrics.RequestsFailed)
    fmt.Printf("  Avg latency: %v\n", metrics.LatencyStats.Avg)
    fmt.Printf("  Bytes sent: %d\n", metrics.BytesSent)
    fmt.Printf("  Bytes received: %d\n", metrics.BytesReceived)
}
```

## Running the Example

```bash
# Build and run
go run main.go

# Specify target device
go run main.go -device 1234

# Enable verbose logging
go run main.go -verbose

# Custom timeout
go run main.go -timeout 5s
```

## Expected Output

```
Connecting to BACnet network...
Connected successfully

Discovering devices...
Found 2 device(s):
  Device 1234 (Vendor ID: 123, Max APDU: 1476)
  Device 5678 (Vendor ID: 456, Max APDU: 480)

Using device 1234

--- Device Properties ---
  Object Name: Building Controller
  Vendor Name: ACME Controls
  Model Name: BC-2000
  Firmware: v2.3.1
  Description: Main building controller
  Location: Mechanical Room 1

--- Object List ---
  device: 1
  analog-input: 12
  analog-output: 6
  analog-value: 8
  binary-input: 5
  binary-output: 4
  schedule: 3
  trend-log: 7

--- Analog Inputs ---
  AI:1 - Zone Temp = 72.5 °F
  AI:2 - Supply Air = 55.3 °F
  AI:3 - Return Air = 71.8 °F
  AI:4 - Outside Air = 45.2 °F
  AI:5 - Humidity = 45 %RH

--- Binary Inputs ---
  BI:1 - Occupied = active
  BI:2 - Fire Alarm = inactive
  BI:3 - Filter Status = inactive

--- ReadPropertyMultiple Example ---
  object-name = Zone Temp
  present-value = 72.5
  units = 64
  status-flags = [0]

--- Metrics ---
  Uptime: 3.456s
  Requests sent: 32
  Requests succeeded: 31
  Requests failed: 1
  Avg latency: 45.123ms
  Bytes sent: 1824
  Bytes received: 2156
```

## Key Concepts Demonstrated

1. **Client Creation** - Using functional options for configuration
2. **Connection Management** - Connect/Close with defer pattern
3. **Device Discovery** - Who-Is with timeout
4. **Single Property Reads** - ReadProperty for individual values
5. **Batch Reads** - ReadPropertyMultiple for efficiency
6. **Error Handling** - Checking for specific error conditions
7. **Metrics Collection** - Accessing client metrics
8. **Object Identifiers** - Creating and using ObjectIdentifier
9. **Property Identifiers** - Standard BACnet properties

## Next Steps

- [Device Discovery Example](discovery.md) - Advanced discovery patterns
- [COV Subscriptions](../cov.md) - Real-time value monitoring
- [Error Handling](../errors.md) - Comprehensive error patterns
