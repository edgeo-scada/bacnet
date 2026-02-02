# Device Discovery

BACnet device discovery uses the Who-Is/I-Am protocol to find devices on the network.

## How Discovery Works

1. The client sends a **Who-Is** broadcast message
2. All devices (or devices in the specified range) respond with **I-Am** messages
3. The client collects responses and builds a device database
4. Discovered device information is cached for subsequent operations

## Basic Discovery

### Discover All Devices

```go
ctx := context.Background()

// Discover all devices on the local network
devices, err := client.WhoIs(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d devices\n", len(devices))
for _, dev := range devices {
    fmt.Printf("Device %d:\n", dev.ObjectID.Instance)
    fmt.Printf("  Vendor ID: %d\n", dev.VendorID)
    fmt.Printf("  Max APDU: %d\n", dev.MaxAPDULength)
    fmt.Printf("  Segmentation: %s\n", dev.Segmentation)
}
```

### Discover Devices in Range

```go
// Only discover devices with IDs between 1000 and 2000
devices, err := client.WhoIs(ctx,
    bacnet.WithDeviceRange(1000, 2000),
)
```

### Discover a Specific Device

```go
// Find a specific device
targetID := uint32(1234)
devices, err := client.WhoIs(ctx,
    bacnet.WithDeviceRange(targetID, targetID),
    bacnet.WithDiscoveryTimeout(3 * time.Second),
)

if len(devices) == 0 {
    log.Printf("Device %d not found", targetID)
}
```

## Discovery Options

### WithDeviceRange

Limits discovery to devices within a specific instance ID range.

```go
// Discover devices 1-1000
devices, err := client.WhoIs(ctx,
    bacnet.WithDeviceRange(1, 1000),
)
```

### WithDiscoveryTimeout

Sets how long to wait for I-Am responses.

```go
// Wait up to 10 seconds
devices, err := client.WhoIs(ctx,
    bacnet.WithDiscoveryTimeout(10 * time.Second),
)
```

**Default:** 5 seconds

### WithTargetNetwork

Specifies a remote BACnet network to discover.

```go
// Discover devices on network 5
devices, err := client.WhoIs(ctx,
    bacnet.WithTargetNetwork(5),
)
```

**Note:** Requires a router to the target network.

## Device Information

The `DeviceInfo` struct contains information collected from I-Am responses:

```go
type DeviceInfo struct {
    ObjectID            ObjectIdentifier  // Device object ID
    Address             Address           // Network address
    MaxAPDULength       uint16            // Max APDU size supported
    Segmentation        Segmentation      // Segmentation capability
    VendorID            uint16            // Manufacturer vendor ID
    VendorName          string            // Vendor name (if retrieved)
    ModelName           string            // Model name (if retrieved)
    FirmwareRevision    string            // Firmware version (if retrieved)
    ApplicationSoftware string            // Application software version
    Description         string            // Device description
    Location            string            // Device location
    ObjectList          []ObjectIdentifier // List of objects (if retrieved)
}
```

## Retrieving Cached Device Info

After discovery, device information is cached:

```go
// Get cached device info
if dev, ok := client.GetDevice(1234); ok {
    fmt.Printf("Device %d cached:\n", dev.ObjectID.Instance)
    fmt.Printf("  Address: %v\n", dev.Address)
    fmt.Printf("  Vendor: %d\n", dev.VendorID)
} else {
    fmt.Println("Device not in cache")
}
```

## Extended Device Information

The I-Am response provides basic information. For extended details, read device properties:

```go
deviceID := uint32(1234)

// Read vendor name
vendorName, err := client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID),
    bacnet.PropertyVendorName,
)

// Read model name
modelName, err := client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID),
    bacnet.PropertyModelName,
)

// Read firmware revision
firmware, err := client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID),
    bacnet.PropertyFirmwareRevision,
)

// Read application software version
appVersion, err := client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID),
    bacnet.PropertyApplicationSoftwareVersion,
)

// Read description
description, err := client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID),
    bacnet.PropertyDescription,
)

// Read location
location, err := client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID),
    bacnet.PropertyLocation,
)
```

## Getting Object List

Retrieve all objects from a device:

```go
objects, err := client.GetObjectList(ctx, deviceID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Device %d has %d objects:\n", deviceID, len(objects))

// Group by object type
byType := make(map[bacnet.ObjectType][]bacnet.ObjectIdentifier)
for _, obj := range objects {
    byType[obj.Type] = append(byType[obj.Type], obj)
}

for objType, objs := range byType {
    fmt.Printf("  %s: %d objects\n", objType, len(objs))
}
```

## Cross-Subnet Discovery (BBMD)

To discover devices on different subnets, configure BBMD:

```go
// Configure client with BBMD
client, err := bacnet.NewClient(
    bacnet.WithBBMD("192.168.1.1", 47808, 60*time.Second),
)

// Connect - this registers as a foreign device
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}

// Now discovery works across subnets
devices, err := client.WhoIs(ctx)
```

## Discovery Patterns

### Periodic Discovery

```go
func periodicDiscovery(ctx context.Context, client *bacnet.Client, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            devices, err := client.WhoIs(ctx,
                bacnet.WithDiscoveryTimeout(5*time.Second),
            )
            if err != nil {
                log.Printf("Discovery error: %v", err)
                continue
            }
            log.Printf("Found %d devices", len(devices))
        }
    }
}
```

### Discovery with Retry

```go
func discoverWithRetry(ctx context.Context, client *bacnet.Client, targetID uint32) (*bacnet.DeviceInfo, error) {
    for attempt := 0; attempt < 3; attempt++ {
        devices, err := client.WhoIs(ctx,
            bacnet.WithDeviceRange(targetID, targetID),
            bacnet.WithDiscoveryTimeout(2*time.Second),
        )
        if err != nil {
            return nil, err
        }

        if len(devices) > 0 {
            return devices[0], nil
        }

        // Wait before retry
        time.Sleep(time.Duration(attempt+1) * time.Second)
    }

    return nil, bacnet.ErrDeviceNotFound
}
```

### Full Device Scan

```go
func scanDevice(ctx context.Context, client *bacnet.Client, deviceID uint32) error {
    // Get device info
    dev, ok := client.GetDevice(deviceID)
    if !ok {
        _, err := client.WhoIs(ctx,
            bacnet.WithDeviceRange(deviceID, deviceID),
        )
        if err != nil {
            return err
        }
        dev, ok = client.GetDevice(deviceID)
        if !ok {
            return bacnet.ErrDeviceNotFound
        }
    }

    fmt.Printf("Device: %d\n", dev.ObjectID.Instance)
    fmt.Printf("Vendor: %d\n", dev.VendorID)

    // Read extended info
    props := []struct {
        name string
        prop bacnet.PropertyIdentifier
    }{
        {"Vendor Name", bacnet.PropertyVendorName},
        {"Model Name", bacnet.PropertyModelName},
        {"Firmware", bacnet.PropertyFirmwareRevision},
        {"Description", bacnet.PropertyDescription},
        {"Location", bacnet.PropertyLocation},
    }

    deviceObj := bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID)

    for _, p := range props {
        value, err := client.ReadProperty(ctx, deviceID, deviceObj, p.prop)
        if err == nil {
            fmt.Printf("%s: %v\n", p.name, value)
        }
    }

    // Get object list
    objects, err := client.GetObjectList(ctx, deviceID)
    if err != nil {
        return err
    }

    fmt.Printf("\nObjects (%d):\n", len(objects))
    for _, obj := range objects {
        name, _ := client.ReadProperty(ctx, deviceID, obj, bacnet.PropertyObjectName)
        fmt.Printf("  %s: %v\n", obj, name)
    }

    return nil
}
```

## CLI Discovery

Using the CLI tool:

```bash
# Discover all devices
edgeo-bacnet scan

# Discover specific range
edgeo-bacnet scan --low 1000 --high 2000

# Longer timeout
edgeo-bacnet scan -t 10s

# JSON output
edgeo-bacnet scan -o json

# Get full device info
edgeo-bacnet info -d 1234

# Dump all objects
edgeo-bacnet dump -d 1234
```
