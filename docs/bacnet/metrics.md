# Metrics & Observability

The BACnet driver provides comprehensive metrics for monitoring client performance and health.

## Accessing Metrics

```go
// Get metrics from client
metrics := client.Metrics()

// Get a point-in-time snapshot
snapshot := metrics.Snapshot()

fmt.Printf("Uptime: %v\n", snapshot.Uptime)
fmt.Printf("Requests sent: %d\n", snapshot.RequestsSent)
fmt.Printf("Requests succeeded: %d\n", snapshot.RequestsSucceeded)
```

## Metric Categories

### Connection Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `ConnectAttempts` | Counter | Total connection attempts |
| `ConnectSuccesses` | Counter | Successful connections |
| `ConnectFailures` | Counter | Failed connections |
| `Disconnects` | Counter | Disconnection events |

```go
fmt.Printf("Connect attempts: %d\n", snapshot.ConnectAttempts)
fmt.Printf("Connect successes: %d\n", snapshot.ConnectSuccesses)
fmt.Printf("Connect failures: %d\n", snapshot.ConnectFailures)
fmt.Printf("Disconnects: %d\n", snapshot.Disconnects)
```

### Request Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `RequestsSent` | Counter | Total requests sent |
| `RequestsSucceeded` | Counter | Successful requests |
| `RequestsFailed` | Counter | Failed requests |
| `RequestsTimedOut` | Counter | Timed out requests |
| `ActiveRequests` | Gauge | Currently active requests |

```go
fmt.Printf("Requests sent: %d\n", snapshot.RequestsSent)
fmt.Printf("Success rate: %.1f%%\n",
    float64(snapshot.RequestsSucceeded)/float64(snapshot.RequestsSent)*100)
fmt.Printf("Timeout rate: %.1f%%\n",
    float64(snapshot.RequestsTimedOut)/float64(snapshot.RequestsSent)*100)
fmt.Printf("Active requests: %d\n", snapshot.ActiveRequests)
```

### Response Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `ResponsesReceived` | Counter | Total responses received |
| `ErrorsReceived` | Counter | Error responses |
| `RejectsReceived` | Counter | Reject responses |
| `AbortsReceived` | Counter | Abort responses |

```go
fmt.Printf("Responses: %d\n", snapshot.ResponsesReceived)
fmt.Printf("Errors: %d\n", snapshot.ErrorsReceived)
fmt.Printf("Rejects: %d\n", snapshot.RejectsReceived)
fmt.Printf("Aborts: %d\n", snapshot.AbortsReceived)
```

### Discovery Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `WhoIsSent` | Counter | Who-Is broadcasts sent |
| `IAmReceived` | Counter | I-Am responses received |
| `DevicesDiscovered` | Counter | Unique devices discovered |

```go
fmt.Printf("Who-Is sent: %d\n", snapshot.WhoIsSent)
fmt.Printf("I-Am received: %d\n", snapshot.IAmReceived)
fmt.Printf("Devices discovered: %d\n", snapshot.DevicesDiscovered)
```

### COV Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `COVSubscriptions` | Counter | COV subscriptions created |
| `COVNotifications` | Counter | COV notifications received |
| `ActiveSubscriptions` | Gauge | Currently active subscriptions |

```go
fmt.Printf("COV subscriptions: %d\n", snapshot.COVSubscriptions)
fmt.Printf("COV notifications: %d\n", snapshot.COVNotifications)
fmt.Printf("Active subscriptions: %d\n", snapshot.ActiveSubscriptions)
```

### Bandwidth Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `BytesSent` | Counter | Total bytes transmitted |
| `BytesReceived` | Counter | Total bytes received |

```go
fmt.Printf("Bytes sent: %d\n", snapshot.BytesSent)
fmt.Printf("Bytes received: %d\n", snapshot.BytesReceived)
fmt.Printf("Total bandwidth: %d bytes\n", snapshot.BytesSent+snapshot.BytesReceived)
```

### Timing Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `Uptime` | Duration | Time since client started |
| `LastActivity` | Time | Timestamp of last activity |

```go
fmt.Printf("Uptime: %v\n", snapshot.Uptime)
fmt.Printf("Last activity: %v\n", snapshot.LastActivity)
fmt.Printf("Idle time: %v\n", time.Since(snapshot.LastActivity))
```

## Latency Histogram

Request latency is tracked in a histogram with predefined buckets.

### Latency Stats

```go
stats := snapshot.LatencyStats

fmt.Printf("Request count: %d\n", stats.Count)
fmt.Printf("Min latency: %v\n", stats.Min)
fmt.Printf("Max latency: %v\n", stats.Max)
fmt.Printf("Avg latency: %v\n", stats.Avg)
```

### Histogram Buckets

| Bucket | Index | Threshold |
|--------|-------|-----------|
| 0 | 0 | < 1ms |
| 1 | 1 | < 5ms |
| 2 | 2 | < 10ms |
| 3 | 3 | < 25ms |
| 4 | 4 | < 50ms |
| 5 | 5 | < 100ms |
| 6 | 6 | < 250ms |
| 7 | 7 | < 500ms |
| 8 | 8 | < 1000ms |
| 9 | 9 | >= 1000ms |

```go
bucketLabels := []string{
    "<1ms", "<5ms", "<10ms", "<25ms", "<50ms",
    "<100ms", "<250ms", "<500ms", "<1s", ">=1s",
}

fmt.Println("Latency distribution:")
for i, count := range stats.Buckets {
    if count > 0 {
        pct := float64(count) / float64(stats.Count) * 100
        fmt.Printf("  %s: %d (%.1f%%)\n", bucketLabels[i], count, pct)
    }
}
```

## Metric Types

### Counter

Thread-safe incrementing counter.

```go
type Counter struct {
    value int64
}

func (c *Counter) Inc()           // Increment by 1
func (c *Counter) Add(delta int64) // Add arbitrary value
func (c *Counter) Value() int64   // Get current value
func (c *Counter) Reset()         // Reset to 0
```

### Gauge

Thread-safe gauge that can increase or decrease.

```go
type Gauge struct {
    value int64
}

func (g *Gauge) Set(value int64)   // Set to specific value
func (g *Gauge) Inc()              // Increment by 1
func (g *Gauge) Dec()              // Decrement by 1
func (g *Gauge) Add(delta int64)   // Add (positive or negative)
func (g *Gauge) Value() int64      // Get current value
```

### LatencyHistogram

Thread-safe histogram for latency measurements.

```go
type LatencyHistogram struct {
    // internal fields
}

func (h *LatencyHistogram) Record(d time.Duration)  // Record measurement
func (h *LatencyHistogram) Stats() LatencyStats     // Get statistics
func (h *LatencyHistogram) Reset()                  // Reset all data
```

## Prometheus Integration

Export metrics to Prometheus:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    bacnetRequests = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "bacnet_requests_total",
            Help: "Total BACnet requests",
        },
        []string{"status"},
    )

    bacnetLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "bacnet_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"operation"},
    )

    bacnetDevices = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "bacnet_discovered_devices",
            Help: "Number of discovered devices",
        },
    )
)

// Periodic export
func exportMetrics(client *bacnet.Client) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    var lastSnapshot bacnet.MetricsSnapshot

    for range ticker.C {
        snapshot := client.Metrics().Snapshot()

        // Calculate deltas
        succeeded := snapshot.RequestsSucceeded - lastSnapshot.RequestsSucceeded
        failed := snapshot.RequestsFailed - lastSnapshot.RequestsFailed

        bacnetRequests.WithLabelValues("success").Add(float64(succeeded))
        bacnetRequests.WithLabelValues("failure").Add(float64(failed))

        bacnetDevices.Set(float64(snapshot.DevicesDiscovered))

        lastSnapshot = snapshot
    }
}
```

## expvar Integration

Export metrics via expvar for standard Go tooling:

```go
import "expvar"

func init() {
    expvar.Publish("bacnet", expvar.Func(func() interface{} {
        return client.Metrics().Snapshot()
    }))
}
```

Access at `http://localhost:8080/debug/vars`

## Resetting Metrics

```go
// Reset all metrics
client.Metrics().Reset()
```

**Note:** Reset clears all counters and the latency histogram. Uptime restarts from the reset time.

## Metrics Logging

```go
func logMetrics(logger *slog.Logger, client *bacnet.Client, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for range ticker.C {
        snapshot := client.Metrics().Snapshot()

        logger.Info("BACnet metrics",
            slog.Duration("uptime", snapshot.Uptime),
            slog.Int64("requests_sent", snapshot.RequestsSent),
            slog.Int64("requests_succeeded", snapshot.RequestsSucceeded),
            slog.Int64("requests_failed", snapshot.RequestsFailed),
            slog.Int64("requests_timed_out", snapshot.RequestsTimedOut),
            slog.Int64("devices_discovered", snapshot.DevicesDiscovered),
            slog.Duration("avg_latency", snapshot.LatencyStats.Avg),
            slog.Int64("bytes_sent", snapshot.BytesSent),
            slog.Int64("bytes_received", snapshot.BytesReceived),
        )
    }
}
```

## Health Checks

```go
func healthCheck(client *bacnet.Client) error {
    snapshot := client.Metrics().Snapshot()

    // Check if client is connected
    if client.State() != bacnet.StateConnected {
        return fmt.Errorf("client not connected")
    }

    // Check for recent activity
    idleTime := time.Since(snapshot.LastActivity)
    if idleTime > 5*time.Minute {
        return fmt.Errorf("no activity for %v", idleTime)
    }

    // Check error rate
    if snapshot.RequestsSent > 0 {
        errorRate := float64(snapshot.RequestsFailed) / float64(snapshot.RequestsSent)
        if errorRate > 0.1 { // 10% error rate
            return fmt.Errorf("high error rate: %.1f%%", errorRate*100)
        }
    }

    return nil
}
```

## CLI Metrics

```bash
# Get device info with metrics
edgeo-bacnet info -d 1234

# Verbose output includes operation metrics
edgeo-bacnet read -d 1234 -o ai:1 -p pv -v
```
