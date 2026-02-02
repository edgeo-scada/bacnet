# COV Subscriptions

Change of Value (COV) subscriptions provide real-time notifications when property values change.

## How COV Works

1. Client sends a **SubscribeCOV** request to a device
2. Device accepts the subscription and assigns a subscription ID
3. When the monitored property changes, the device sends a **COVNotification**
4. Client receives and processes the notification via a callback handler
5. Subscription remains active until cancelled or lifetime expires

## Basic COV Subscription

```go
ctx := context.Background()
deviceID := uint32(1234)
objectID := bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1)

// Define handler for COV notifications
handler := func(deviceID uint32, objectID bacnet.ObjectIdentifier, values []bacnet.PropertyValue) {
    fmt.Printf("COV notification from device %d, object %s:\n", deviceID, objectID)
    for _, pv := range values {
        fmt.Printf("  %s = %v\n", pv.PropertyID, pv.Value)
    }
}

// Subscribe to COV
subID, err := client.SubscribeCOV(ctx, deviceID, objectID, handler)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Subscribed with ID: %d\n", subID)

// Later: unsubscribe
err = client.UnsubscribeCOV(ctx, deviceID, objectID, subID)
```

## Subscription Options

### WithSubscriptionLifetime

Sets how long the subscription remains active (in seconds).

```go
// Subscription expires after 5 minutes
subID, err := client.SubscribeCOV(ctx, deviceID, objectID, handler,
    bacnet.WithSubscriptionLifetime(300),
)
```

**Note:** Without a lifetime, the subscription remains active until cancelled.

### WithCOVIncrement

For analog values, sets the minimum change required to trigger a notification.

```go
// Notify when value changes by at least 0.5
subID, err := client.SubscribeCOV(ctx, deviceID, objectID, handler,
    bacnet.WithCOVIncrement(0.5),
)
```

**Note:** Only applicable to analog objects. Binary objects notify on any state change.

### WithConfirmedNotifications

Request confirmed notifications instead of unconfirmed.

```go
// Use confirmed notifications
subID, err := client.SubscribeCOV(ctx, deviceID, objectID, handler,
    bacnet.WithConfirmedNotifications(true),
)
```

**Differences:**
- **Unconfirmed (default):** Device sends notification without expecting acknowledgment
- **Confirmed:** Device expects acknowledgment; retries if not received

## COV Handler

The handler function receives:

- `deviceID`: Device that sent the notification
- `objectID`: Object that changed
- `values`: Slice of property values (typically present-value and status-flags)

```go
handler := func(deviceID uint32, objectID bacnet.ObjectIdentifier, values []bacnet.PropertyValue) {
    for _, pv := range values {
        switch pv.PropertyID {
        case bacnet.PropertyPresentValue:
            fmt.Printf("New value: %v\n", pv.Value)
        case bacnet.PropertyStatusFlags:
            // Status flags indicate alarm, fault, overridden, out-of-service
            if flags, ok := pv.Value.(byte); ok {
                sf := bacnet.DecodeStatusFlags(flags)
                fmt.Printf("Status: %v\n", sf)
            }
        }
    }
}
```

## Unsubscribing

```go
// Unsubscribe using the subscription ID
err := client.UnsubscribeCOV(ctx, deviceID, objectID, subID)
if err != nil {
    log.Printf("Unsubscribe failed: %v", err)
}
```

## COV Patterns

### Multiple Subscriptions

```go
type subscription struct {
    deviceID uint32
    objectID bacnet.ObjectIdentifier
    subID    uint32
}

var subscriptions []subscription

// Subscribe to multiple objects
objects := []bacnet.ObjectIdentifier{
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 2),
    bacnet.NewObjectIdentifier(bacnet.ObjectTypeBinaryInput, 1),
}

for _, obj := range objects {
    subID, err := client.SubscribeCOV(ctx, deviceID, obj, handler,
        bacnet.WithSubscriptionLifetime(300),
    )
    if err != nil {
        log.Printf("Failed to subscribe to %s: %v", obj, err)
        continue
    }
    subscriptions = append(subscriptions, subscription{
        deviceID: deviceID,
        objectID: obj,
        subID:    subID,
    })
}

// Cleanup: unsubscribe all
for _, sub := range subscriptions {
    client.UnsubscribeCOV(ctx, sub.deviceID, sub.objectID, sub.subID)
}
```

### Subscription Renewal

```go
func renewSubscription(ctx context.Context, client *bacnet.Client, deviceID uint32, objectID bacnet.ObjectIdentifier, handler bacnet.COVHandler) {
    lifetime := uint32(300) // 5 minutes
    renewBefore := 60 * time.Second

    for {
        subID, err := client.SubscribeCOV(ctx, deviceID, objectID, handler,
            bacnet.WithSubscriptionLifetime(lifetime),
        )
        if err != nil {
            log.Printf("Subscription failed: %v", err)
            time.Sleep(10 * time.Second)
            continue
        }

        // Wait until near expiration
        select {
        case <-ctx.Done():
            client.UnsubscribeCOV(ctx, deviceID, objectID, subID)
            return
        case <-time.After(time.Duration(lifetime)*time.Second - renewBefore):
            // Time to renew - resubscription will replace the old subscription
        }
    }
}
```

### COV with Error Handling

```go
handler := func(deviceID uint32, objectID bacnet.ObjectIdentifier, values []bacnet.PropertyValue) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Handler panic: %v", r)
        }
    }()

    // Process values safely
    for _, pv := range values {
        if pv.Value == nil {
            continue
        }

        switch v := pv.Value.(type) {
        case float32:
            fmt.Printf("%s.%s = %.2f\n", objectID, pv.PropertyID, v)
        case float64:
            fmt.Printf("%s.%s = %.2f\n", objectID, pv.PropertyID, v)
        case bool:
            fmt.Printf("%s.%s = %v\n", objectID, pv.PropertyID, v)
        default:
            fmt.Printf("%s.%s = %v\n", objectID, pv.PropertyID, v)
        }
    }
}
```

### Channel-Based Notifications

```go
type COVEvent struct {
    DeviceID uint32
    ObjectID bacnet.ObjectIdentifier
    Values   []bacnet.PropertyValue
}

func subscribeWithChannel(ctx context.Context, client *bacnet.Client, deviceID uint32, objectID bacnet.ObjectIdentifier) (<-chan COVEvent, func(), error) {
    events := make(chan COVEvent, 100)

    handler := func(devID uint32, objID bacnet.ObjectIdentifier, values []bacnet.PropertyValue) {
        select {
        case events <- COVEvent{
            DeviceID: devID,
            ObjectID: objID,
            Values:   values,
        }:
        default:
            log.Println("COV event channel full, dropping event")
        }
    }

    subID, err := client.SubscribeCOV(ctx, deviceID, objectID, handler)
    if err != nil {
        close(events)
        return nil, nil, err
    }

    cancel := func() {
        client.UnsubscribeCOV(ctx, deviceID, objectID, subID)
        close(events)
    }

    return events, cancel, nil
}

// Usage
events, cancel, err := subscribeWithChannel(ctx, client, deviceID, objectID)
if err != nil {
    log.Fatal(err)
}
defer cancel()

for event := range events {
    fmt.Printf("Received COV: %v\n", event)
}
```

## COV vs Polling

| Aspect | COV | Polling |
|--------|-----|---------|
| Network traffic | Low (only on change) | Constant |
| Latency | Immediate | Poll interval |
| Device support | Required | Universal |
| Scalability | Excellent | Limited |
| Battery impact | Minimal | Significant |

**Use COV when:**
- Real-time updates are needed
- Monitoring many values
- Network bandwidth is limited
- Device supports COV

**Use Polling when:**
- Device doesn't support COV
- Periodic samples are sufficient
- COV increment can't capture needed precision

## Supported Object Types

COV is commonly supported on:

- Analog Input/Output/Value
- Binary Input/Output/Value
- Multi-State Input/Output/Value
- Loop (setpoint, controlled variable)

**Note:** Not all devices support COV on all object types. Check `not-cov-property` error.

## CLI COV Monitoring

```bash
# Watch for changes
edgeo-bacnet watch -d 1234 -o ai:1

# Watch multiple objects
edgeo-bacnet watch -d 1234 -o ai:1,ai:2,bi:1

# With timeout
edgeo-bacnet watch -d 1234 -o ai:1 --duration 5m
```
