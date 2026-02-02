# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-02-01

### Added

#### Core Client
- BACnet/IP client implementation with full UDP support
- BVLC (BACnet Virtual Link Control) encoding/decoding
- NPDU (Network Protocol Data Unit) handling
- APDU (Application Protocol Data Unit) support for all PDU types
- Connection state management (Disconnected, Connecting, Connected)
- Automatic device resolution from discovery cache

#### Device Discovery
- Who-Is/I-Am broadcast discovery mechanism
- Device range filtering (low/high instance limits)
- Target network specification for remote discovery
- Configurable discovery timeout
- Device information caching

#### Property Operations
- ReadProperty service for single property reads
- WriteProperty service with priority support (1-16)
- ReadPropertyMultiple for efficient batch reads
- Array index support for property access
- Automatic value type encoding/decoding

#### COV Subscriptions
- SubscribeCOV for change notifications
- Unsubscribe functionality
- Subscription lifetime management
- COV increment for analog values
- Confirmed and unconfirmed notification support

#### BBMD Support
- Foreign device registration
- Configurable TTL for registrations
- Cross-subnet communication

#### Object Types
- 60+ BACnet object types supported
- Shorthand aliases (ai, ao, av, bi, bo, bv, etc.)
- ParseObjectType helper function
- String representation for all types

#### Property Identifiers
- 150+ property identifiers supported
- Shorthand aliases (pv, name, oos, pa, rd, etc.)
- ParsePropertyIdentifier helper function
- String representation for all properties

#### Error Handling
- Sentinel errors for common conditions
- BACnetError type for protocol errors
- RejectError for request rejections
- AbortError for transaction aborts
- Helper functions: IsTimeout, IsDeviceNotFound, IsPropertyNotFound, IsAccessDenied

#### Metrics
- Connection metrics (attempts, successes, failures)
- Request metrics (sent, succeeded, failed, timed out)
- Response metrics (received, errors, rejects, aborts)
- Discovery metrics (Who-Is sent, I-Am received, devices discovered)
- COV metrics (subscriptions, notifications)
- Latency histogram with 10 buckets
- Bandwidth tracking (bytes sent/received)
- Uptime and last activity timestamps

#### Configuration
- Functional options pattern
- Device ID configuration
- Local address binding
- Network number assignment
- Timeout and retry configuration
- APDU settings (max length, segmentation, window size)
- Auto-discovery option
- Structured logging with slog

#### CLI Tool (edgeo-bacnet)
- `scan` command for device discovery
- `read` command for property reading
- `write` command for property writing
- `watch` command for COV monitoring
- `dump` command for device point export
- `info` command for device information
- `interactive` command for shell mode
- Multiple output formats (table, JSON, CSV, raw)
- Configuration file support (~/.edgeo-bacnet.yaml)
- Environment variable configuration

### Security

- Input validation on all protocol messages
- Timeout protection for all network operations
- Safe concurrent access with mutex protection

---

## Version Access

```go
package main

import "fmt"

const Version = "1.0.0"

func main() {
    fmt.Println("BACnet Driver Version:", Version)
}
```

## Versioning Convention

This project follows semantic versioning:

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

## Upgrade Guide

### From Pre-release to 1.0.0

This is the initial stable release. No migration required.

## Reporting Issues

Report bugs and feature requests at: https://github.com/edgeo/drivers/bacnet/issues
