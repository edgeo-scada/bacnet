// Copyright 2025 Edgeo SCADA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bacnet

import (
	"log/slog"
	"time"
)

// ClientOptions holds configuration for the BACnet client
type clientOptions struct {
	// Device configuration
	localDeviceID uint32
	localAddress  string

	// Network configuration
	networkNumber uint16
	bbmdAddress   string
	bbmdPort      int
	foreignDeviceTTL time.Duration

	// Timeouts
	timeout        time.Duration
	retries        int
	retryDelay     time.Duration

	// APDU configuration
	maxAPDULength  uint16
	segmentation   Segmentation
	proposedWindowSize uint8

	// Auto-discovery
	autoDiscover   bool
	discoverTimeout time.Duration

	// Logging
	logger         *slog.Logger
}

// defaultOptions returns the default client options
func defaultOptions() *clientOptions {
	return &clientOptions{
		localDeviceID:     0xFFFFFFFF, // Uninitialized
		networkNumber:     0,
		timeout:           3 * time.Second,
		retries:           3,
		retryDelay:        500 * time.Millisecond,
		maxAPDULength:     MaxAPDULength,
		segmentation:      SegmentationNone,
		proposedWindowSize: 1,
		autoDiscover:      false,
		discoverTimeout:   5 * time.Second,
		logger:            slog.Default(),
	}
}

// Option is a functional option for configuring the client
type Option func(*clientOptions)

// WithDeviceID sets the local device ID for the client
func WithDeviceID(id uint32) Option {
	return func(o *clientOptions) {
		o.localDeviceID = id
	}
}

// WithLocalAddress sets the local address to bind to
func WithLocalAddress(addr string) Option {
	return func(o *clientOptions) {
		o.localAddress = addr
	}
}

// WithNetworkNumber sets the BACnet network number
func WithNetworkNumber(net uint16) Option {
	return func(o *clientOptions) {
		o.networkNumber = net
	}
}

// WithBBMD sets the BBMD (BACnet Broadcast Management Device) address for foreign device registration
func WithBBMD(addr string, port int, ttl time.Duration) Option {
	return func(o *clientOptions) {
		o.bbmdAddress = addr
		o.bbmdPort = port
		o.foreignDeviceTTL = ttl
	}
}

// WithTimeout sets the request timeout
func WithTimeout(d time.Duration) Option {
	return func(o *clientOptions) {
		o.timeout = d
	}
}

// WithRetries sets the number of retries for failed requests
func WithRetries(n int) Option {
	return func(o *clientOptions) {
		o.retries = n
	}
}

// WithRetryDelay sets the delay between retries
func WithRetryDelay(d time.Duration) Option {
	return func(o *clientOptions) {
		o.retryDelay = d
	}
}

// WithMaxAPDULength sets the maximum APDU length
func WithMaxAPDULength(length uint16) Option {
	return func(o *clientOptions) {
		o.maxAPDULength = length
	}
}

// WithSegmentation sets the segmentation capability
func WithSegmentation(seg Segmentation) Option {
	return func(o *clientOptions) {
		o.segmentation = seg
	}
}

// WithProposedWindowSize sets the proposed window size for segmentation
func WithProposedWindowSize(size uint8) Option {
	return func(o *clientOptions) {
		o.proposedWindowSize = size
	}
}

// WithAutoDiscover enables automatic device discovery
func WithAutoDiscover(enable bool) Option {
	return func(o *clientOptions) {
		o.autoDiscover = enable
	}
}

// WithDiscoverTimeout sets the timeout for device discovery
func WithDiscoverTimeout(d time.Duration) Option {
	return func(o *clientOptions) {
		o.discoverTimeout = d
	}
}

// WithLogger sets the logger for the client
func WithLogger(logger *slog.Logger) Option {
	return func(o *clientOptions) {
		o.logger = logger
	}
}

// DiscoverOptions holds configuration for device discovery
type DiscoverOptions struct {
	// Range limits for WhoIs
	LowLimit  *uint32
	HighLimit *uint32

	// Timeout for discovery
	Timeout time.Duration

	// Network to search (0 = local)
	Network uint16
}

// DiscoverOption is a functional option for discovery
type DiscoverOption func(*DiscoverOptions)

// defaultDiscoverOptions returns default discovery options
func defaultDiscoverOptions() *DiscoverOptions {
	return &DiscoverOptions{
		Timeout: 5 * time.Second,
		Network: 0,
	}
}

// WithDeviceRange sets the device ID range for discovery
func WithDeviceRange(low, high uint32) DiscoverOption {
	return func(o *DiscoverOptions) {
		o.LowLimit = &low
		o.HighLimit = &high
	}
}

// WithDiscoveryTimeout sets the discovery timeout
func WithDiscoveryTimeout(d time.Duration) DiscoverOption {
	return func(o *DiscoverOptions) {
		o.Timeout = d
	}
}

// WithTargetNetwork sets the target network for discovery
func WithTargetNetwork(net uint16) DiscoverOption {
	return func(o *DiscoverOptions) {
		o.Network = net
	}
}

// ReadOptions holds configuration for read operations
type ReadOptions struct {
	ArrayIndex *uint32
}

// ReadOption is a functional option for read operations
type ReadOption func(*ReadOptions)

// WithArrayIndex sets the array index for reading array properties
func WithArrayIndex(index uint32) ReadOption {
	return func(o *ReadOptions) {
		o.ArrayIndex = &index
	}
}

// WriteOptions holds configuration for write operations
type WriteOptions struct {
	ArrayIndex *uint32
	Priority   *uint8
}

// WriteOption is a functional option for write operations
type WriteOption func(*WriteOptions)

// WithWriteArrayIndex sets the array index for writing array properties
func WithWriteArrayIndex(index uint32) WriteOption {
	return func(o *WriteOptions) {
		o.ArrayIndex = &index
	}
}

// WithPriority sets the priority for writing (1-16, where 1 is highest)
func WithPriority(priority uint8) WriteOption {
	return func(o *WriteOptions) {
		if priority >= 1 && priority <= 16 {
			o.Priority = &priority
		}
	}
}

// SubscribeOptions holds configuration for COV subscriptions
type SubscribeOptions struct {
	Lifetime     *uint32
	COVIncrement *float32
	Confirmed    bool
}

// SubscribeOption is a functional option for COV subscriptions
type SubscribeOption func(*SubscribeOptions)

// WithSubscriptionLifetime sets the subscription lifetime in seconds
func WithSubscriptionLifetime(seconds uint32) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Lifetime = &seconds
	}
}

// WithCOVIncrement sets the COV increment for analog values
func WithCOVIncrement(increment float32) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.COVIncrement = &increment
	}
}

// WithConfirmedNotifications requests confirmed COV notifications
func WithConfirmedNotifications(confirmed bool) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Confirmed = confirmed
	}
}
