# Device Discovery Example

This example demonstrates comprehensive device discovery patterns for BACnet networks.

## Complete Example

```go
// Package main demonstrates BACnet device discovery patterns
package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "log/slog"
    "os"
    "sort"
    "time"

    "github.com/edgeo-scada/bacnet/bacnet"
)

type DeviceDetails struct {
    DeviceID     uint32   `json:"device_id"`
    VendorID     uint16   `json:"vendor_id"`
    VendorName   string   `json:"vendor_name,omitempty"`
    ModelName    string   `json:"model_name,omitempty"`
    Firmware     string   `json:"firmware,omitempty"`
    Description  string   `json:"description,omitempty"`
    Location     string   `json:"location,omitempty"`
    MaxAPDU      uint16   `json:"max_apdu"`
    Segmentation string   `json:"segmentation"`
    ObjectCount  int      `json:"object_count,omitempty"`
    ObjectTypes  []string `json:"object_types,omitempty"`
}

func main() {
    // Command line flags
    lowLimit := flag.Uint("low", 0, "Low device ID limit")
    highLimit := flag.Uint("high", 0, "High device ID limit (0 = no limit)")
    timeout := flag.Duration("timeout", 5*time.Second, "Discovery timeout")
    detailed := flag.Bool("detailed", false, "Fetch detailed device info")
    jsonOutput := flag.Bool("json", false, "Output as JSON")
    bbmdAddr := flag.String("bbmd", "", "BBMD address for cross-subnet discovery")
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

    // Build client options
    opts := []bacnet.Option{
        bacnet.WithTimeout(3 * time.Second),
        bacnet.WithRetries(3),
        bacnet.WithLogger(logger),
    }

    // Add BBMD if specified
    if *bbmdAddr != "" {
        opts = append(opts, bacnet.WithBBMD(*bbmdAddr, 47808, 60*time.Second))
    }

    // Create client
    client, err := bacnet.NewClient(opts...)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    ctx := context.Background()

    // Connect
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Close()

    if !*jsonOutput {
        fmt.Println("Discovering BACnet devices...")
        if *bbmdAddr != "" {
            fmt.Printf("Using BBMD: %s\n", *bbmdAddr)
        }
    }

    // Build discovery options
    var discoverOpts []bacnet.DiscoverOption
    discoverOpts = append(discoverOpts, bacnet.WithDiscoveryTimeout(*timeout))

    if *lowLimit > 0 || *highLimit > 0 {
        low := uint32(*lowLimit)
        high := uint32(*highLimit)
        if high == 0 {
            high = 0x3FFFFF // Max BACnet device ID
        }
        discoverOpts = append(discoverOpts, bacnet.WithDeviceRange(low, high))
        if !*jsonOutput {
            fmt.Printf("Searching range: %d - %d\n", low, high)
        }
    }

    // Run discovery
    startTime := time.Now()
    devices, err := client.WhoIs(ctx, discoverOpts...)
    if err != nil {
        log.Fatalf("Discovery failed: %v", err)
    }
    discoveryDuration := time.Since(startTime)

    if len(devices) == 0 {
        if *jsonOutput {
            fmt.Println("[]")
        } else {
            fmt.Println("No devices found")
        }
        return
    }

    // Sort by device ID
    sort.Slice(devices, func(i, j int) bool {
        return devices[i].ObjectID.Instance < devices[j].ObjectID.Instance
    })

    // Collect device details
    var details []DeviceDetails
    for _, dev := range devices {
        dd := DeviceDetails{
            DeviceID:     dev.ObjectID.Instance,
            VendorID:     dev.VendorID,
            MaxAPDU:      dev.MaxAPDULength,
            Segmentation: dev.Segmentation.String(),
        }

        if *detailed {
            // Fetch additional properties
            deviceObj := bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, dev.ObjectID.Instance)

            if v, err := client.ReadProperty(ctx, dev.ObjectID.Instance, deviceObj, bacnet.PropertyVendorName); err == nil {
                if s, ok := v.(string); ok {
                    dd.VendorName = s
                }
            }

            if v, err := client.ReadProperty(ctx, dev.ObjectID.Instance, deviceObj, bacnet.PropertyModelName); err == nil {
                if s, ok := v.(string); ok {
                    dd.ModelName = s
                }
            }

            if v, err := client.ReadProperty(ctx, dev.ObjectID.Instance, deviceObj, bacnet.PropertyFirmwareRevision); err == nil {
                if s, ok := v.(string); ok {
                    dd.Firmware = s
                }
            }

            if v, err := client.ReadProperty(ctx, dev.ObjectID.Instance, deviceObj, bacnet.PropertyDescription); err == nil {
                if s, ok := v.(string); ok {
                    dd.Description = s
                }
            }

            if v, err := client.ReadProperty(ctx, dev.ObjectID.Instance, deviceObj, bacnet.PropertyLocation); err == nil {
                if s, ok := v.(string); ok {
                    dd.Location = s
                }
            }

            // Get object count
            if objects, err := client.GetObjectList(ctx, dev.ObjectID.Instance); err == nil {
                dd.ObjectCount = len(objects)

                // Count by type
                typeCount := make(map[bacnet.ObjectType]int)
                for _, obj := range objects {
                    typeCount[obj.Type]++
                }
                for objType := range typeCount {
                    dd.ObjectTypes = append(dd.ObjectTypes, objType.String())
                }
                sort.Strings(dd.ObjectTypes)
            }
        }

        details = append(details, dd)
    }

    // Output results
    if *jsonOutput {
        output, _ := json.MarshalIndent(details, "", "  ")
        fmt.Println(string(output))
    } else {
        fmt.Printf("\nFound %d device(s) in %v:\n\n", len(devices), discoveryDuration.Round(time.Millisecond))

        for _, dd := range details {
            fmt.Printf("Device %d\n", dd.DeviceID)
            fmt.Printf("  Vendor ID:     %d\n", dd.VendorID)
            if dd.VendorName != "" {
                fmt.Printf("  Vendor Name:   %s\n", dd.VendorName)
            }
            if dd.ModelName != "" {
                fmt.Printf("  Model:         %s\n", dd.ModelName)
            }
            if dd.Firmware != "" {
                fmt.Printf("  Firmware:      %s\n", dd.Firmware)
            }
            if dd.Description != "" {
                fmt.Printf("  Description:   %s\n", dd.Description)
            }
            if dd.Location != "" {
                fmt.Printf("  Location:      %s\n", dd.Location)
            }
            fmt.Printf("  Max APDU:      %d\n", dd.MaxAPDU)
            fmt.Printf("  Segmentation:  %s\n", dd.Segmentation)
            if dd.ObjectCount > 0 {
                fmt.Printf("  Objects:       %d (%v)\n", dd.ObjectCount, dd.ObjectTypes)
            }
            fmt.Println()
        }
    }

    // Print metrics
    if !*jsonOutput && *verbose {
        metrics := client.Metrics().Snapshot()
        fmt.Println("--- Discovery Metrics ---")
        fmt.Printf("  Who-Is sent: %d\n", metrics.WhoIsSent)
        fmt.Printf("  I-Am received: %d\n", metrics.IAmReceived)
        fmt.Printf("  Devices discovered: %d\n", metrics.DevicesDiscovered)
        fmt.Printf("  Requests sent: %d\n", metrics.RequestsSent)
        fmt.Printf("  Avg latency: %v\n", metrics.LatencyStats.Avg)
    }
}
```

## Running the Example

```bash
# Basic discovery
go run main.go

# Discovery with device range
go run main.go -low 1000 -high 2000

# Detailed device information
go run main.go -detailed

# JSON output
go run main.go -json

# Cross-subnet discovery via BBMD
go run main.go -bbmd 192.168.1.1

# Combined options
go run main.go -detailed -json -timeout 10s -verbose
```

## Expected Output

### Standard Output

```
Discovering BACnet devices...

Found 3 device(s) in 2.345s:

Device 1234
  Vendor ID:     123
  Vendor Name:   ACME Controls
  Model:         VAV-2000
  Firmware:      v2.3.1
  Description:   VAV Box Floor 2
  Location:      Room 201
  Max APDU:      1476
  Segmentation:  no-segmentation
  Objects:       45 ([analog-input analog-output binary-input schedule])

Device 2345
  Vendor ID:     456
  Vendor Name:   BuildCo
  Model:         AHU-500
  Firmware:      v1.0.5
  Max APDU:      480
  Segmentation:  no-segmentation

Device 5678
  Vendor ID:     123
  Vendor Name:   ACME Controls
  Model:         BC-1000
  Firmware:      v3.1.0
  Description:   Building Controller
  Location:      Mechanical Room
  Max APDU:      1476
  Segmentation:  segmented-both
  Objects:       127 ([analog-input analog-output analog-value binary-input binary-output schedule trend-log])
```

### JSON Output

```json
[
  {
    "device_id": 1234,
    "vendor_id": 123,
    "vendor_name": "ACME Controls",
    "model_name": "VAV-2000",
    "firmware": "v2.3.1",
    "description": "VAV Box Floor 2",
    "location": "Room 201",
    "max_apdu": 1476,
    "segmentation": "no-segmentation",
    "object_count": 45,
    "object_types": ["analog-input", "analog-output", "binary-input", "schedule"]
  },
  {
    "device_id": 2345,
    "vendor_id": 456,
    "vendor_name": "BuildCo",
    "model_name": "AHU-500",
    "firmware": "v1.0.5",
    "max_apdu": 480,
    "segmentation": "no-segmentation"
  }
]
```

## Discovery Patterns

### Periodic Network Scan

```go
func periodicScan(ctx context.Context, client *bacnet.Client, interval time.Duration, onChange func([]*bacnet.DeviceInfo)) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    var lastDevices map[uint32]bool

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            devices, err := client.WhoIs(ctx, bacnet.WithDiscoveryTimeout(5*time.Second))
            if err != nil {
                log.Printf("Scan failed: %v", err)
                continue
            }

            // Check for changes
            currentDevices := make(map[uint32]bool)
            for _, dev := range devices {
                currentDevices[dev.ObjectID.Instance] = true
            }

            if !mapsEqual(lastDevices, currentDevices) {
                onChange(devices)
            }
            lastDevices = currentDevices
        }
    }
}
```

### Find Specific Device

```go
func findDevice(ctx context.Context, client *bacnet.Client, deviceID uint32) (*bacnet.DeviceInfo, error) {
    // Try direct discovery
    devices, err := client.WhoIs(ctx,
        bacnet.WithDeviceRange(deviceID, deviceID),
        bacnet.WithDiscoveryTimeout(3*time.Second),
    )
    if err != nil {
        return nil, err
    }

    if len(devices) > 0 {
        return devices[0], nil
    }

    // Retry with longer timeout
    devices, err = client.WhoIs(ctx,
        bacnet.WithDeviceRange(deviceID, deviceID),
        bacnet.WithDiscoveryTimeout(10*time.Second),
    )
    if err != nil {
        return nil, err
    }

    if len(devices) > 0 {
        return devices[0], nil
    }

    return nil, bacnet.ErrDeviceNotFound
}
```

### Parallel Device Detail Fetching

```go
func fetchDeviceDetails(ctx context.Context, client *bacnet.Client, devices []*bacnet.DeviceInfo) []DeviceDetails {
    results := make(chan DeviceDetails, len(devices))

    for _, dev := range devices {
        go func(d *bacnet.DeviceInfo) {
            dd := DeviceDetails{
                DeviceID: d.ObjectID.Instance,
                VendorID: d.VendorID,
            }

            deviceObj := bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, d.ObjectID.Instance)

            // Fetch properties (with timeout per device)
            devCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
            defer cancel()

            if v, err := client.ReadProperty(devCtx, d.ObjectID.Instance, deviceObj, bacnet.PropertyVendorName); err == nil {
                dd.VendorName, _ = v.(string)
            }
            if v, err := client.ReadProperty(devCtx, d.ObjectID.Instance, deviceObj, bacnet.PropertyModelName); err == nil {
                dd.ModelName, _ = v.(string)
            }

            results <- dd
        }(dev)
    }

    var details []DeviceDetails
    for range devices {
        details = append(details, <-results)
    }
    return details
}
```

## Key Concepts

1. **Range-based Discovery** - Limiting search to specific device ID ranges
2. **Cross-subnet Discovery** - Using BBMD for remote network access
3. **Detailed Information Gathering** - Reading extended device properties
4. **JSON Output** - Machine-readable output for automation
5. **Parallel Fetching** - Concurrent property reads for efficiency

## Related Documentation

- [Client API](../client.md) - Full client reference
- [COV Subscriptions](../cov.md) - Real-time monitoring
- [CLI Reference](../cli.md) - Command-line tool usage
