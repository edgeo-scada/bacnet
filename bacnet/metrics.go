package bacnet

import (
	"sync"
	"sync/atomic"
	"time"
)

// Counter is a thread-safe counter
type Counter struct {
	value int64
}

// Add adds a delta to the counter
func (c *Counter) Add(delta int64) {
	atomic.AddInt64(&c.value, delta)
}

// Inc increments the counter by 1
func (c *Counter) Inc() {
	c.Add(1)
}

// Value returns the current counter value
func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

// Reset resets the counter to 0
func (c *Counter) Reset() {
	atomic.StoreInt64(&c.value, 0)
}

// Gauge is a thread-safe gauge that can go up and down
type Gauge struct {
	value int64
}

// Set sets the gauge value
func (g *Gauge) Set(value int64) {
	atomic.StoreInt64(&g.value, value)
}

// Add adds a delta to the gauge
func (g *Gauge) Add(delta int64) {
	atomic.AddInt64(&g.value, delta)
}

// Inc increments the gauge by 1
func (g *Gauge) Inc() {
	g.Add(1)
}

// Dec decrements the gauge by 1
func (g *Gauge) Dec() {
	g.Add(-1)
}

// Value returns the current gauge value
func (g *Gauge) Value() int64 {
	return atomic.LoadInt64(&g.value)
}

// LatencyHistogram tracks latency measurements
type LatencyHistogram struct {
	mu      sync.RWMutex
	count   int64
	sum     int64 // nanoseconds
	min     int64
	max     int64
	buckets []int64 // counts for each bucket
}

// NewLatencyHistogram creates a new latency histogram
func NewLatencyHistogram() *LatencyHistogram {
	return &LatencyHistogram{
		min:     -1, // Indicates no measurements yet
		buckets: make([]int64, 10), // <1ms, <5ms, <10ms, <25ms, <50ms, <100ms, <250ms, <500ms, <1s, >=1s
	}
}

// Record records a latency measurement
func (h *LatencyHistogram) Record(d time.Duration) {
	ns := d.Nanoseconds()

	h.mu.Lock()
	defer h.mu.Unlock()

	h.count++
	h.sum += ns

	if h.min < 0 || ns < h.min {
		h.min = ns
	}
	if ns > h.max {
		h.max = ns
	}

	// Update bucket
	ms := d.Milliseconds()
	switch {
	case ms < 1:
		h.buckets[0]++
	case ms < 5:
		h.buckets[1]++
	case ms < 10:
		h.buckets[2]++
	case ms < 25:
		h.buckets[3]++
	case ms < 50:
		h.buckets[4]++
	case ms < 100:
		h.buckets[5]++
	case ms < 250:
		h.buckets[6]++
	case ms < 500:
		h.buckets[7]++
	case ms < 1000:
		h.buckets[8]++
	default:
		h.buckets[9]++
	}
}

// Stats returns histogram statistics
func (h *LatencyHistogram) Stats() LatencyStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := LatencyStats{
		Count:   h.count,
		Buckets: make([]int64, len(h.buckets)),
	}
	copy(stats.Buckets, h.buckets)

	if h.count > 0 {
		stats.Min = time.Duration(h.min)
		stats.Max = time.Duration(h.max)
		stats.Avg = time.Duration(h.sum / h.count)
	}

	return stats
}

// Reset resets the histogram
func (h *LatencyHistogram) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.count = 0
	h.sum = 0
	h.min = -1
	h.max = 0
	for i := range h.buckets {
		h.buckets[i] = 0
	}
}

// LatencyStats contains latency statistics
type LatencyStats struct {
	Count   int64
	Min     time.Duration
	Max     time.Duration
	Avg     time.Duration
	Buckets []int64
}

// Metrics holds client metrics
type Metrics struct {
	// Connection metrics
	ConnectAttempts  Counter
	ConnectSuccesses Counter
	ConnectFailures  Counter
	Disconnects      Counter

	// Request metrics
	RequestsSent     Counter
	RequestsSucceeded Counter
	RequestsFailed   Counter
	RequestsTimedOut Counter

	// Response metrics
	ResponsesReceived Counter
	ErrorsReceived   Counter
	RejectsReceived  Counter
	AbortsReceived   Counter

	// Discovery metrics
	WhoIsSent        Counter
	IAmReceived      Counter
	DevicesDiscovered Counter

	// COV metrics
	COVSubscriptions Counter
	COVNotifications Counter

	// Latency
	RequestLatency *LatencyHistogram

	// Bytes
	BytesSent     Counter
	BytesReceived Counter

	// Current state
	ActiveRequests Gauge
	ActiveSubscriptions Gauge

	// Timestamps
	startTime     time.Time
	lastActivity  atomic.Int64
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		RequestLatency: NewLatencyHistogram(),
		startTime:      time.Now(),
	}
}

// RecordActivity records the last activity time
func (m *Metrics) RecordActivity() {
	m.lastActivity.Store(time.Now().UnixNano())
}

// LastActivity returns the last activity time
func (m *Metrics) LastActivity() time.Time {
	ns := m.lastActivity.Load()
	if ns == 0 {
		return m.startTime
	}
	return time.Unix(0, ns)
}

// Uptime returns the time since metrics started
func (m *Metrics) Uptime() time.Duration {
	return time.Since(m.startTime)
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.ConnectAttempts.Reset()
	m.ConnectSuccesses.Reset()
	m.ConnectFailures.Reset()
	m.Disconnects.Reset()
	m.RequestsSent.Reset()
	m.RequestsSucceeded.Reset()
	m.RequestsFailed.Reset()
	m.RequestsTimedOut.Reset()
	m.ResponsesReceived.Reset()
	m.ErrorsReceived.Reset()
	m.RejectsReceived.Reset()
	m.AbortsReceived.Reset()
	m.WhoIsSent.Reset()
	m.IAmReceived.Reset()
	m.DevicesDiscovered.Reset()
	m.COVSubscriptions.Reset()
	m.COVNotifications.Reset()
	m.RequestLatency.Reset()
	m.BytesSent.Reset()
	m.BytesReceived.Reset()
	m.ActiveRequests.Set(0)
	m.ActiveSubscriptions.Set(0)
	m.startTime = time.Now()
	m.lastActivity.Store(0)
}

// Snapshot returns a snapshot of current metrics
func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		Uptime: m.Uptime(),

		ConnectAttempts:  m.ConnectAttempts.Value(),
		ConnectSuccesses: m.ConnectSuccesses.Value(),
		ConnectFailures:  m.ConnectFailures.Value(),
		Disconnects:      m.Disconnects.Value(),

		RequestsSent:      m.RequestsSent.Value(),
		RequestsSucceeded: m.RequestsSucceeded.Value(),
		RequestsFailed:    m.RequestsFailed.Value(),
		RequestsTimedOut:  m.RequestsTimedOut.Value(),

		ResponsesReceived: m.ResponsesReceived.Value(),
		ErrorsReceived:    m.ErrorsReceived.Value(),
		RejectsReceived:   m.RejectsReceived.Value(),
		AbortsReceived:    m.AbortsReceived.Value(),

		WhoIsSent:         m.WhoIsSent.Value(),
		IAmReceived:       m.IAmReceived.Value(),
		DevicesDiscovered: m.DevicesDiscovered.Value(),

		COVSubscriptions: m.COVSubscriptions.Value(),
		COVNotifications: m.COVNotifications.Value(),

		LatencyStats: m.RequestLatency.Stats(),

		BytesSent:     m.BytesSent.Value(),
		BytesReceived: m.BytesReceived.Value(),

		ActiveRequests:      m.ActiveRequests.Value(),
		ActiveSubscriptions: m.ActiveSubscriptions.Value(),

		LastActivity: m.LastActivity(),
	}
}

// MetricsSnapshot is a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	Uptime time.Duration

	ConnectAttempts  int64
	ConnectSuccesses int64
	ConnectFailures  int64
	Disconnects      int64

	RequestsSent      int64
	RequestsSucceeded int64
	RequestsFailed    int64
	RequestsTimedOut  int64

	ResponsesReceived int64
	ErrorsReceived    int64
	RejectsReceived   int64
	AbortsReceived    int64

	WhoIsSent         int64
	IAmReceived       int64
	DevicesDiscovered int64

	COVSubscriptions int64
	COVNotifications int64

	LatencyStats LatencyStats

	BytesSent     int64
	BytesReceived int64

	ActiveRequests      int64
	ActiveSubscriptions int64

	LastActivity time.Time
}
