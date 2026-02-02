# Configuration Options

The BACnet client supports functional options for flexible configuration.

## Client Options

Options are passed to `NewClient()` to configure client behavior.

### WithDeviceID

```go
func WithDeviceID(id uint32) Option
```

Sets the local device ID for the client. Used when the client needs to identify itself.

```go
client, err := bacnet.NewClient(
    bacnet.WithDeviceID(99999),
)
```

**Default:** `0xFFFFFFFF` (uninitialized)

### WithLocalAddress

```go
func WithLocalAddress(addr string) Option
```

Sets the local address to bind to.

```go
client, err := bacnet.NewClient(
    bacnet.WithLocalAddress("192.168.1.100:47808"),
)
```

**Default:** System-assigned address

### WithNetworkNumber

```go
func WithNetworkNumber(net uint16) Option
```

Sets the BACnet network number for the client.

```go
client, err := bacnet.NewClient(
    bacnet.WithNetworkNumber(5),
)
```

**Default:** `0` (local network)

### WithBBMD

```go
func WithBBMD(addr string, port int, ttl time.Duration) Option
```

Configures BBMD (BACnet Broadcast Management Device) for foreign device registration. Required for cross-subnet communication.

```go
client, err := bacnet.NewClient(
    bacnet.WithBBMD("192.168.1.1", 47808, 60*time.Second),
)
```

**Default:** Not configured (local network only)

### WithTimeout

```go
func WithTimeout(d time.Duration) Option
```

Sets the request timeout for BACnet operations.

```go
client, err := bacnet.NewClient(
    bacnet.WithTimeout(5 * time.Second),
)
```

**Default:** `3s`

### WithRetries

```go
func WithRetries(n int) Option
```

Sets the number of retries for failed requests.

```go
client, err := bacnet.NewClient(
    bacnet.WithRetries(5),
)
```

**Default:** `3`

### WithRetryDelay

```go
func WithRetryDelay(d time.Duration) Option
```

Sets the delay between retry attempts.

```go
client, err := bacnet.NewClient(
    bacnet.WithRetryDelay(1 * time.Second),
)
```

**Default:** `500ms`

### WithMaxAPDULength

```go
func WithMaxAPDULength(length uint16) Option
```

Sets the maximum APDU length accepted by the client.

```go
client, err := bacnet.NewClient(
    bacnet.WithMaxAPDULength(1024),
)
```

**Default:** `1476` (maximum for BACnet/IP)

### WithSegmentation

```go
func WithSegmentation(seg Segmentation) Option
```

Sets the segmentation capability advertised by the client.

```go
client, err := bacnet.NewClient(
    bacnet.WithSegmentation(bacnet.SegmentationBoth),
)
```

**Values:**
- `SegmentationBoth` - Can send and receive segmented messages
- `SegmentationTransmit` - Can only send segmented messages
- `SegmentationReceive` - Can only receive segmented messages
- `SegmentationNone` - No segmentation support

**Default:** `SegmentationNone`

### WithProposedWindowSize

```go
func WithProposedWindowSize(size uint8) Option
```

Sets the proposed window size for segmented transfers.

```go
client, err := bacnet.NewClient(
    bacnet.WithProposedWindowSize(4),
)
```

**Default:** `1`

### WithAutoDiscover

```go
func WithAutoDiscover(enable bool) Option
```

Enables automatic device discovery on connect.

```go
client, err := bacnet.NewClient(
    bacnet.WithAutoDiscover(true),
)
```

**Default:** `false`

### WithDiscoverTimeout

```go
func WithDiscoverTimeout(d time.Duration) Option
```

Sets the timeout for automatic device discovery.

```go
client, err := bacnet.NewClient(
    bacnet.WithDiscoverTimeout(10 * time.Second),
)
```

**Default:** `5s`

### WithLogger

```go
func WithLogger(logger *slog.Logger) Option
```

Sets the logger for the client.

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
client, err := bacnet.NewClient(
    bacnet.WithLogger(logger),
)
```

**Default:** `slog.Default()`

## Discovery Options

Options for the `WhoIs()` method.

### WithDeviceRange

```go
func WithDeviceRange(low, high uint32) DiscoverOption
```

Limits discovery to devices within a specific instance ID range.

```go
devices, err := client.WhoIs(ctx,
    bacnet.WithDeviceRange(1000, 2000),
)
```

### WithDiscoveryTimeout

```go
func WithDiscoveryTimeout(d time.Duration) DiscoverOption
```

Sets the time to wait for I-Am responses.

```go
devices, err := client.WhoIs(ctx,
    bacnet.WithDiscoveryTimeout(10 * time.Second),
)
```

**Default:** `5s`

### WithTargetNetwork

```go
func WithTargetNetwork(net uint16) DiscoverOption
```

Specifies a target network for discovery.

```go
devices, err := client.WhoIs(ctx,
    bacnet.WithTargetNetwork(5),
)
```

**Default:** `0` (local network)

## Read Options

Options for `ReadProperty()`.

### WithArrayIndex

```go
func WithArrayIndex(index uint32) ReadOption
```

Specifies an array index for reading array properties.

```go
// Get array length
length, err := client.ReadProperty(ctx, deviceID, objectID,
    bacnet.PropertyObjectList,
    bacnet.WithArrayIndex(0),
)

// Get specific element
element, err := client.ReadProperty(ctx, deviceID, objectID,
    bacnet.PropertyObjectList,
    bacnet.WithArrayIndex(1),
)
```

## Write Options

Options for `WriteProperty()`.

### WithWriteArrayIndex

```go
func WithWriteArrayIndex(index uint32) WriteOption
```

Specifies an array index for writing to array properties.

```go
err := client.WriteProperty(ctx, deviceID, objectID,
    bacnet.PropertyWeeklySchedule,
    scheduleData,
    bacnet.WithWriteArrayIndex(1),
)
```

### WithPriority

```go
func WithPriority(priority uint8) WriteOption
```

Sets the priority for the write operation (1-16, where 1 is highest).

```go
err := client.WriteProperty(ctx, deviceID, objectID,
    bacnet.PropertyPresentValue,
    72.5,
    bacnet.WithPriority(8), // Manual operator priority
)
```

**Priority Levels:**

| Priority | Typical Use |
|----------|-------------|
| 1 | Manual-Life Safety |
| 2 | Automatic-Life Safety |
| 3 | Available |
| 4 | Available |
| 5 | Critical Equipment Control |
| 6 | Minimum On/Off |
| 7 | Available |
| 8 | Manual Operator |
| 9-15 | Available |
| 16 | Default/Relinquish |

## Subscribe Options

Options for `SubscribeCOV()`.

### WithSubscriptionLifetime

```go
func WithSubscriptionLifetime(seconds uint32) SubscribeOption
```

Sets the subscription lifetime in seconds.

```go
subID, err := client.SubscribeCOV(ctx, deviceID, objectID, handler,
    bacnet.WithSubscriptionLifetime(300), // 5 minutes
)
```

**Default:** No lifetime (subscription remains until cancelled)

### WithCOVIncrement

```go
func WithCOVIncrement(increment float32) SubscribeOption
```

Sets the COV increment for analog values.

```go
subID, err := client.SubscribeCOV(ctx, deviceID, objectID, handler,
    bacnet.WithCOVIncrement(0.5), // Notify on 0.5 change
)
```

### WithConfirmedNotifications

```go
func WithConfirmedNotifications(confirmed bool) SubscribeOption
```

Requests confirmed COV notifications instead of unconfirmed.

```go
subID, err := client.SubscribeCOV(ctx, deviceID, objectID, handler,
    bacnet.WithConfirmedNotifications(true),
)
```

**Default:** `false` (unconfirmed notifications)

## Complete Configuration Example

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

client, err := bacnet.NewClient(
    // Device identity
    bacnet.WithDeviceID(99999),
    bacnet.WithLocalAddress("0.0.0.0:47808"),

    // Network configuration
    bacnet.WithNetworkNumber(1),
    bacnet.WithBBMD("192.168.1.1", 47808, 60*time.Second),

    // Timeouts and retries
    bacnet.WithTimeout(5 * time.Second),
    bacnet.WithRetries(3),
    bacnet.WithRetryDelay(500 * time.Millisecond),

    // APDU configuration
    bacnet.WithMaxAPDULength(1476),
    bacnet.WithSegmentation(bacnet.SegmentationNone),

    // Auto-discovery
    bacnet.WithAutoDiscover(false),
    bacnet.WithDiscoverTimeout(5 * time.Second),

    // Logging
    bacnet.WithLogger(logger),
)
```

## Default Values Summary

| Option | Default Value |
|--------|---------------|
| DeviceID | `0xFFFFFFFF` |
| NetworkNumber | `0` |
| Timeout | `3s` |
| Retries | `3` |
| RetryDelay | `500ms` |
| MaxAPDULength | `1476` |
| Segmentation | `SegmentationNone` |
| ProposedWindowSize | `1` |
| AutoDiscover | `false` |
| DiscoverTimeout | `5s` |
| Logger | `slog.Default()` |
