# Client API Reference

The `Client` type is the main interface for BACnet/IP communication.

## Creating a Client

```go
func NewClient(opts ...Option) (*Client, error)
```

Creates a new BACnet client with the specified options.

```go
client, err := bacnet.NewClient(
    bacnet.WithTimeout(3 * time.Second),
    bacnet.WithRetries(3),
    bacnet.WithLogger(logger),
)
if err != nil {
    log.Fatal(err)
}
```

See [Configuration Options](options.md) for all available options.

## Connection Management

### Connect

```go
func (c *Client) Connect(ctx context.Context) error
```

Opens the client connection and starts the receiver goroutine.

```go
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
```

If BBMD is configured, the client automatically registers as a foreign device.

### Close

```go
func (c *Client) Close() error
```

Closes the client connection and releases resources.

```go
defer client.Close()
```

### State

```go
func (c *Client) State() ConnectionState
```

Returns the current connection state.

```go
state := client.State()
switch state {
case bacnet.StateDisconnected:
    fmt.Println("Not connected")
case bacnet.StateConnecting:
    fmt.Println("Connecting...")
case bacnet.StateConnected:
    fmt.Println("Connected")
}
```

## Device Discovery

### WhoIs

```go
func (c *Client) WhoIs(ctx context.Context, opts ...DiscoverOption) ([]*DeviceInfo, error)
```

Sends a Who-Is broadcast and collects I-Am responses.

```go
// Discover all devices
devices, err := client.WhoIs(ctx)

// Discover devices in range
devices, err := client.WhoIs(ctx,
    bacnet.WithDeviceRange(1000, 2000),
)

// With custom timeout
devices, err := client.WhoIs(ctx,
    bacnet.WithDiscoveryTimeout(10 * time.Second),
)

// On specific network
devices, err := client.WhoIs(ctx,
    bacnet.WithTargetNetwork(5),
)
```

### GetDevice

```go
func (c *Client) GetDevice(deviceID uint32) (*DeviceInfo, bool)
```

Returns information about a previously discovered device.

```go
if dev, ok := client.GetDevice(1234); ok {
    fmt.Printf("Device: %d\n", dev.ObjectID.Instance)
    fmt.Printf("Vendor: %d\n", dev.VendorID)
    fmt.Printf("Max APDU: %d\n", dev.MaxAPDULength)
}
```

## Property Operations

### ReadProperty

```go
func (c *Client) ReadProperty(ctx context.Context, deviceID uint32, objectID ObjectIdentifier, propertyID PropertyIdentifier, opts ...ReadOption) (interface{}, error)
```

Reads a single property from a BACnet object.

```go
// Read present value
value, err := client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
    bacnet.PropertyPresentValue,
)

// Read with array index
value, err := client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID),
    bacnet.PropertyObjectList,
    bacnet.WithArrayIndex(0), // Returns array length
)

// Read specific array element
value, err := client.ReadProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID),
    bacnet.PropertyObjectList,
    bacnet.WithArrayIndex(1), // First element
)
```

**Return Types:**

| BACnet Type | Go Type |
|-------------|---------|
| Null | `nil` |
| Boolean | `bool` |
| Unsigned Integer | `uint32` |
| Signed Integer | `int32` |
| Real | `float32` |
| Double | `float64` |
| Character String | `string` |
| Octet String | `[]byte` |
| Enumerated | `uint32` |
| Object Identifier | `ObjectIdentifier` |

### WriteProperty

```go
func (c *Client) WriteProperty(ctx context.Context, deviceID uint32, objectID ObjectIdentifier, propertyID PropertyIdentifier, value interface{}, opts ...WriteOption) error
```

Writes a property value to a BACnet object.

```go
// Write a float value
err := client.WriteProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogOutput, 1),
    bacnet.PropertyPresentValue,
    72.5,
)

// Write with priority
err := client.WriteProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeBinaryOutput, 1),
    bacnet.PropertyPresentValue,
    true,
    bacnet.WithPriority(8),
)

// Write to array index
err := client.WriteProperty(ctx, deviceID,
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeSchedule, 1),
    bacnet.PropertyWeeklySchedule,
    scheduleData,
    bacnet.WithWriteArrayIndex(1),
)
```

**Supported Value Types:**

- `nil` - Null value
- `bool` - Boolean
- `int`, `int32` - Signed integer
- `uint32` - Unsigned integer
- `float32` - Real
- `float64` - Double
- `string` - Character string
- `ObjectIdentifier` - Object identifier

### ReadPropertyMultiple

```go
func (c *Client) ReadPropertyMultiple(ctx context.Context, deviceID uint32, requests []ReadPropertyRequest) ([]PropertyValue, error)
```

Reads multiple properties in a single request.

```go
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
        ObjectID:   bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
        PropertyID: bacnet.PropertyStatusFlags,
    },
})
if err != nil {
    log.Fatal(err)
}

for _, pv := range results {
    fmt.Printf("%s.%s = %v\n", pv.ObjectID, pv.PropertyID, pv.Value)
}
```

### GetObjectList

```go
func (c *Client) GetObjectList(ctx context.Context, deviceID uint32) ([]ObjectIdentifier, error)
```

Retrieves the complete list of objects from a device.

```go
objects, err := client.GetObjectList(ctx, deviceID)
if err != nil {
    log.Fatal(err)
}

for _, obj := range objects {
    fmt.Printf("  %s\n", obj)
}
```

## COV Subscriptions

### SubscribeCOV

```go
func (c *Client) SubscribeCOV(ctx context.Context, deviceID uint32, objectID ObjectIdentifier, handler COVHandler, opts ...SubscribeOption) (uint32, error)
```

Subscribes to Change of Value notifications. See [COV Subscriptions](cov.md) for details.

### UnsubscribeCOV

```go
func (c *Client) UnsubscribeCOV(ctx context.Context, deviceID uint32, objectID ObjectIdentifier, subID uint32) error
```

Cancels a COV subscription.

## Metrics

### Metrics

```go
func (c *Client) Metrics() *Metrics
```

Returns the client metrics collector.

```go
metrics := client.Metrics()
snapshot := metrics.Snapshot()

fmt.Printf("Requests sent: %d\n", snapshot.RequestsSent)
fmt.Printf("Requests succeeded: %d\n", snapshot.RequestsSucceeded)
fmt.Printf("Avg latency: %v\n", snapshot.LatencyStats.Avg)
```

See [Metrics](metrics.md) for complete metrics documentation.

## Object Identifiers

### NewObjectIdentifier

```go
func NewObjectIdentifier(objectType ObjectType, instance uint32) ObjectIdentifier
```

Creates a new object identifier.

```go
// Analog Input instance 1
ai1 := bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1)

// Device object
dev := bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID)

// Binary Output instance 5
bo5 := bacnet.NewObjectIdentifier(bacnet.ObjectTypeBinaryOutput, 5)
```

### Object Types

Common object types:

```go
bacnet.ObjectTypeAnalogInput      // 0
bacnet.ObjectTypeAnalogOutput     // 1
bacnet.ObjectTypeAnalogValue      // 2
bacnet.ObjectTypeBinaryInput      // 3
bacnet.ObjectTypeBinaryOutput     // 4
bacnet.ObjectTypeBinaryValue      // 5
bacnet.ObjectTypeCalendar         // 6
bacnet.ObjectTypeDevice           // 8
bacnet.ObjectTypeMultiStateInput  // 13
bacnet.ObjectTypeMultiStateOutput // 14
bacnet.ObjectTypeMultiStateValue  // 19
bacnet.ObjectTypeSchedule         // 17
bacnet.ObjectTypeTrendLog         // 20
```

### Property Identifiers

Common property identifiers:

```go
bacnet.PropertyPresentValue     // 85
bacnet.PropertyObjectName       // 77
bacnet.PropertyObjectIdentifier // 75
bacnet.PropertyObjectType       // 79
bacnet.PropertyDescription      // 28
bacnet.PropertyStatusFlags      // 111
bacnet.PropertyEventState       // 36
bacnet.PropertyOutOfService     // 81
bacnet.PropertyUnits            // 117
bacnet.PropertyPriorityArray    // 87
bacnet.PropertyRelinquishDefault // 104
bacnet.PropertyObjectList       // 76
```

## Thread Safety

The `Client` is safe for concurrent use from multiple goroutines. Internal state is protected by mutexes and atomic operations.

```go
// Safe to call from multiple goroutines
go func() {
    value, _ := client.ReadProperty(ctx, deviceID, obj1, prop)
}()

go func() {
    value, _ := client.ReadProperty(ctx, deviceID, obj2, prop)
}()
```
