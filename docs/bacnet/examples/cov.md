# COV Subscription Example

This example demonstrates Change of Value (COV) subscriptions for real-time property monitoring.

## Basic COV Subscription

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "time"

    "github.com/edgeo/drivers/bacnet/bacnet"
)

func main() {
    client := bacnet.NewClient(
        bacnet.WithTimeout(3*time.Second),
    )

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    deviceID := uint32(1234)

    // COV notification handler
    handler := func(n *bacnet.COVNotification) {
        log.Printf("COV: Device %d, %s.%s = %v (flags: %v)",
            n.DeviceID,
            n.Object,
            n.Property,
            n.Value,
            n.StatusFlags)
    }

    // Subscribe to AI:0
    subID, err := client.SubscribeCOV(ctx,
        deviceID,
        bacnet.ObjectAnalogInput(0),
        handler,
        bacnet.WithSubscriptionLifetime(300), // 5 minutes
    )
    if err != nil {
        log.Fatalf("Subscribe error: %v", err)
    }
    log.Printf("Subscribed to AI:0, subscription ID: %d", subID)

    // Wait for interrupt
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt)
    <-sigCh

    // Unsubscribe
    log.Println("Unsubscribing...")
    if err := client.UnsubscribeCOV(ctx, deviceID, bacnet.ObjectAnalogInput(0), subID); err != nil {
        log.Printf("Unsubscribe error: %v", err)
    }
}
```

## Multiple Subscriptions

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "sync"
    "time"

    "github.com/edgeo/drivers/bacnet/bacnet"
)

type Subscription struct {
    DeviceID uint32
    Object   bacnet.ObjectIdentifier
    SubID    uint32
}

func main() {
    client := bacnet.NewClient(
        bacnet.WithTimeout(3*time.Second),
    )

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    deviceID := uint32(1234)
    var subscriptions []Subscription
    var mu sync.Mutex

    // Handler with object context
    handler := func(n *bacnet.COVNotification) {
        log.Printf("[%s] %s = %v",
            time.Now().Format("15:04:05.000"),
            n.Object,
            n.Value)
    }

    // Objects to subscribe
    objects := []bacnet.ObjectIdentifier{
        bacnet.ObjectAnalogInput(0),
        bacnet.ObjectAnalogInput(1),
        bacnet.ObjectAnalogInput(2),
        bacnet.ObjectBinaryInput(0),
        bacnet.ObjectBinaryInput(1),
    }

    // Subscribe to all objects
    for _, obj := range objects {
        subID, err := client.SubscribeCOV(ctx, deviceID, obj, handler,
            bacnet.WithSubscriptionLifetime(600), // 10 minutes
        )
        if err != nil {
            log.Printf("Subscribe %s error: %v", obj, err)
            continue
        }

        mu.Lock()
        subscriptions = append(subscriptions, Subscription{
            DeviceID: deviceID,
            Object:   obj,
            SubID:    subID,
        })
        mu.Unlock()

        log.Printf("Subscribed to %s", obj)
    }

    log.Printf("Monitoring %d objects...", len(subscriptions))

    // Wait for interrupt
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt)
    <-sigCh

    // Unsubscribe all
    log.Println("Cleaning up subscriptions...")
    for _, sub := range subscriptions {
        if err := client.UnsubscribeCOV(ctx, sub.DeviceID, sub.Object, sub.SubID); err != nil {
            log.Printf("Unsubscribe %s error: %v", sub.Object, err)
        }
    }
}
```

## COV with Increment Threshold

For analog objects, you can set a minimum change threshold:

```go
// Only notify if value changes by at least 0.5
subID, err := client.SubscribeCOV(ctx,
    deviceID,
    bacnet.ObjectAnalogInput(0),
    handler,
    bacnet.WithSubscriptionLifetime(300),
    bacnet.WithCOVIncrement(0.5), // Trigger on 0.5 change
)
```

## COV with Confirmed Notifications

Request confirmed (acknowledged) notifications:

```go
subID, err := client.SubscribeCOV(ctx,
    deviceID,
    bacnet.ObjectAnalogInput(0),
    handler,
    bacnet.WithSubscriptionLifetime(300),
    bacnet.WithConfirmedNotifications(true),
)
```

## Automatic Resubscription

Resubscribe before expiration:

```go
func maintainSubscription(ctx context.Context, client *bacnet.Client, deviceID uint32, object bacnet.ObjectIdentifier, handler bacnet.COVHandler) {
    lifetime := 300 // 5 minutes
    renewBefore := 60 // Renew 1 minute before expiration

    for {
        // Subscribe
        subID, err := client.SubscribeCOV(ctx, deviceID, object, handler,
            bacnet.WithSubscriptionLifetime(uint32(lifetime)),
        )
        if err != nil {
            log.Printf("Subscribe error: %v", err)
            time.Sleep(10 * time.Second)
            continue
        }

        log.Printf("Subscribed to %s, ID: %d", object, subID)

        // Wait until renewal time
        select {
        case <-ctx.Done():
            client.UnsubscribeCOV(ctx, deviceID, object, subID)
            return
        case <-time.After(time.Duration(lifetime-renewBefore) * time.Second):
            // Time to renew - unsubscribe first
            client.UnsubscribeCOV(ctx, deviceID, object, subID)
        }
    }
}

// Usage
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go maintainSubscription(ctx, client, 1234, bacnet.ObjectAnalogInput(0), handler)
```

## Processing COV Data

```go
type SensorData struct {
    Timestamp time.Time
    Object    string
    Value     float64
    InAlarm   bool
    Fault     bool
}

var dataChannel = make(chan SensorData, 100)

func covHandler(n *bacnet.COVNotification) {
    data := SensorData{
        Timestamp: n.Timestamp,
        Object:    n.Object.String(),
    }

    // Parse value
    if v, ok := n.Value.(float64); ok {
        data.Value = v
    } else if v, ok := n.Value.(float32); ok {
        data.Value = float64(v)
    }

    // Parse status flags
    if n.StatusFlags != nil {
        data.InAlarm = n.StatusFlags.InAlarm
        data.Fault = n.StatusFlags.Fault
    }

    select {
    case dataChannel <- data:
    default:
        log.Println("Data channel full, dropping notification")
    }
}

func processData() {
    for data := range dataChannel {
        // Store in database, send to MQTT, etc.
        log.Printf("Processing: %s = %.2f at %v",
            data.Object, data.Value, data.Timestamp)
    }
}
```

## Best Practices

1. **Set appropriate lifetime** - Balance between network traffic and reliability
2. **Handle resubscription** - Subscriptions expire; plan for renewal
3. **Use increment threshold** - Reduce notifications for analog values
4. **Unsubscribe on shutdown** - Clean up subscriptions properly
5. **Buffer notifications** - Use channels to handle bursts
6. **Monitor subscription status** - Log subscription failures
