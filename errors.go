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
	"errors"
	"fmt"
)

// Sentinel errors
var (
	ErrTimeout           = errors.New("bacnet: request timeout")
	ErrConnectionClosed  = errors.New("bacnet: connection closed")
	ErrInvalidResponse   = errors.New("bacnet: invalid response")
	ErrInvalidAPDU       = errors.New("bacnet: invalid APDU")
	ErrInvalidNPDU       = errors.New("bacnet: invalid NPDU")
	ErrInvalidBVLC       = errors.New("bacnet: invalid BVLC header")
	ErrSegmentationNotSupported = errors.New("bacnet: segmentation not supported")
	ErrDeviceNotFound    = errors.New("bacnet: device not found")
	ErrPropertyNotFound  = errors.New("bacnet: property not found")
	ErrWriteFailed       = errors.New("bacnet: write failed")
	ErrNotConnected      = errors.New("bacnet: not connected")
	ErrAlreadyConnected  = errors.New("bacnet: already connected")
)

// ErrorClass represents BACnet error classes
type ErrorClass uint8

const (
	ErrorClassDevice          ErrorClass = 0
	ErrorClassObject          ErrorClass = 1
	ErrorClassProperty        ErrorClass = 2
	ErrorClassResources       ErrorClass = 3
	ErrorClassSecurity        ErrorClass = 4
	ErrorClassServices        ErrorClass = 5
	ErrorClassVT              ErrorClass = 6
	ErrorClassCommunication   ErrorClass = 7
)

func (e ErrorClass) String() string {
	names := map[ErrorClass]string{
		ErrorClassDevice:        "device",
		ErrorClassObject:        "object",
		ErrorClassProperty:      "property",
		ErrorClassResources:     "resources",
		ErrorClassSecurity:      "security",
		ErrorClassServices:      "services",
		ErrorClassVT:            "vt",
		ErrorClassCommunication: "communication",
	}
	if name, ok := names[e]; ok {
		return name
	}
	return fmt.Sprintf("error-class(%d)", e)
}

// ErrorCode represents BACnet error codes
type ErrorCode uint8

const (
	// Device errors
	ErrorCodeOther                        ErrorCode = 0
	ErrorCodeConfigurationInProgress      ErrorCode = 2
	ErrorCodeDeviceBusy                   ErrorCode = 3

	// Object errors
	ErrorCodeDynamicCreationNotSupported  ErrorCode = 4
	ErrorCodeNoObjectsOfSpecifiedType     ErrorCode = 17
	ErrorCodeObjectDeletionNotPermitted   ErrorCode = 23
	ErrorCodeObjectIdentifierAlreadyExists ErrorCode = 24
	ErrorCodeUnknownObject                ErrorCode = 31

	// Property errors
	ErrorCodeCharacterSetNotSupported     ErrorCode = 41
	ErrorCodeDatatypeNotSupported         ErrorCode = 47
	ErrorCodeInconsistentParameters       ErrorCode = 7
	ErrorCodeInvalidArrayIndex            ErrorCode = 42
	ErrorCodeInvalidDataType              ErrorCode = 9
	ErrorCodeNotCovProperty               ErrorCode = 44
	ErrorCodeOptionalFunctionalityNotSupported ErrorCode = 45
	ErrorCodePropertyIsNotAList           ErrorCode = 22
	ErrorCodePropertyIsNotAnArray         ErrorCode = 50
	ErrorCodeReadAccessDenied             ErrorCode = 27
	ErrorCodeUnknownProperty              ErrorCode = 32
	ErrorCodeValueOutOfRange              ErrorCode = 37
	ErrorCodeWriteAccessDenied            ErrorCode = 40

	// Resources errors
	ErrorCodeNoSpaceForObject             ErrorCode = 18
	ErrorCodeNoSpaceToAddListElement      ErrorCode = 19
	ErrorCodeNoSpaceToWriteProperty       ErrorCode = 20

	// Security errors
	ErrorCodeAuthenticationFailed         ErrorCode = 1
	ErrorCodeIncompatibleSecurityLevels   ErrorCode = 6
	ErrorCodeInvalidOperatorName          ErrorCode = 12
	ErrorCodeKeyGenerationError           ErrorCode = 15
	ErrorCodePasswordFailure              ErrorCode = 26
	ErrorCodeSecurityNotSupported         ErrorCode = 28

	// Services errors
	ErrorCodeCovSubscriptionFailed        ErrorCode = 43
	ErrorCodeDuplicateName                ErrorCode = 48
	ErrorCodeDuplicateObjectId            ErrorCode = 49
	ErrorCodeFileAccessDenied             ErrorCode = 5
	ErrorCodeInconsistentSelectionCriterion ErrorCode = 8
	ErrorCodeInvalidConfigurationData     ErrorCode = 46
	ErrorCodeInvalidFileAccessMethod      ErrorCode = 10
	ErrorCodeInvalidFileStartPosition     ErrorCode = 11
	ErrorCodeInvalidParameterDataType     ErrorCode = 13
	ErrorCodeInvalidTimeStamp             ErrorCode = 14
	ErrorCodeMissingRequiredParameter     ErrorCode = 16
	ErrorCodeNoAlarmsOfSpecifiedType      ErrorCode = 51
	ErrorCodeNotConfiguredForTriggeredLogging ErrorCode = 21
	ErrorCodeServiceRequestDenied         ErrorCode = 29
	ErrorCodeUnknownSubscription          ErrorCode = 33
	ErrorCodeUnknownVtClass               ErrorCode = 34
	ErrorCodeUnknownVtSession             ErrorCode = 35

	// Communication errors
	ErrorCodeAbortBufferOverflow          ErrorCode = 51
	ErrorCodeAbortInvalidApduInThisState  ErrorCode = 52
	ErrorCodeAbortPreemptedByHigherPriorityTask ErrorCode = 53
	ErrorCodeAbortSegmentationNotSupported ErrorCode = 54
	ErrorCodeAbortProprietary             ErrorCode = 55
	ErrorCodeAbortOther                   ErrorCode = 56
	ErrorCodeInvalidTag                   ErrorCode = 57
	ErrorCodeNetworkDown                  ErrorCode = 58
	ErrorCodeRejectBufferOverflow         ErrorCode = 59
	ErrorCodeRejectInconsistentParameters ErrorCode = 60
	ErrorCodeRejectInvalidParameterDataType ErrorCode = 61
	ErrorCodeRejectInvalidTag             ErrorCode = 62
	ErrorCodeRejectMissingRequiredParameter ErrorCode = 63
	ErrorCodeRejectParameterOutOfRange    ErrorCode = 64
	ErrorCodeRejectTooManyArguments       ErrorCode = 65
	ErrorCodeRejectUndefinedEnumeration   ErrorCode = 66
	ErrorCodeRejectUnrecognizedService    ErrorCode = 67
	ErrorCodeRejectProprietary            ErrorCode = 68
	ErrorCodeRejectOther                  ErrorCode = 69
	ErrorCodeUnknownDevice                ErrorCode = 70
	ErrorCodeUnknownRoute                 ErrorCode = 71
	ErrorCodeValueTooLong                 ErrorCode = 72
	ErrorCodeAbortApduTooLong             ErrorCode = 73
	ErrorCodeAbortApplicationExceededReplyTime ErrorCode = 74
	ErrorCodeAbortOutOfResources          ErrorCode = 75
	ErrorCodeAbortTsmTimeout              ErrorCode = 76
	ErrorCodeAbortWindowSizeOutOfRange    ErrorCode = 77
	ErrorCodeListItemNotNumbered          ErrorCode = 123
)

func (e ErrorCode) String() string {
	names := map[ErrorCode]string{
		ErrorCodeOther:                        "other",
		ErrorCodeConfigurationInProgress:      "configuration-in-progress",
		ErrorCodeDeviceBusy:                   "device-busy",
		ErrorCodeDynamicCreationNotSupported:  "dynamic-creation-not-supported",
		ErrorCodeNoObjectsOfSpecifiedType:     "no-objects-of-specified-type",
		ErrorCodeObjectDeletionNotPermitted:   "object-deletion-not-permitted",
		ErrorCodeObjectIdentifierAlreadyExists: "object-identifier-already-exists",
		ErrorCodeUnknownObject:                "unknown-object",
		ErrorCodeCharacterSetNotSupported:     "character-set-not-supported",
		ErrorCodeDatatypeNotSupported:         "datatype-not-supported",
		ErrorCodeInconsistentParameters:       "inconsistent-parameters",
		ErrorCodeInvalidArrayIndex:            "invalid-array-index",
		ErrorCodeInvalidDataType:              "invalid-data-type",
		ErrorCodeNotCovProperty:               "not-cov-property",
		ErrorCodeOptionalFunctionalityNotSupported: "optional-functionality-not-supported",
		ErrorCodePropertyIsNotAList:           "property-is-not-a-list",
		ErrorCodePropertyIsNotAnArray:         "property-is-not-an-array",
		ErrorCodeReadAccessDenied:             "read-access-denied",
		ErrorCodeUnknownProperty:              "unknown-property",
		ErrorCodeValueOutOfRange:              "value-out-of-range",
		ErrorCodeWriteAccessDenied:            "write-access-denied",
		ErrorCodeNoSpaceForObject:             "no-space-for-object",
		ErrorCodeNoSpaceToAddListElement:      "no-space-to-add-list-element",
		ErrorCodeNoSpaceToWriteProperty:       "no-space-to-write-property",
		ErrorCodeAuthenticationFailed:         "authentication-failed",
		ErrorCodePasswordFailure:              "password-failure",
		ErrorCodeSecurityNotSupported:         "security-not-supported",
		ErrorCodeServiceRequestDenied:         "service-request-denied",
		ErrorCodeUnknownDevice:                "unknown-device",
		ErrorCodeUnknownRoute:                 "unknown-route",
	}
	if name, ok := names[e]; ok {
		return name
	}
	return fmt.Sprintf("error-code(%d)", e)
}

// BACnetError represents a BACnet protocol error
type BACnetError struct {
	Class ErrorClass
	Code  ErrorCode
}

func (e *BACnetError) Error() string {
	return fmt.Sprintf("bacnet error: class=%s, code=%s", e.Class, e.Code)
}

func (e *BACnetError) Is(target error) bool {
	t, ok := target.(*BACnetError)
	if !ok {
		return false
	}
	return e.Class == t.Class && e.Code == t.Code
}

// NewBACnetError creates a new BACnet error
func NewBACnetError(class ErrorClass, code ErrorCode) *BACnetError {
	return &BACnetError{
		Class: class,
		Code:  code,
	}
}

// RejectReason represents BACnet reject reasons
type RejectReason uint8

const (
	RejectReasonOther                    RejectReason = 0
	RejectReasonBufferOverflow           RejectReason = 1
	RejectReasonInconsistentParameters   RejectReason = 2
	RejectReasonInvalidParameterDataType RejectReason = 3
	RejectReasonInvalidTag               RejectReason = 4
	RejectReasonMissingRequiredParameter RejectReason = 5
	RejectReasonParameterOutOfRange      RejectReason = 6
	RejectReasonTooManyArguments         RejectReason = 7
	RejectReasonUndefinedEnumeration     RejectReason = 8
	RejectReasonUnrecognizedService      RejectReason = 9
)

func (r RejectReason) String() string {
	names := map[RejectReason]string{
		RejectReasonOther:                    "other",
		RejectReasonBufferOverflow:           "buffer-overflow",
		RejectReasonInconsistentParameters:   "inconsistent-parameters",
		RejectReasonInvalidParameterDataType: "invalid-parameter-data-type",
		RejectReasonInvalidTag:               "invalid-tag",
		RejectReasonMissingRequiredParameter: "missing-required-parameter",
		RejectReasonParameterOutOfRange:      "parameter-out-of-range",
		RejectReasonTooManyArguments:         "too-many-arguments",
		RejectReasonUndefinedEnumeration:     "undefined-enumeration",
		RejectReasonUnrecognizedService:      "unrecognized-service",
	}
	if name, ok := names[r]; ok {
		return name
	}
	return fmt.Sprintf("reject-reason(%d)", r)
}

// RejectError represents a BACnet reject response
type RejectError struct {
	InvokeID uint8
	Reason   RejectReason
}

func (e *RejectError) Error() string {
	return fmt.Sprintf("bacnet reject: invoke-id=%d, reason=%s", e.InvokeID, e.Reason)
}

// AbortReason represents BACnet abort reasons
type AbortReason uint8

const (
	AbortReasonOther                        AbortReason = 0
	AbortReasonBufferOverflow               AbortReason = 1
	AbortReasonInvalidApduInThisState       AbortReason = 2
	AbortReasonPreemptedByHigherPriorityTask AbortReason = 3
	AbortReasonSegmentationNotSupported     AbortReason = 4
	AbortReasonSecurityError                AbortReason = 5
	AbortReasonInsufficientSecurity         AbortReason = 6
	AbortReasonWindowSizeOutOfRange         AbortReason = 7
	AbortReasonApplicationExceededReplyTime AbortReason = 8
	AbortReasonOutOfResources               AbortReason = 9
	AbortReasonTsmTimeout                   AbortReason = 10
	AbortReasonApduTooLong                  AbortReason = 11
)

func (a AbortReason) String() string {
	names := map[AbortReason]string{
		AbortReasonOther:                        "other",
		AbortReasonBufferOverflow:               "buffer-overflow",
		AbortReasonInvalidApduInThisState:       "invalid-apdu-in-this-state",
		AbortReasonPreemptedByHigherPriorityTask: "preempted-by-higher-priority-task",
		AbortReasonSegmentationNotSupported:     "segmentation-not-supported",
		AbortReasonSecurityError:                "security-error",
		AbortReasonInsufficientSecurity:         "insufficient-security",
		AbortReasonWindowSizeOutOfRange:         "window-size-out-of-range",
		AbortReasonApplicationExceededReplyTime: "application-exceeded-reply-time",
		AbortReasonOutOfResources:               "out-of-resources",
		AbortReasonTsmTimeout:                   "tsm-timeout",
		AbortReasonApduTooLong:                  "apdu-too-long",
	}
	if name, ok := names[a]; ok {
		return name
	}
	return fmt.Sprintf("abort-reason(%d)", a)
}

// AbortError represents a BACnet abort response
type AbortError struct {
	InvokeID uint8
	Server   bool
	Reason   AbortReason
}

func (e *AbortError) Error() string {
	origin := "client"
	if e.Server {
		origin = "server"
	}
	return fmt.Sprintf("bacnet abort: invoke-id=%d, origin=%s, reason=%s", e.InvokeID, origin, e.Reason)
}

// IsTimeout returns true if the error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsDeviceNotFound returns true if the error indicates device not found
func IsDeviceNotFound(err error) bool {
	if errors.Is(err, ErrDeviceNotFound) {
		return true
	}
	var bacnetErr *BACnetError
	if errors.As(err, &bacnetErr) {
		return bacnetErr.Code == ErrorCodeUnknownDevice || bacnetErr.Code == ErrorCodeUnknownObject
	}
	return false
}

// IsPropertyNotFound returns true if the error indicates property not found
func IsPropertyNotFound(err error) bool {
	if errors.Is(err, ErrPropertyNotFound) {
		return true
	}
	var bacnetErr *BACnetError
	if errors.As(err, &bacnetErr) {
		return bacnetErr.Code == ErrorCodeUnknownProperty
	}
	return false
}

// IsAccessDenied returns true if the error indicates access denied
func IsAccessDenied(err error) bool {
	var bacnetErr *BACnetError
	if errors.As(err, &bacnetErr) {
		return bacnetErr.Code == ErrorCodeReadAccessDenied || bacnetErr.Code == ErrorCodeWriteAccessDenied
	}
	return false
}
