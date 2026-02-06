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

// Package bacnet provides a BACnet/IP client implementation for building automation systems.
package bacnet

import (
	"encoding/binary"
	"fmt"
)

// DefaultPort is the standard BACnet/IP UDP port
const DefaultPort = 47808

// MaxAPDULength is the maximum APDU length for BACnet/IP
const MaxAPDULength = 1476

// BVLC Types (BACnet Virtual Link Control)
type BVLCType uint8

const (
	BVLCTypeBACnetIP BVLCType = 0x81
)

// BVLC Functions
type BVLCFunction uint8

const (
	BVLCResult                          BVLCFunction = 0x00
	BVLCWriteBroadcastDistributionTable BVLCFunction = 0x01
	BVLCReadBroadcastDistributionTable  BVLCFunction = 0x02
	BVLCReadBroadcastDistributionTableAck BVLCFunction = 0x03
	BVLCForwardedNPDU                   BVLCFunction = 0x04
	BVLCRegisterForeignDevice           BVLCFunction = 0x05
	BVLCReadForeignDeviceTable          BVLCFunction = 0x06
	BVLCReadForeignDeviceTableAck       BVLCFunction = 0x07
	BVLCDeleteForeignDeviceTableEntry   BVLCFunction = 0x08
	BVLCDistributeBroadcastToNetwork    BVLCFunction = 0x09
	BVLCOriginalUnicastNPDU             BVLCFunction = 0x0A
	BVLCOriginalBroadcastNPDU           BVLCFunction = 0x0B
	BVLCSecureBVLL                       BVLCFunction = 0x0C
)

// NPDU Network Layer Protocol Control Information
type NPDUControl uint8

const (
	NPDUControlNetworkLayerMessage NPDUControl = 0x80
	NPDUControlDestSpecifier       NPDUControl = 0x20
	NPDUControlSourceSpecifier     NPDUControl = 0x08
	NPDUControlExpectingReply      NPDUControl = 0x04
	NPDUControlPriorityNormal      NPDUControl = 0x00
	NPDUControlPriorityUrgent      NPDUControl = 0x01
	NPDUControlPriorityCritical    NPDUControl = 0x02
	NPDUControlPriorityLifeSafety  NPDUControl = 0x03
)

// Network Layer Message Types
type NetworkMessageType uint8

const (
	NetworkMessageWhoIsRouterToNetwork   NetworkMessageType = 0x00
	NetworkMessageIAmRouterToNetwork     NetworkMessageType = 0x01
	NetworkMessageICouldBeRouterToNetwork NetworkMessageType = 0x02
	NetworkMessageRejectMessageToNetwork NetworkMessageType = 0x03
	NetworkMessageRouterBusyToNetwork    NetworkMessageType = 0x04
	NetworkMessageRouterAvailableToNetwork NetworkMessageType = 0x05
	NetworkMessageInitializeRoutingTable NetworkMessageType = 0x06
	NetworkMessageInitializeRoutingTableAck NetworkMessageType = 0x07
	NetworkMessageEstablishConnectionToNetwork NetworkMessageType = 0x08
	NetworkMessageDisconnectConnectionToNetwork NetworkMessageType = 0x09
	NetworkMessageWhatIsNetworkNumber    NetworkMessageType = 0x12
	NetworkMessageNetworkNumberIs        NetworkMessageType = 0x13
)

// PDU Types (Application Layer)
type PDUType uint8

const (
	PDUTypeConfirmedRequest   PDUType = 0x00
	PDUTypeUnconfirmedRequest PDUType = 0x10
	PDUTypeSimpleAck          PDUType = 0x20
	PDUTypeComplexAck         PDUType = 0x30
	PDUTypeSegmentAck         PDUType = 0x40
	PDUTypeError              PDUType = 0x50
	PDUTypeReject             PDUType = 0x60
	PDUTypeAbort              PDUType = 0x70
)

// Confirmed Service Choices
type ConfirmedServiceChoice uint8

const (
	ServiceAcknowledgeAlarm          ConfirmedServiceChoice = 0
	ServiceConfirmedCOVNotification  ConfirmedServiceChoice = 1
	ServiceConfirmedEventNotification ConfirmedServiceChoice = 2
	ServiceGetAlarmSummary           ConfirmedServiceChoice = 3
	ServiceGetEnrollmentSummary      ConfirmedServiceChoice = 4
	ServiceSubscribeCOV              ConfirmedServiceChoice = 5
	ServiceAtomicReadFile            ConfirmedServiceChoice = 6
	ServiceAtomicWriteFile           ConfirmedServiceChoice = 7
	ServiceAddListElement            ConfirmedServiceChoice = 8
	ServiceRemoveListElement         ConfirmedServiceChoice = 9
	ServiceCreateObject              ConfirmedServiceChoice = 10
	ServiceDeleteObject              ConfirmedServiceChoice = 11
	ServiceReadProperty              ConfirmedServiceChoice = 12
	ServiceReadPropertyConditional   ConfirmedServiceChoice = 13
	ServiceReadPropertyMultiple      ConfirmedServiceChoice = 14
	ServiceWriteProperty             ConfirmedServiceChoice = 15
	ServiceWritePropertyMultiple     ConfirmedServiceChoice = 16
	ServiceDeviceCommunicationControl ConfirmedServiceChoice = 17
	ServiceConfirmedPrivateTransfer  ConfirmedServiceChoice = 18
	ServiceConfirmedTextMessage      ConfirmedServiceChoice = 19
	ServiceReinitializeDevice        ConfirmedServiceChoice = 20
	ServiceVTOpen                    ConfirmedServiceChoice = 21
	ServiceVTClose                   ConfirmedServiceChoice = 22
	ServiceVTData                    ConfirmedServiceChoice = 23
	ServiceAuthenticate              ConfirmedServiceChoice = 24
	ServiceRequestKey                ConfirmedServiceChoice = 25
	ServiceReadRange                 ConfirmedServiceChoice = 26
	ServiceLifeSafetyOperation       ConfirmedServiceChoice = 27
	ServiceSubscribeCOVProperty      ConfirmedServiceChoice = 28
	ServiceGetEventInformation       ConfirmedServiceChoice = 29
)

func (s ConfirmedServiceChoice) String() string {
	names := map[ConfirmedServiceChoice]string{
		ServiceAcknowledgeAlarm:          "AcknowledgeAlarm",
		ServiceConfirmedCOVNotification:  "ConfirmedCOVNotification",
		ServiceConfirmedEventNotification: "ConfirmedEventNotification",
		ServiceGetAlarmSummary:           "GetAlarmSummary",
		ServiceGetEnrollmentSummary:      "GetEnrollmentSummary",
		ServiceSubscribeCOV:              "SubscribeCOV",
		ServiceAtomicReadFile:            "AtomicReadFile",
		ServiceAtomicWriteFile:           "AtomicWriteFile",
		ServiceAddListElement:            "AddListElement",
		ServiceRemoveListElement:         "RemoveListElement",
		ServiceCreateObject:              "CreateObject",
		ServiceDeleteObject:              "DeleteObject",
		ServiceReadProperty:              "ReadProperty",
		ServiceReadPropertyConditional:   "ReadPropertyConditional",
		ServiceReadPropertyMultiple:      "ReadPropertyMultiple",
		ServiceWriteProperty:             "WriteProperty",
		ServiceWritePropertyMultiple:     "WritePropertyMultiple",
		ServiceDeviceCommunicationControl: "DeviceCommunicationControl",
		ServiceConfirmedPrivateTransfer:  "ConfirmedPrivateTransfer",
		ServiceConfirmedTextMessage:      "ConfirmedTextMessage",
		ServiceReinitializeDevice:        "ReinitializeDevice",
		ServiceVTOpen:                    "VTOpen",
		ServiceVTClose:                   "VTClose",
		ServiceVTData:                    "VTData",
		ServiceAuthenticate:              "Authenticate",
		ServiceRequestKey:                "RequestKey",
		ServiceReadRange:                 "ReadRange",
		ServiceLifeSafetyOperation:       "LifeSafetyOperation",
		ServiceSubscribeCOVProperty:      "SubscribeCOVProperty",
		ServiceGetEventInformation:       "GetEventInformation",
	}
	if name, ok := names[s]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", s)
}

// Unconfirmed Service Choices
type UnconfirmedServiceChoice uint8

const (
	ServiceIAm                        UnconfirmedServiceChoice = 0
	ServiceIHave                      UnconfirmedServiceChoice = 1
	ServiceUnconfirmedCOVNotification UnconfirmedServiceChoice = 2
	ServiceUnconfirmedEventNotification UnconfirmedServiceChoice = 3
	ServiceUnconfirmedPrivateTransfer UnconfirmedServiceChoice = 4
	ServiceUnconfirmedTextMessage     UnconfirmedServiceChoice = 5
	ServiceTimeSynchronization        UnconfirmedServiceChoice = 6
	ServiceWhoHas                     UnconfirmedServiceChoice = 7
	ServiceWhoIs                      UnconfirmedServiceChoice = 8
	ServiceUTCTimeSynchronization     UnconfirmedServiceChoice = 9
	ServiceWriteGroup                 UnconfirmedServiceChoice = 10
)

func (s UnconfirmedServiceChoice) String() string {
	names := map[UnconfirmedServiceChoice]string{
		ServiceIAm:                        "I-Am",
		ServiceIHave:                      "I-Have",
		ServiceUnconfirmedCOVNotification: "UnconfirmedCOVNotification",
		ServiceUnconfirmedEventNotification: "UnconfirmedEventNotification",
		ServiceUnconfirmedPrivateTransfer: "UnconfirmedPrivateTransfer",
		ServiceUnconfirmedTextMessage:     "UnconfirmedTextMessage",
		ServiceTimeSynchronization:        "TimeSynchronization",
		ServiceWhoHas:                     "Who-Has",
		ServiceWhoIs:                      "Who-Is",
		ServiceUTCTimeSynchronization:     "UTCTimeSynchronization",
		ServiceWriteGroup:                 "WriteGroup",
	}
	if name, ok := names[s]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", s)
}

// ObjectType represents BACnet object types
type ObjectType uint16

const (
	ObjectTypeAnalogInput        ObjectType = 0
	ObjectTypeAnalogOutput       ObjectType = 1
	ObjectTypeAnalogValue        ObjectType = 2
	ObjectTypeBinaryInput        ObjectType = 3
	ObjectTypeBinaryOutput       ObjectType = 4
	ObjectTypeBinaryValue        ObjectType = 5
	ObjectTypeCalendar           ObjectType = 6
	ObjectTypeCommand            ObjectType = 7
	ObjectTypeDevice             ObjectType = 8
	ObjectTypeEventEnrollment    ObjectType = 9
	ObjectTypeFile               ObjectType = 10
	ObjectTypeGroup              ObjectType = 11
	ObjectTypeLoop               ObjectType = 12
	ObjectTypeMultiStateInput    ObjectType = 13
	ObjectTypeMultiStateOutput   ObjectType = 14
	ObjectTypeNotificationClass  ObjectType = 15
	ObjectTypeProgram            ObjectType = 16
	ObjectTypeSchedule           ObjectType = 17
	ObjectTypeAveraging          ObjectType = 18
	ObjectTypeMultiStateValue    ObjectType = 19
	ObjectTypeTrendLog           ObjectType = 20
	ObjectTypeLifeSafetyPoint    ObjectType = 21
	ObjectTypeLifeSafetyZone     ObjectType = 22
	ObjectTypeAccumulator        ObjectType = 23
	ObjectTypePulseConverter     ObjectType = 24
	ObjectTypeEventLog           ObjectType = 25
	ObjectTypeGlobalGroup        ObjectType = 26
	ObjectTypeTrendLogMultiple   ObjectType = 27
	ObjectTypeLoadControl        ObjectType = 28
	ObjectTypeStructuredView     ObjectType = 29
	ObjectTypeAccessDoor         ObjectType = 30
	ObjectTypeTimer              ObjectType = 31
	ObjectTypeAccessCredential   ObjectType = 32
	ObjectTypeAccessPoint        ObjectType = 33
	ObjectTypeAccessRights       ObjectType = 34
	ObjectTypeAccessUser         ObjectType = 35
	ObjectTypeAccessZone         ObjectType = 36
	ObjectTypeCredentialDataInput ObjectType = 37
	ObjectTypeNetworkSecurity    ObjectType = 38
	ObjectTypeBitStringValue     ObjectType = 39
	ObjectTypeCharacterStringValue ObjectType = 40
	ObjectTypeDatePatternValue   ObjectType = 41
	ObjectTypeDateValue          ObjectType = 42
	ObjectTypeDateTimePatternValue ObjectType = 43
	ObjectTypeDateTimeValue      ObjectType = 44
	ObjectTypeIntegerValue       ObjectType = 45
	ObjectTypeLargeAnalogValue   ObjectType = 46
	ObjectTypeOctetStringValue   ObjectType = 47
	ObjectTypePositiveIntegerValue ObjectType = 48
	ObjectTypeTimePatternValue   ObjectType = 49
	ObjectTypeTimeValue          ObjectType = 50
	ObjectTypeNotificationForwarder ObjectType = 51
	ObjectTypeAlertEnrollment    ObjectType = 52
	ObjectTypeChannel            ObjectType = 53
	ObjectTypeLightingOutput     ObjectType = 54
	ObjectTypeBinaryLightingOutput ObjectType = 55
	ObjectTypeNetworkPort        ObjectType = 56
	ObjectTypeElevatorGroup      ObjectType = 57
	ObjectTypeEscalator          ObjectType = 58
	ObjectTypeLift               ObjectType = 59
)

func (o ObjectType) String() string {
	names := map[ObjectType]string{
		ObjectTypeAnalogInput:        "analog-input",
		ObjectTypeAnalogOutput:       "analog-output",
		ObjectTypeAnalogValue:        "analog-value",
		ObjectTypeBinaryInput:        "binary-input",
		ObjectTypeBinaryOutput:       "binary-output",
		ObjectTypeBinaryValue:        "binary-value",
		ObjectTypeCalendar:           "calendar",
		ObjectTypeCommand:            "command",
		ObjectTypeDevice:             "device",
		ObjectTypeEventEnrollment:    "event-enrollment",
		ObjectTypeFile:               "file",
		ObjectTypeGroup:              "group",
		ObjectTypeLoop:               "loop",
		ObjectTypeMultiStateInput:    "multi-state-input",
		ObjectTypeMultiStateOutput:   "multi-state-output",
		ObjectTypeNotificationClass:  "notification-class",
		ObjectTypeProgram:            "program",
		ObjectTypeSchedule:           "schedule",
		ObjectTypeAveraging:          "averaging",
		ObjectTypeMultiStateValue:    "multi-state-value",
		ObjectTypeTrendLog:           "trend-log",
		ObjectTypeLifeSafetyPoint:    "life-safety-point",
		ObjectTypeLifeSafetyZone:     "life-safety-zone",
		ObjectTypeAccumulator:        "accumulator",
		ObjectTypePulseConverter:     "pulse-converter",
		ObjectTypeEventLog:           "event-log",
		ObjectTypeGlobalGroup:        "global-group",
		ObjectTypeTrendLogMultiple:   "trend-log-multiple",
		ObjectTypeLoadControl:        "load-control",
		ObjectTypeStructuredView:     "structured-view",
		ObjectTypeAccessDoor:         "access-door",
		ObjectTypeTimer:              "timer",
		ObjectTypeAccessCredential:   "access-credential",
		ObjectTypeAccessPoint:        "access-point",
		ObjectTypeAccessRights:       "access-rights",
		ObjectTypeAccessUser:         "access-user",
		ObjectTypeAccessZone:         "access-zone",
		ObjectTypeCredentialDataInput: "credential-data-input",
		ObjectTypeNetworkSecurity:    "network-security",
		ObjectTypeBitStringValue:     "bitstring-value",
		ObjectTypeCharacterStringValue: "characterstring-value",
		ObjectTypeDatePatternValue:   "date-pattern-value",
		ObjectTypeDateValue:          "date-value",
		ObjectTypeDateTimePatternValue: "datetime-pattern-value",
		ObjectTypeDateTimeValue:      "datetime-value",
		ObjectTypeIntegerValue:       "integer-value",
		ObjectTypeLargeAnalogValue:   "large-analog-value",
		ObjectTypeOctetStringValue:   "octetstring-value",
		ObjectTypePositiveIntegerValue: "positive-integer-value",
		ObjectTypeTimePatternValue:   "time-pattern-value",
		ObjectTypeTimeValue:          "time-value",
		ObjectTypeNotificationForwarder: "notification-forwarder",
		ObjectTypeAlertEnrollment:    "alert-enrollment",
		ObjectTypeChannel:            "channel",
		ObjectTypeLightingOutput:     "lighting-output",
		ObjectTypeBinaryLightingOutput: "binary-lighting-output",
		ObjectTypeNetworkPort:        "network-port",
		ObjectTypeElevatorGroup:      "elevator-group",
		ObjectTypeEscalator:          "escalator",
		ObjectTypeLift:               "lift",
	}
	if name, ok := names[o]; ok {
		return name
	}
	return fmt.Sprintf("vendor-specific(%d)", o)
}

// ParseObjectType parses a string to ObjectType
func ParseObjectType(s string) (ObjectType, bool) {
	types := map[string]ObjectType{
		"analog-input":        ObjectTypeAnalogInput,
		"ai":                  ObjectTypeAnalogInput,
		"analog-output":       ObjectTypeAnalogOutput,
		"ao":                  ObjectTypeAnalogOutput,
		"analog-value":        ObjectTypeAnalogValue,
		"av":                  ObjectTypeAnalogValue,
		"binary-input":        ObjectTypeBinaryInput,
		"bi":                  ObjectTypeBinaryInput,
		"binary-output":       ObjectTypeBinaryOutput,
		"bo":                  ObjectTypeBinaryOutput,
		"binary-value":        ObjectTypeBinaryValue,
		"bv":                  ObjectTypeBinaryValue,
		"device":              ObjectTypeDevice,
		"dev":                 ObjectTypeDevice,
		"multi-state-input":   ObjectTypeMultiStateInput,
		"msi":                 ObjectTypeMultiStateInput,
		"multi-state-output":  ObjectTypeMultiStateOutput,
		"mso":                 ObjectTypeMultiStateOutput,
		"multi-state-value":   ObjectTypeMultiStateValue,
		"msv":                 ObjectTypeMultiStateValue,
		"schedule":            ObjectTypeSchedule,
		"sch":                 ObjectTypeSchedule,
		"trend-log":           ObjectTypeTrendLog,
		"tl":                  ObjectTypeTrendLog,
		"calendar":            ObjectTypeCalendar,
		"cal":                 ObjectTypeCalendar,
		"notification-class":  ObjectTypeNotificationClass,
		"nc":                  ObjectTypeNotificationClass,
		"file":                ObjectTypeFile,
		"loop":                ObjectTypeLoop,
		"program":             ObjectTypeProgram,
		"prg":                 ObjectTypeProgram,
	}
	if t, ok := types[s]; ok {
		return t, true
	}
	return 0, false
}

// PropertyIdentifier represents BACnet property identifiers
type PropertyIdentifier uint32

const (
	PropertyAckedTransitions          PropertyIdentifier = 0
	PropertyAckRequired               PropertyIdentifier = 1
	PropertyAction                    PropertyIdentifier = 2
	PropertyActionText                PropertyIdentifier = 3
	PropertyActiveText                PropertyIdentifier = 4
	PropertyActiveVtSessions          PropertyIdentifier = 5
	PropertyAlarmValue                PropertyIdentifier = 6
	PropertyAlarmValues               PropertyIdentifier = 7
	PropertyAll                       PropertyIdentifier = 8
	PropertyAllWritesSuccessful       PropertyIdentifier = 9
	PropertyApduSegmentTimeout        PropertyIdentifier = 10
	PropertyApduTimeout               PropertyIdentifier = 11
	PropertyApplicationSoftwareVersion PropertyIdentifier = 12
	PropertyArchive                   PropertyIdentifier = 13
	PropertyBias                      PropertyIdentifier = 14
	PropertyChangeOfStateCount        PropertyIdentifier = 15
	PropertyChangeOfStateTime         PropertyIdentifier = 16
	PropertyNotificationClass         PropertyIdentifier = 17
	PropertyControlledVariableReference PropertyIdentifier = 19
	PropertyControlledVariableUnits   PropertyIdentifier = 20
	PropertyControlledVariableValue   PropertyIdentifier = 21
	PropertyCOVIncrement              PropertyIdentifier = 22
	PropertyDateList                  PropertyIdentifier = 23
	PropertyDaylightSavingsStatus     PropertyIdentifier = 24
	PropertyDeadband                  PropertyIdentifier = 25
	PropertyDerivativeConstant        PropertyIdentifier = 26
	PropertyDerivativeConstantUnits   PropertyIdentifier = 27
	PropertyDescription               PropertyIdentifier = 28
	PropertyDescriptionOfHalt         PropertyIdentifier = 29
	PropertyDeviceAddressBinding      PropertyIdentifier = 30
	PropertyDeviceType                PropertyIdentifier = 31
	PropertyEffectivePeriod           PropertyIdentifier = 32
	PropertyElapsedActiveTime         PropertyIdentifier = 33
	PropertyErrorLimit                PropertyIdentifier = 34
	PropertyEventEnable               PropertyIdentifier = 35
	PropertyEventState                PropertyIdentifier = 36
	PropertyEventType                 PropertyIdentifier = 37
	PropertyExceptionSchedule         PropertyIdentifier = 38
	PropertyFaultValues               PropertyIdentifier = 39
	PropertyFeedbackValue             PropertyIdentifier = 40
	PropertyFileAccessMethod          PropertyIdentifier = 41
	PropertyFileSize                  PropertyIdentifier = 42
	PropertyFileType                  PropertyIdentifier = 43
	PropertyFirmwareRevision          PropertyIdentifier = 44
	PropertyHighLimit                 PropertyIdentifier = 45
	PropertyInactiveText              PropertyIdentifier = 46
	PropertyInProcess                 PropertyIdentifier = 47
	PropertyInstanceOf                PropertyIdentifier = 48
	PropertyIntegralConstant          PropertyIdentifier = 49
	PropertyIntegralConstantUnits     PropertyIdentifier = 50
	PropertyLimitEnable               PropertyIdentifier = 52
	PropertyListOfGroupMembers        PropertyIdentifier = 53
	PropertyListOfObjectPropertyReferences PropertyIdentifier = 54
	PropertyLocalDate                 PropertyIdentifier = 56
	PropertyLocalTime                 PropertyIdentifier = 57
	PropertyLocation                  PropertyIdentifier = 58
	PropertyLowLimit                  PropertyIdentifier = 59
	PropertyManipulatedVariableReference PropertyIdentifier = 60
	PropertyMaximumOutput             PropertyIdentifier = 61
	PropertyMaxApduLengthAccepted     PropertyIdentifier = 62
	PropertyMaxInfoFrames             PropertyIdentifier = 63
	PropertyMaxMaster                 PropertyIdentifier = 64
	PropertyMaxPresValue              PropertyIdentifier = 65
	PropertyMinimumOffTime            PropertyIdentifier = 66
	PropertyMinimumOnTime             PropertyIdentifier = 67
	PropertyMinimumOutput             PropertyIdentifier = 68
	PropertyMinPresValue              PropertyIdentifier = 69
	PropertyModelName                 PropertyIdentifier = 70
	PropertyModificationDate          PropertyIdentifier = 71
	PropertyNotifyType                PropertyIdentifier = 72
	PropertyNumberOfApduRetries       PropertyIdentifier = 73
	PropertyNumberOfStates            PropertyIdentifier = 74
	PropertyObjectIdentifier          PropertyIdentifier = 75
	PropertyObjectList                PropertyIdentifier = 76
	PropertyObjectName                PropertyIdentifier = 77
	PropertyObjectPropertyReference   PropertyIdentifier = 78
	PropertyObjectType                PropertyIdentifier = 79
	PropertyOptional                  PropertyIdentifier = 80
	PropertyOutOfService              PropertyIdentifier = 81
	PropertyOutputUnits               PropertyIdentifier = 82
	PropertyEventParameters           PropertyIdentifier = 83
	PropertyPolarity                  PropertyIdentifier = 84
	PropertyPresentValue              PropertyIdentifier = 85
	PropertyPriority                  PropertyIdentifier = 86
	PropertyPriorityArray             PropertyIdentifier = 87
	PropertyPriorityForWriting        PropertyIdentifier = 88
	PropertyProcessIdentifier         PropertyIdentifier = 89
	PropertyProgramChange             PropertyIdentifier = 90
	PropertyProgramLocation           PropertyIdentifier = 91
	PropertyProgramState              PropertyIdentifier = 92
	PropertyProportionalConstant      PropertyIdentifier = 93
	PropertyProportionalConstantUnits PropertyIdentifier = 94
	PropertyProtocolObjectTypesSupported PropertyIdentifier = 96
	PropertyProtocolServicesSupported PropertyIdentifier = 97
	PropertyProtocolVersion           PropertyIdentifier = 98
	PropertyReadOnly                  PropertyIdentifier = 99
	PropertyReasonForHalt             PropertyIdentifier = 100
	PropertyRecipientList             PropertyIdentifier = 102
	PropertyReliability               PropertyIdentifier = 103
	PropertyRelinquishDefault         PropertyIdentifier = 104
	PropertyRequired                  PropertyIdentifier = 105
	PropertyResolution                PropertyIdentifier = 106
	PropertySegmentationSupported     PropertyIdentifier = 107
	PropertySetpoint                  PropertyIdentifier = 108
	PropertySetpointReference         PropertyIdentifier = 109
	PropertyStateText                 PropertyIdentifier = 110
	PropertyStatusFlags               PropertyIdentifier = 111
	PropertySystemStatus              PropertyIdentifier = 112
	PropertyTimeDelay                 PropertyIdentifier = 113
	PropertyTimeOfActiveTimeReset     PropertyIdentifier = 114
	PropertyTimeOfStateCountReset     PropertyIdentifier = 115
	PropertyTimeSynchronizationRecipients PropertyIdentifier = 116
	PropertyUnits                     PropertyIdentifier = 117
	PropertyUpdateInterval            PropertyIdentifier = 118
	PropertyUtcOffset                 PropertyIdentifier = 119
	PropertyVendorIdentifier          PropertyIdentifier = 120
	PropertyVendorName                PropertyIdentifier = 121
	PropertyVtClassesSupported        PropertyIdentifier = 122
	PropertyWeeklySchedule            PropertyIdentifier = 123
	PropertyAttemptedSamples          PropertyIdentifier = 124
	PropertyAverageValue              PropertyIdentifier = 125
	PropertyBufferSize                PropertyIdentifier = 126
	PropertyClientCovIncrement        PropertyIdentifier = 127
	PropertyCOVResubscriptionInterval PropertyIdentifier = 128
	PropertyEventTimeStamps           PropertyIdentifier = 130
	PropertyLogBuffer                 PropertyIdentifier = 131
	PropertyLogDeviceObjectProperty   PropertyIdentifier = 132
	PropertyLogEnable                 PropertyIdentifier = 133
	PropertyLogInterval               PropertyIdentifier = 134
	PropertyMaximumValue              PropertyIdentifier = 135
	PropertyMinimumValue              PropertyIdentifier = 136
	PropertyNotificationThreshold     PropertyIdentifier = 137
	PropertyPreviousNotifyRecord      PropertyIdentifier = 138
	PropertyProtocolRevision          PropertyIdentifier = 139
	PropertyRecordsSinceNotification  PropertyIdentifier = 140
	PropertyRecordCount               PropertyIdentifier = 141
	PropertyStartTime                 PropertyIdentifier = 142
	PropertyStopTime                  PropertyIdentifier = 143
	PropertyStopWhenFull              PropertyIdentifier = 144
	PropertyTotalRecordCount          PropertyIdentifier = 145
	PropertyValidSamples              PropertyIdentifier = 146
	PropertyWindowInterval            PropertyIdentifier = 147
	PropertyWindowSamples             PropertyIdentifier = 148
	PropertyMaximumValueTimestamp     PropertyIdentifier = 149
	PropertyMinimumValueTimestamp     PropertyIdentifier = 150
	PropertyVarianceValue             PropertyIdentifier = 151
	PropertyActiveCOVSubscriptions    PropertyIdentifier = 152
	PropertyBackupFailureTimeout      PropertyIdentifier = 153
	PropertyConfigurationFiles        PropertyIdentifier = 154
	PropertyDatabaseRevision          PropertyIdentifier = 155
	PropertyDirectReading             PropertyIdentifier = 156
	PropertyLastRestoreTime           PropertyIdentifier = 157
	PropertyMaintenanceRequired       PropertyIdentifier = 158
	PropertyMemberOf                  PropertyIdentifier = 159
	PropertyMode                      PropertyIdentifier = 160
	PropertyOperationExpected         PropertyIdentifier = 161
	PropertySetting                   PropertyIdentifier = 162
	PropertySilenced                  PropertyIdentifier = 163
	PropertyTrackingValue             PropertyIdentifier = 164
	PropertyZoneMembers               PropertyIdentifier = 165
	PropertyLifeSafetyAlarmValues     PropertyIdentifier = 166
	PropertyMaxSegmentsAccepted       PropertyIdentifier = 167
	PropertyProfileName               PropertyIdentifier = 168
)

func (p PropertyIdentifier) String() string {
	names := map[PropertyIdentifier]string{
		PropertyObjectIdentifier: "object-identifier",
		PropertyObjectName:       "object-name",
		PropertyObjectType:       "object-type",
		PropertyPresentValue:     "present-value",
		PropertyDescription:      "description",
		PropertyDeviceType:       "device-type",
		PropertyStatusFlags:      "status-flags",
		PropertyEventState:       "event-state",
		PropertyReliability:      "reliability",
		PropertyOutOfService:     "out-of-service",
		PropertyUnits:            "units",
		PropertyPriorityArray:    "priority-array",
		PropertyRelinquishDefault: "relinquish-default",
		PropertyCOVIncrement:     "cov-increment",
		PropertyHighLimit:        "high-limit",
		PropertyLowLimit:         "low-limit",
		PropertyDeadband:         "deadband",
		PropertyVendorName:       "vendor-name",
		PropertyVendorIdentifier: "vendor-identifier",
		PropertyModelName:        "model-name",
		PropertyFirmwareRevision: "firmware-revision",
		PropertyApplicationSoftwareVersion: "application-software-version",
		PropertyProtocolVersion:  "protocol-version",
		PropertyProtocolRevision: "protocol-revision",
		PropertySystemStatus:     "system-status",
		PropertyMaxApduLengthAccepted: "max-apdu-length-accepted",
		PropertySegmentationSupported: "segmentation-supported",
		PropertyObjectList:       "object-list",
		PropertyDatabaseRevision: "database-revision",
		PropertyAll:              "all",
		PropertyRequired:         "required",
		PropertyOptional:         "optional",
	}
	if name, ok := names[p]; ok {
		return name
	}
	return fmt.Sprintf("property(%d)", p)
}

// ParsePropertyIdentifier parses a string to PropertyIdentifier
func ParsePropertyIdentifier(s string) (PropertyIdentifier, bool) {
	props := map[string]PropertyIdentifier{
		"object-identifier":       PropertyObjectIdentifier,
		"oid":                     PropertyObjectIdentifier,
		"object-name":             PropertyObjectName,
		"name":                    PropertyObjectName,
		"object-type":             PropertyObjectType,
		"type":                    PropertyObjectType,
		"present-value":           PropertyPresentValue,
		"pv":                      PropertyPresentValue,
		"description":             PropertyDescription,
		"desc":                    PropertyDescription,
		"status-flags":            PropertyStatusFlags,
		"sf":                      PropertyStatusFlags,
		"event-state":             PropertyEventState,
		"reliability":             PropertyReliability,
		"out-of-service":          PropertyOutOfService,
		"oos":                     PropertyOutOfService,
		"units":                   PropertyUnits,
		"priority-array":          PropertyPriorityArray,
		"pa":                      PropertyPriorityArray,
		"relinquish-default":      PropertyRelinquishDefault,
		"rd":                      PropertyRelinquishDefault,
		"cov-increment":           PropertyCOVIncrement,
		"vendor-name":             PropertyVendorName,
		"vendor-identifier":       PropertyVendorIdentifier,
		"model-name":              PropertyModelName,
		"firmware-revision":       PropertyFirmwareRevision,
		"application-software-version": PropertyApplicationSoftwareVersion,
		"protocol-version":        PropertyProtocolVersion,
		"protocol-revision":       PropertyProtocolRevision,
		"system-status":           PropertySystemStatus,
		"object-list":             PropertyObjectList,
		"database-revision":       PropertyDatabaseRevision,
		"all":                     PropertyAll,
	}
	if p, ok := props[s]; ok {
		return p, true
	}
	return 0, false
}

// ObjectIdentifier represents a BACnet object identifier (type + instance)
type ObjectIdentifier struct {
	Type     ObjectType
	Instance uint32
}

// NewObjectIdentifier creates a new ObjectIdentifier
func NewObjectIdentifier(objectType ObjectType, instance uint32) ObjectIdentifier {
	return ObjectIdentifier{
		Type:     objectType,
		Instance: instance,
	}
}

// Encode encodes the object identifier to a 4-byte value
func (o ObjectIdentifier) Encode() uint32 {
	return (uint32(o.Type) << 22) | (o.Instance & 0x3FFFFF)
}

// DecodeObjectIdentifier decodes a 4-byte value to an ObjectIdentifier
func DecodeObjectIdentifier(value uint32) ObjectIdentifier {
	return ObjectIdentifier{
		Type:     ObjectType((value >> 22) & 0x3FF),
		Instance: value & 0x3FFFFF,
	}
}

func (o ObjectIdentifier) String() string {
	return fmt.Sprintf("%s:%d", o.Type.String(), o.Instance)
}

// StatusFlags represents the BACnet status flags
type StatusFlags struct {
	InAlarm      bool
	Fault        bool
	Overridden   bool
	OutOfService bool
}

// DecodeStatusFlags decodes a byte to StatusFlags
func DecodeStatusFlags(b byte) StatusFlags {
	return StatusFlags{
		InAlarm:      b&0x08 != 0,
		Fault:        b&0x04 != 0,
		Overridden:   b&0x02 != 0,
		OutOfService: b&0x01 != 0,
	}
}

func (s StatusFlags) String() string {
	return fmt.Sprintf("{in-alarm:%v, fault:%v, overridden:%v, out-of-service:%v}",
		s.InAlarm, s.Fault, s.Overridden, s.OutOfService)
}

// EventState represents the BACnet event state
type EventState uint8

const (
	EventStateNormal       EventState = 0
	EventStateFault        EventState = 1
	EventStateOffNormal    EventState = 2
	EventStateHighLimit    EventState = 3
	EventStateLowLimit     EventState = 4
	EventStateLifeSafetyAlarm EventState = 5
)

func (e EventState) String() string {
	names := map[EventState]string{
		EventStateNormal:       "normal",
		EventStateFault:        "fault",
		EventStateOffNormal:    "off-normal",
		EventStateHighLimit:    "high-limit",
		EventStateLowLimit:     "low-limit",
		EventStateLifeSafetyAlarm: "life-safety-alarm",
	}
	if name, ok := names[e]; ok {
		return name
	}
	return fmt.Sprintf("event-state(%d)", e)
}

// Reliability represents the BACnet reliability
type Reliability uint8

const (
	ReliabilityNoFaultDetected            Reliability = 0
	ReliabilityNoSensor                   Reliability = 1
	ReliabilityOverRange                  Reliability = 2
	ReliabilityUnderRange                 Reliability = 3
	ReliabilityOpenLoop                   Reliability = 4
	ReliabilityShortedLoop                Reliability = 5
	ReliabilityNoOutput                   Reliability = 6
	ReliabilityUnreliableOther            Reliability = 7
	ReliabilityProcessError               Reliability = 8
	ReliabilityMultiStateFault            Reliability = 9
	ReliabilityConfigurationError         Reliability = 10
	ReliabilityCommunicationFailure       Reliability = 12
	ReliabilityMemberFault                Reliability = 13
	ReliabilityMonitoredObjectFault       Reliability = 14
	ReliabilityTripped                    Reliability = 15
	ReliabilityLampFailure                Reliability = 16
	ReliabilityActivationFailure          Reliability = 17
	ReliabilityRenewDhcpFailure           Reliability = 18
	ReliabilityRenewFdRegistrationFailure Reliability = 19
	ReliabilityRestartAutoNegotiationFailure Reliability = 20
	ReliabilityRestartFailure             Reliability = 21
	ReliabilityProprietaryCommandFailure  Reliability = 22
	ReliabilityFaultsListed               Reliability = 23
	ReliabilityReferencedObjectFault      Reliability = 24
)

func (r Reliability) String() string {
	names := map[Reliability]string{
		ReliabilityNoFaultDetected:      "no-fault-detected",
		ReliabilityNoSensor:             "no-sensor",
		ReliabilityOverRange:            "over-range",
		ReliabilityUnderRange:           "under-range",
		ReliabilityOpenLoop:             "open-loop",
		ReliabilityShortedLoop:          "shorted-loop",
		ReliabilityNoOutput:             "no-output",
		ReliabilityUnreliableOther:      "unreliable-other",
		ReliabilityProcessError:         "process-error",
		ReliabilityMultiStateFault:      "multi-state-fault",
		ReliabilityConfigurationError:   "configuration-error",
		ReliabilityCommunicationFailure: "communication-failure",
	}
	if name, ok := names[r]; ok {
		return name
	}
	return fmt.Sprintf("reliability(%d)", r)
}

// EngineeringUnits represents BACnet engineering units
type EngineeringUnits uint16

const (
	UnitsSquareMeters           EngineeringUnits = 0
	UnitsSquareFeet             EngineeringUnits = 1
	UnitsMilliamperes           EngineeringUnits = 2
	UnitsAmperes                EngineeringUnits = 3
	UnitsOhms                   EngineeringUnits = 4
	UnitsVolts                  EngineeringUnits = 5
	UnitsKilovolts              EngineeringUnits = 6
	UnitsMegavolts              EngineeringUnits = 7
	UnitsVoltAmperes            EngineeringUnits = 8
	UnitsKilovoltAmperes        EngineeringUnits = 9
	UnitsMegavoltAmperes        EngineeringUnits = 10
	UnitsVoltAmperesReactive    EngineeringUnits = 11
	UnitsKilovoltAmperesReactive EngineeringUnits = 12
	UnitsMegavoltAmperesReactive EngineeringUnits = 13
	UnitsDegreesPhase           EngineeringUnits = 14
	UnitsPowerFactor            EngineeringUnits = 15
	UnitsJoules                 EngineeringUnits = 16
	UnitsKilojoules             EngineeringUnits = 17
	UnitsWattHours              EngineeringUnits = 18
	UnitsKilowattHours          EngineeringUnits = 19
	UnitsBtus                   EngineeringUnits = 20
	UnitsTherms                 EngineeringUnits = 21
	UnitsTonHours               EngineeringUnits = 22
	UnitsJoulesPerKilogramDryAir EngineeringUnits = 23
	UnitsBtusPerPoundDryAir     EngineeringUnits = 24
	UnitsCyclesPerHour          EngineeringUnits = 25
	UnitsCyclesPerMinute        EngineeringUnits = 26
	UnitsHertz                  EngineeringUnits = 27
	UnitsGramsOfWaterPerKilogramDryAir EngineeringUnits = 28
	UnitsPercentRelativeHumidity EngineeringUnits = 29
	UnitsMillimeters            EngineeringUnits = 30
	UnitsMeters                 EngineeringUnits = 31
	UnitsInches                 EngineeringUnits = 32
	UnitsFeet                   EngineeringUnits = 33
	UnitsWattsPerSquareFoot     EngineeringUnits = 34
	UnitsWattsPerSquareMeter    EngineeringUnits = 35
	UnitsLumens                 EngineeringUnits = 36
	UnitsLuxes                  EngineeringUnits = 37
	UnitsFootCandles            EngineeringUnits = 38
	UnitsKilograms              EngineeringUnits = 39
	UnitsPounds                 EngineeringUnits = 40
	UnitsWatts                  EngineeringUnits = 41
	UnitsKilowatts              EngineeringUnits = 42
	UnitsMegawatts              EngineeringUnits = 43
	UnitsBtusPerHour            EngineeringUnits = 44
	UnitsHorsepower             EngineeringUnits = 45
	UnitsTonsRefrigeration      EngineeringUnits = 46
	UnitsPascals                EngineeringUnits = 47
	UnitsKilopascals            EngineeringUnits = 48
	UnitsBars                   EngineeringUnits = 49
	UnitsPoundsForcePerSquareInch EngineeringUnits = 50
	UnitsCentimetersOfWater     EngineeringUnits = 51
	UnitsInchesOfWater          EngineeringUnits = 52
	UnitsMillimetersOfMercury   EngineeringUnits = 53
	UnitsCentimetersOfMercury   EngineeringUnits = 54
	UnitsInchesOfMercury        EngineeringUnits = 55
	UnitsDegreesCelsius         EngineeringUnits = 62
	UnitsDegreesKelvin          EngineeringUnits = 63
	UnitsDegreesFahrenheit      EngineeringUnits = 64
	UnitsDegreeDaysCelsius      EngineeringUnits = 65
	UnitsDegreeDaysFahrenheit   EngineeringUnits = 66
	UnitsYears                  EngineeringUnits = 67
	UnitsMonths                 EngineeringUnits = 68
	UnitsWeeks                  EngineeringUnits = 69
	UnitsDays                   EngineeringUnits = 70
	UnitsHours                  EngineeringUnits = 71
	UnitsMinutes                EngineeringUnits = 72
	UnitsSeconds                EngineeringUnits = 73
	UnitsMetersPerSecond        EngineeringUnits = 74
	UnitsKilometersPerHour      EngineeringUnits = 75
	UnitsFeetPerSecond          EngineeringUnits = 76
	UnitsFeetPerMinute          EngineeringUnits = 77
	UnitsMilesPerHour           EngineeringUnits = 78
	UnitsCubicFeet              EngineeringUnits = 79
	UnitsCubicMeters            EngineeringUnits = 80
	UnitsImperialGallons        EngineeringUnits = 81
	UnitsLiters                 EngineeringUnits = 82
	UnitsUsGallons              EngineeringUnits = 83
	UnitsCubicFeetPerMinute     EngineeringUnits = 84
	UnitsCubicMetersPerSecond   EngineeringUnits = 85
	UnitsImperialGallonsPerMinute EngineeringUnits = 86
	UnitsLitersPerSecond        EngineeringUnits = 87
	UnitsLitersPerMinute        EngineeringUnits = 88
	UnitsUsGallonsPerMinute     EngineeringUnits = 89
	UnitsDegreesAngular         EngineeringUnits = 90
	UnitsDegreesCelsiusPerHour  EngineeringUnits = 91
	UnitsDegreesCelsiusPerMinute EngineeringUnits = 92
	UnitsDegreesFahrenheitPerHour EngineeringUnits = 93
	UnitsDegreesFahrenheitPerMinute EngineeringUnits = 94
	UnitsNoUnits                EngineeringUnits = 95
	UnitsPartsPerMillion        EngineeringUnits = 96
	UnitsPartsPerBillion        EngineeringUnits = 97
	UnitsPercent                EngineeringUnits = 98
	UnitsPercentPerSecond       EngineeringUnits = 99
	UnitsPerMinute              EngineeringUnits = 100
	UnitsPerSecond              EngineeringUnits = 101
	UnitsPsiPerDegreeFahrenheit EngineeringUnits = 102
	UnitsRadians                EngineeringUnits = 103
	UnitsRevolutionsPerMinute   EngineeringUnits = 104
)

func (u EngineeringUnits) String() string {
	names := map[EngineeringUnits]string{
		UnitsDegreesCelsius:    "°C",
		UnitsDegreesFahrenheit: "°F",
		UnitsDegreesKelvin:     "K",
		UnitsPercent:           "%",
		UnitsPercentRelativeHumidity: "%RH",
		UnitsMeters:            "m",
		UnitsFeet:              "ft",
		UnitsMillimeters:       "mm",
		UnitsInches:            "in",
		UnitsVolts:             "V",
		UnitsAmperes:           "A",
		UnitsMilliamperes:      "mA",
		UnitsWatts:             "W",
		UnitsKilowatts:         "kW",
		UnitsMegawatts:         "MW",
		UnitsKilowattHours:     "kWh",
		UnitsHertz:             "Hz",
		UnitsPascals:           "Pa",
		UnitsKilopascals:       "kPa",
		UnitsBars:              "bar",
		UnitsLiters:            "L",
		UnitsCubicMeters:       "m³",
		UnitsLitersPerSecond:   "L/s",
		UnitsLitersPerMinute:   "L/min",
		UnitsMetersPerSecond:   "m/s",
		UnitsSeconds:           "s",
		UnitsMinutes:           "min",
		UnitsHours:             "h",
		UnitsDays:              "d",
		UnitsNoUnits:           "",
	}
	if name, ok := names[u]; ok {
		return name
	}
	return fmt.Sprintf("units(%d)", u)
}

// Segmentation represents the BACnet segmentation capability
type Segmentation uint8

const (
	SegmentationBoth          Segmentation = 0
	SegmentationTransmit      Segmentation = 1
	SegmentationReceive       Segmentation = 2
	SegmentationNone          Segmentation = 3
)

func (s Segmentation) String() string {
	names := map[Segmentation]string{
		SegmentationBoth:     "segmented-both",
		SegmentationTransmit: "segmented-transmit",
		SegmentationReceive:  "segmented-receive",
		SegmentationNone:     "no-segmentation",
	}
	if name, ok := names[s]; ok {
		return name
	}
	return fmt.Sprintf("segmentation(%d)", s)
}

// DeviceStatus represents the BACnet device status
type DeviceStatus uint8

const (
	DeviceStatusOperational         DeviceStatus = 0
	DeviceStatusOperationalReadOnly DeviceStatus = 1
	DeviceStatusDownloadRequired    DeviceStatus = 2
	DeviceStatusDownloadInProgress  DeviceStatus = 3
	DeviceStatusNonOperational      DeviceStatus = 4
	DeviceStatusBackupInProgress    DeviceStatus = 5
)

func (d DeviceStatus) String() string {
	names := map[DeviceStatus]string{
		DeviceStatusOperational:         "operational",
		DeviceStatusOperationalReadOnly: "operational-read-only",
		DeviceStatusDownloadRequired:    "download-required",
		DeviceStatusDownloadInProgress:  "download-in-progress",
		DeviceStatusNonOperational:      "non-operational",
		DeviceStatusBackupInProgress:    "backup-in-progress",
	}
	if name, ok := names[d]; ok {
		return name
	}
	return fmt.Sprintf("device-status(%d)", d)
}

// Address represents a BACnet address
type Address struct {
	Net  uint16
	Addr []byte
}

// DeviceInfo represents information about a BACnet device
type DeviceInfo struct {
	ObjectID            ObjectIdentifier
	Address             Address
	MaxAPDULength       uint16
	Segmentation        Segmentation
	VendorID            uint16
	VendorName          string
	ModelName           string
	FirmwareRevision    string
	ApplicationSoftware string
	Description         string
	Location            string
	ObjectList          []ObjectIdentifier
}

// PropertyValue represents a property value with metadata
type PropertyValue struct {
	ObjectID   ObjectIdentifier
	PropertyID PropertyIdentifier
	ArrayIndex *uint32
	Value      interface{}
	Priority   *uint8
}

// ReadPropertyRequest represents a ReadProperty request
type ReadPropertyRequest struct {
	ObjectID   ObjectIdentifier
	PropertyID PropertyIdentifier
	ArrayIndex *uint32
}

// WritePropertyRequest represents a WriteProperty request
type WritePropertyRequest struct {
	ObjectID   ObjectIdentifier
	PropertyID PropertyIdentifier
	ArrayIndex *uint32
	Value      interface{}
	Priority   *uint8
}

// Tag types for BACnet encoding
type TagClass uint8

const (
	TagClassApplication TagClass = 0
	TagClassContext     TagClass = 1
)

type ApplicationTag uint8

const (
	TagNull            ApplicationTag = 0
	TagBoolean         ApplicationTag = 1
	TagUnsignedInt     ApplicationTag = 2
	TagSignedInt       ApplicationTag = 3
	TagReal            ApplicationTag = 4
	TagDouble          ApplicationTag = 5
	TagOctetString     ApplicationTag = 6
	TagCharacterString ApplicationTag = 7
	TagBitString       ApplicationTag = 8
	TagEnumerated      ApplicationTag = 9
	TagDate            ApplicationTag = 10
	TagTime            ApplicationTag = 11
	TagObjectID        ApplicationTag = 12
)

// Helper functions for encoding
func encodeUint16(buf []byte, v uint16) {
	binary.BigEndian.PutUint16(buf, v)
}

func encodeUint32(buf []byte, v uint32) {
	binary.BigEndian.PutUint32(buf, v)
}

func decodeUint16(buf []byte) uint16 {
	return binary.BigEndian.Uint16(buf)
}

func decodeUint32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}
