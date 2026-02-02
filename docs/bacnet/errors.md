# Error Handling

The BACnet driver provides comprehensive error handling with sentinel errors, typed errors, and helper functions.

## Sentinel Errors

These are predefined errors that can be checked using `errors.Is()`.

```go
import "errors"

// Check for specific error
if errors.Is(err, bacnet.ErrTimeout) {
    // Handle timeout
}
```

### Connection Errors

| Error | Description |
|-------|-------------|
| `ErrNotConnected` | Client is not connected |
| `ErrAlreadyConnected` | Client is already connected |
| `ErrConnectionClosed` | Connection was closed unexpectedly |
| `ErrTimeout` | Request timed out |

### Protocol Errors

| Error | Description |
|-------|-------------|
| `ErrInvalidResponse` | Response could not be decoded |
| `ErrInvalidAPDU` | Invalid APDU format |
| `ErrInvalidNPDU` | Invalid NPDU format |
| `ErrInvalidBVLC` | Invalid BVLC header |
| `ErrSegmentationNotSupported` | Device doesn't support required segmentation |

### Operation Errors

| Error | Description |
|-------|-------------|
| `ErrDeviceNotFound` | Device not found on network |
| `ErrPropertyNotFound` | Property does not exist |
| `ErrWriteFailed` | Write operation failed |

## BACnet Protocol Errors

BACnet devices return errors with a class and code. These are wrapped in `BACnetError`.

```go
var bacnetErr *bacnet.BACnetError
if errors.As(err, &bacnetErr) {
    fmt.Printf("Error class: %s\n", bacnetErr.Class)
    fmt.Printf("Error code: %s\n", bacnetErr.Code)
}
```

### Error Classes

| Class | Description |
|-------|-------------|
| `ErrorClassDevice` | Device-related errors |
| `ErrorClassObject` | Object-related errors |
| `ErrorClassProperty` | Property-related errors |
| `ErrorClassResources` | Resource constraint errors |
| `ErrorClassSecurity` | Security/authentication errors |
| `ErrorClassServices` | Service-related errors |
| `ErrorClassVT` | Virtual terminal errors |
| `ErrorClassCommunication` | Communication errors |

### Common Error Codes

#### Device Errors

| Code | Description |
|------|-------------|
| `ErrorCodeConfigurationInProgress` | Device is being configured |
| `ErrorCodeDeviceBusy` | Device is busy |

#### Object Errors

| Code | Description |
|------|-------------|
| `ErrorCodeUnknownObject` | Object does not exist |
| `ErrorCodeNoObjectsOfSpecifiedType` | No objects of requested type |
| `ErrorCodeObjectDeletionNotPermitted` | Cannot delete object |

#### Property Errors

| Code | Description |
|------|-------------|
| `ErrorCodeUnknownProperty` | Property does not exist |
| `ErrorCodeReadAccessDenied` | Read access denied |
| `ErrorCodeWriteAccessDenied` | Write access denied |
| `ErrorCodeInvalidArrayIndex` | Array index out of bounds |
| `ErrorCodeInvalidDataType` | Wrong data type for property |
| `ErrorCodeValueOutOfRange` | Value outside allowed range |
| `ErrorCodeDatatypeNotSupported` | Data type not supported |
| `ErrorCodeNotCovProperty` | Property doesn't support COV |

#### Resource Errors

| Code | Description |
|------|-------------|
| `ErrorCodeNoSpaceForObject` | No space to create object |
| `ErrorCodeNoSpaceToWriteProperty` | No space to write property |

#### Security Errors

| Code | Description |
|------|-------------|
| `ErrorCodeAuthenticationFailed` | Authentication failed |
| `ErrorCodePasswordFailure` | Incorrect password |
| `ErrorCodeSecurityNotSupported` | Security not supported |

## Reject Errors

Devices may reject requests. These are wrapped in `RejectError`.

```go
var rejectErr *bacnet.RejectError
if errors.As(err, &rejectErr) {
    fmt.Printf("Invoke ID: %d\n", rejectErr.InvokeID)
    fmt.Printf("Reason: %s\n", rejectErr.Reason)
}
```

### Reject Reasons

| Reason | Description |
|--------|-------------|
| `RejectReasonBufferOverflow` | Request too large |
| `RejectReasonInconsistentParameters` | Parameters don't make sense together |
| `RejectReasonInvalidParameterDataType` | Wrong parameter type |
| `RejectReasonInvalidTag` | Invalid ASN.1 tag |
| `RejectReasonMissingRequiredParameter` | Required parameter missing |
| `RejectReasonParameterOutOfRange` | Parameter value out of range |
| `RejectReasonTooManyArguments` | Too many parameters |
| `RejectReasonUndefinedEnumeration` | Invalid enumeration value |
| `RejectReasonUnrecognizedService` | Service not supported |

## Abort Errors

Transactions may be aborted. These are wrapped in `AbortError`.

```go
var abortErr *bacnet.AbortError
if errors.As(err, &abortErr) {
    fmt.Printf("Invoke ID: %d\n", abortErr.InvokeID)
    fmt.Printf("Server abort: %v\n", abortErr.Server)
    fmt.Printf("Reason: %s\n", abortErr.Reason)
}
```

### Abort Reasons

| Reason | Description |
|--------|-------------|
| `AbortReasonBufferOverflow` | Buffer overflow |
| `AbortReasonInvalidApduInThisState` | Invalid APDU for current state |
| `AbortReasonPreemptedByHigherPriorityTask` | Preempted |
| `AbortReasonSegmentationNotSupported` | Segmentation not supported |
| `AbortReasonSecurityError` | Security error |
| `AbortReasonInsufficientSecurity` | Insufficient security level |
| `AbortReasonWindowSizeOutOfRange` | Window size invalid |
| `AbortReasonApplicationExceededReplyTime` | Reply took too long |
| `AbortReasonOutOfResources` | Out of resources |
| `AbortReasonTsmTimeout` | Transaction state machine timeout |
| `AbortReasonApduTooLong` | APDU too long |

## Helper Functions

### IsTimeout

```go
func IsTimeout(err error) bool
```

Returns true if the error is a timeout.

```go
if bacnet.IsTimeout(err) {
    log.Println("Device not responding - check network connectivity")
}
```

### IsDeviceNotFound

```go
func IsDeviceNotFound(err error) bool
```

Returns true if the device was not found. Checks both the sentinel error and BACnet error codes.

```go
if bacnet.IsDeviceNotFound(err) {
    log.Println("Device not on network - verify device ID and network")
}
```

### IsPropertyNotFound

```go
func IsPropertyNotFound(err error) bool
```

Returns true if the property was not found.

```go
if bacnet.IsPropertyNotFound(err) {
    log.Println("Property not supported by this object")
}
```

### IsAccessDenied

```go
func IsAccessDenied(err error) bool
```

Returns true if access was denied (read or write).

```go
if bacnet.IsAccessDenied(err) {
    log.Println("Permission denied - check device security settings")
}
```

## Error Handling Patterns

### Comprehensive Error Handling

```go
value, err := client.ReadProperty(ctx, deviceID, objectID, propertyID)
if err != nil {
    switch {
    case bacnet.IsTimeout(err):
        return fmt.Errorf("device %d not responding: %w", deviceID, err)

    case bacnet.IsDeviceNotFound(err):
        return fmt.Errorf("device %d not found on network: %w", deviceID, err)

    case bacnet.IsPropertyNotFound(err):
        return fmt.Errorf("property %s not supported: %w", propertyID, err)

    case bacnet.IsAccessDenied(err):
        return fmt.Errorf("access denied to %s: %w", objectID, err)

    default:
        // Check for specific BACnet errors
        var bacnetErr *bacnet.BACnetError
        if errors.As(err, &bacnetErr) {
            return fmt.Errorf("BACnet error [%s/%s]: %w",
                bacnetErr.Class, bacnetErr.Code, err)
        }
        return fmt.Errorf("unexpected error: %w", err)
    }
}
```

### Retry Logic

```go
func readWithRetry(ctx context.Context, client *bacnet.Client, deviceID uint32, objectID bacnet.ObjectIdentifier, propertyID bacnet.PropertyIdentifier) (interface{}, error) {
    var lastErr error

    for attempt := 0; attempt < 3; attempt++ {
        value, err := client.ReadProperty(ctx, deviceID, objectID, propertyID)
        if err == nil {
            return value, nil
        }

        lastErr = err

        // Don't retry on certain errors
        if bacnet.IsDeviceNotFound(err) ||
           bacnet.IsPropertyNotFound(err) ||
           bacnet.IsAccessDenied(err) {
            return nil, err
        }

        // Retry on timeout or transient errors
        if bacnet.IsTimeout(err) {
            time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
            continue
        }

        // Check if device is busy
        var bacnetErr *bacnet.BACnetError
        if errors.As(err, &bacnetErr) {
            if bacnetErr.Code == bacnet.ErrorCodeDeviceBusy {
                time.Sleep(time.Duration(attempt+1) * time.Second)
                continue
            }
        }

        // Unknown error - don't retry
        return nil, err
    }

    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

### Error Logging

```go
func logBACnetError(logger *slog.Logger, operation string, err error) {
    if err == nil {
        return
    }

    attrs := []any{
        slog.String("operation", operation),
        slog.String("error", err.Error()),
    }

    var bacnetErr *bacnet.BACnetError
    if errors.As(err, &bacnetErr) {
        attrs = append(attrs,
            slog.String("error_class", bacnetErr.Class.String()),
            slog.String("error_code", bacnetErr.Code.String()),
        )
    }

    var rejectErr *bacnet.RejectError
    if errors.As(err, &rejectErr) {
        attrs = append(attrs,
            slog.Uint64("invoke_id", uint64(rejectErr.InvokeID)),
            slog.String("reject_reason", rejectErr.Reason.String()),
        )
    }

    var abortErr *bacnet.AbortError
    if errors.As(err, &abortErr) {
        attrs = append(attrs,
            slog.Uint64("invoke_id", uint64(abortErr.InvokeID)),
            slog.Bool("server", abortErr.Server),
            slog.String("abort_reason", abortErr.Reason.String()),
        )
    }

    logger.Error("BACnet operation failed", attrs...)
}
```

## Creating Custom Errors

```go
// Create a BACnet error
err := bacnet.NewBACnetError(
    bacnet.ErrorClassProperty,
    bacnet.ErrorCodeUnknownProperty,
)

// Error message: "bacnet error: class=property, code=unknown-property"
fmt.Println(err.Error())
```
