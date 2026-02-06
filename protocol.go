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
	"encoding/binary"
	"fmt"
	"math"
)

// BVLC Header (BACnet Virtual Link Control)
type BVLCHeader struct {
	Type     BVLCType
	Function BVLCFunction
	Length   uint16
}

// EncodeBVLC encodes a BVLC header
func EncodeBVLC(function BVLCFunction, npduLength int) []byte {
	totalLength := 4 + npduLength // BVLC header is 4 bytes
	buf := make([]byte, 4)
	buf[0] = byte(BVLCTypeBACnetIP)
	buf[1] = byte(function)
	binary.BigEndian.PutUint16(buf[2:], uint16(totalLength))
	return buf
}

// DecodeBVLC decodes a BVLC header
func DecodeBVLC(data []byte) (*BVLCHeader, error) {
	if len(data) < 4 {
		return nil, ErrInvalidBVLC
	}
	return &BVLCHeader{
		Type:     BVLCType(data[0]),
		Function: BVLCFunction(data[1]),
		Length:   binary.BigEndian.Uint16(data[2:4]),
	}, nil
}

// NPDU (Network Protocol Data Unit)
type NPDU struct {
	Version           uint8
	Control           NPDUControl
	DestNet           uint16
	DestAddr          []byte
	DestHopCount      uint8
	SrcNet            uint16
	SrcAddr           []byte
	MessageType       NetworkMessageType
	VendorID          uint16
	Data              []byte
}

// EncodeNPDU encodes an NPDU for unicast without routing
func EncodeNPDU(expectingReply bool, priority NPDUControl) []byte {
	control := priority
	if expectingReply {
		control |= NPDUControlExpectingReply
	}
	return []byte{
		0x01, // Version
		byte(control),
	}
}

// EncodeNPDUWithDest encodes an NPDU with destination address
func EncodeNPDUWithDest(destNet uint16, destAddr []byte, hopCount uint8, expectingReply bool, priority NPDUControl) []byte {
	control := priority | NPDUControlDestSpecifier
	if expectingReply {
		control |= NPDUControlExpectingReply
	}

	buf := make([]byte, 0, 8+len(destAddr))
	buf = append(buf, 0x01) // Version
	buf = append(buf, byte(control))
	buf = append(buf, byte(destNet>>8), byte(destNet))
	buf = append(buf, byte(len(destAddr)))
	buf = append(buf, destAddr...)
	buf = append(buf, hopCount)

	return buf
}

// DecodeNPDU decodes an NPDU
func DecodeNPDU(data []byte) (*NPDU, int, error) {
	if len(data) < 2 {
		return nil, 0, ErrInvalidNPDU
	}

	npdu := &NPDU{
		Version: data[0],
		Control: NPDUControl(data[1]),
	}

	if npdu.Version != 0x01 {
		return nil, 0, fmt.Errorf("%w: unsupported version %d", ErrInvalidNPDU, npdu.Version)
	}

	offset := 2

	// Destination specifier
	if npdu.Control&NPDUControlDestSpecifier != 0 {
		if len(data) < offset+3 {
			return nil, 0, ErrInvalidNPDU
		}
		npdu.DestNet = binary.BigEndian.Uint16(data[offset:])
		offset += 2

		addrLen := int(data[offset])
		offset++

		if len(data) < offset+addrLen+1 {
			return nil, 0, ErrInvalidNPDU
		}
		npdu.DestAddr = make([]byte, addrLen)
		copy(npdu.DestAddr, data[offset:offset+addrLen])
		offset += addrLen

		npdu.DestHopCount = data[offset]
		offset++
	}

	// Source specifier
	if npdu.Control&NPDUControlSourceSpecifier != 0 {
		if len(data) < offset+3 {
			return nil, 0, ErrInvalidNPDU
		}
		npdu.SrcNet = binary.BigEndian.Uint16(data[offset:])
		offset += 2

		addrLen := int(data[offset])
		offset++

		if len(data) < offset+addrLen {
			return nil, 0, ErrInvalidNPDU
		}
		npdu.SrcAddr = make([]byte, addrLen)
		copy(npdu.SrcAddr, data[offset:offset+addrLen])
		offset += addrLen
	}

	// Network layer message
	if npdu.Control&NPDUControlNetworkLayerMessage != 0 {
		if len(data) < offset+1 {
			return nil, 0, ErrInvalidNPDU
		}
		npdu.MessageType = NetworkMessageType(data[offset])
		offset++

		// Vendor-specific message types have vendor ID
		if npdu.MessageType >= 0x80 {
			if len(data) < offset+2 {
				return nil, 0, ErrInvalidNPDU
			}
			npdu.VendorID = binary.BigEndian.Uint16(data[offset:])
			offset += 2
		}
	}

	npdu.Data = data[offset:]
	return npdu, offset, nil
}

// APDU Types
type APDU struct {
	Type         PDUType
	Segmented    bool
	MoreFollows  bool
	SegmentedAck bool
	MaxSegments  uint8
	MaxAPDU      uint8
	InvokeID     uint8
	SequenceNum  uint8
	WindowSize   uint8
	Service      uint8
	Data         []byte
}

// EncodeConfirmedRequest encodes a confirmed service request APDU
func EncodeConfirmedRequest(invokeID uint8, service ConfirmedServiceChoice, data []byte, maxSegments, maxAPDU uint8) []byte {
	buf := make([]byte, 0, 4+len(data))

	// PDU type and flags
	pduType := byte(PDUTypeConfirmedRequest)
	buf = append(buf, pduType)

	// Max segments and max APDU
	buf = append(buf, (maxSegments<<4)|maxAPDU)

	// Invoke ID
	buf = append(buf, invokeID)

	// Service choice
	buf = append(buf, byte(service))

	// Service data
	buf = append(buf, data...)

	return buf
}

// EncodeUnconfirmedRequest encodes an unconfirmed service request APDU
func EncodeUnconfirmedRequest(service UnconfirmedServiceChoice, data []byte) []byte {
	buf := make([]byte, 0, 2+len(data))
	buf = append(buf, byte(PDUTypeUnconfirmedRequest))
	buf = append(buf, byte(service))
	buf = append(buf, data...)
	return buf
}

// DecodeAPDU decodes an APDU
func DecodeAPDU(data []byte) (*APDU, error) {
	if len(data) < 1 {
		return nil, ErrInvalidAPDU
	}

	apdu := &APDU{
		Type: PDUType(data[0] & 0xF0),
	}

	switch apdu.Type {
	case PDUTypeConfirmedRequest:
		return decodeConfirmedRequest(data)
	case PDUTypeUnconfirmedRequest:
		return decodeUnconfirmedRequest(data)
	case PDUTypeSimpleAck:
		return decodeSimpleAck(data)
	case PDUTypeComplexAck:
		return decodeComplexAck(data)
	case PDUTypeError:
		return decodeErrorAPDU(data)
	case PDUTypeReject:
		return decodeRejectAPDU(data)
	case PDUTypeAbort:
		return decodeAbortAPDU(data)
	default:
		return nil, fmt.Errorf("%w: unknown PDU type %02x", ErrInvalidAPDU, apdu.Type)
	}
}

func decodeConfirmedRequest(data []byte) (*APDU, error) {
	if len(data) < 4 {
		return nil, ErrInvalidAPDU
	}

	apdu := &APDU{
		Type:        PDUTypeConfirmedRequest,
		Segmented:   data[0]&0x08 != 0,
		MoreFollows: data[0]&0x04 != 0,
		MaxSegments: (data[1] >> 4) & 0x07,
		MaxAPDU:     data[1] & 0x0F,
		InvokeID:    data[2],
		Service:     data[3],
		Data:        data[4:],
	}

	if apdu.Segmented {
		if len(data) < 6 {
			return nil, ErrInvalidAPDU
		}
		apdu.SequenceNum = data[4]
		apdu.WindowSize = data[5]
		apdu.Data = data[6:]
	}

	return apdu, nil
}

func decodeUnconfirmedRequest(data []byte) (*APDU, error) {
	if len(data) < 2 {
		return nil, ErrInvalidAPDU
	}

	return &APDU{
		Type:    PDUTypeUnconfirmedRequest,
		Service: data[1],
		Data:    data[2:],
	}, nil
}

func decodeSimpleAck(data []byte) (*APDU, error) {
	if len(data) < 3 {
		return nil, ErrInvalidAPDU
	}

	return &APDU{
		Type:     PDUTypeSimpleAck,
		InvokeID: data[1],
		Service:  data[2],
	}, nil
}

func decodeComplexAck(data []byte) (*APDU, error) {
	if len(data) < 3 {
		return nil, ErrInvalidAPDU
	}

	apdu := &APDU{
		Type:        PDUTypeComplexAck,
		Segmented:   data[0]&0x08 != 0,
		MoreFollows: data[0]&0x04 != 0,
		InvokeID:    data[1],
		Service:     data[2],
		Data:        data[3:],
	}

	if apdu.Segmented {
		if len(data) < 5 {
			return nil, ErrInvalidAPDU
		}
		apdu.SequenceNum = data[3]
		apdu.WindowSize = data[4]
		apdu.Data = data[5:]
	}

	return apdu, nil
}

func decodeErrorAPDU(data []byte) (*APDU, error) {
	if len(data) < 3 {
		return nil, ErrInvalidAPDU
	}

	return &APDU{
		Type:     PDUTypeError,
		InvokeID: data[1],
		Service:  data[2],
		Data:     data[3:],
	}, nil
}

func decodeRejectAPDU(data []byte) (*APDU, error) {
	if len(data) < 3 {
		return nil, ErrInvalidAPDU
	}

	return &APDU{
		Type:     PDUTypeReject,
		InvokeID: data[1],
		Service:  data[2], // Reject reason is in service field
	}, nil
}

func decodeAbortAPDU(data []byte) (*APDU, error) {
	if len(data) < 3 {
		return nil, ErrInvalidAPDU
	}

	return &APDU{
		Type:     PDUTypeAbort,
		InvokeID: data[1],
		Service:  data[2], // Abort reason is in service field
	}, nil
}

// Tag encoding/decoding helpers

// EncodeTag encodes a BACnet tag
func EncodeTag(tagNum uint8, class TagClass, length int) []byte {
	if length < 5 && tagNum < 15 {
		// Short form
		tag := (tagNum << 4) | (uint8(class) << 3) | uint8(length)
		return []byte{tag}
	}

	buf := make([]byte, 0, 6)

	// Extended tag number
	if tagNum >= 15 {
		buf = append(buf, 0xF0|(uint8(class)<<3))
		buf = append(buf, tagNum)
	} else {
		buf = append(buf, (tagNum<<4)|(uint8(class)<<3)|0x05)
	}

	// Extended length
	if length >= 5 {
		if length < 254 {
			buf = append(buf, byte(length))
		} else if length < 65536 {
			buf = append(buf, 254)
			buf = append(buf, byte(length>>8), byte(length))
		} else {
			buf = append(buf, 255)
			buf = append(buf, byte(length>>24), byte(length>>16), byte(length>>8), byte(length))
		}
	}

	return buf
}

// EncodeContextTag encodes a context-specific tag
func EncodeContextTag(tagNum uint8, data []byte) []byte {
	tag := EncodeTag(tagNum, TagClassContext, len(data))
	return append(tag, data...)
}

// EncodeOpeningTag encodes an opening tag for constructed data
func EncodeOpeningTag(tagNum uint8) []byte {
	if tagNum < 15 {
		return []byte{(tagNum << 4) | 0x0E}
	}
	return []byte{0xFE, tagNum}
}

// EncodeClosingTag encodes a closing tag for constructed data
func EncodeClosingTag(tagNum uint8) []byte {
	if tagNum < 15 {
		return []byte{(tagNum << 4) | 0x0F}
	}
	return []byte{0xFF, tagNum}
}

// EncodeUnsigned encodes an unsigned integer
func EncodeUnsigned(value uint32) []byte {
	if value < 0x100 {
		return []byte{byte(value)}
	} else if value < 0x10000 {
		return []byte{byte(value >> 8), byte(value)}
	} else if value < 0x1000000 {
		return []byte{byte(value >> 16), byte(value >> 8), byte(value)}
	}
	return []byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)}
}

// EncodeUnsignedTag encodes an unsigned integer with application tag
func EncodeUnsignedTag(value uint32) []byte {
	data := EncodeUnsigned(value)
	tag := EncodeTag(uint8(TagUnsignedInt), TagClassApplication, len(data))
	return append(tag, data...)
}

// EncodeContextUnsigned encodes an unsigned integer with context tag
func EncodeContextUnsigned(tagNum uint8, value uint32) []byte {
	data := EncodeUnsigned(value)
	return EncodeContextTag(tagNum, data)
}

// EncodeSigned encodes a signed integer
func EncodeSigned(value int32) []byte {
	if value >= -128 && value < 128 {
		return []byte{byte(value)}
	} else if value >= -32768 && value < 32768 {
		return []byte{byte(value >> 8), byte(value)}
	} else if value >= -8388608 && value < 8388608 {
		return []byte{byte(value >> 16), byte(value >> 8), byte(value)}
	}
	return []byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)}
}

// EncodeReal encodes a float32
func EncodeReal(value float32) []byte {
	bits := math.Float32bits(value)
	return []byte{byte(bits >> 24), byte(bits >> 16), byte(bits >> 8), byte(bits)}
}

// EncodeRealTag encodes a float32 with application tag
func EncodeRealTag(value float32) []byte {
	data := EncodeReal(value)
	tag := EncodeTag(uint8(TagReal), TagClassApplication, 4)
	return append(tag, data...)
}

// EncodeDouble encodes a float64
func EncodeDouble(value float64) []byte {
	bits := math.Float64bits(value)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, bits)
	return buf
}

// EncodeBooleanTag encodes a boolean with application tag
func EncodeBooleanTag(value bool) []byte {
	if value {
		return []byte{0x11} // Boolean true, length 1, value 1
	}
	return []byte{0x10} // Boolean false, length 1, value 0
}

// EncodeContextBoolean encodes a boolean with context tag
func EncodeContextBoolean(tagNum uint8, value bool) []byte {
	v := byte(0)
	if value {
		v = 1
	}
	return EncodeContextTag(tagNum, []byte{v})
}

// EncodeEnumerated encodes an enumerated value
func EncodeEnumerated(value uint32) []byte {
	return EncodeUnsigned(value)
}

// EncodeEnumeratedTag encodes an enumerated value with application tag
func EncodeEnumeratedTag(value uint32) []byte {
	data := EncodeEnumerated(value)
	tag := EncodeTag(uint8(TagEnumerated), TagClassApplication, len(data))
	return append(tag, data...)
}

// EncodeContextEnumerated encodes an enumerated value with context tag
func EncodeContextEnumerated(tagNum uint8, value uint32) []byte {
	data := EncodeEnumerated(value)
	return EncodeContextTag(tagNum, data)
}

// EncodeObjectIdentifier encodes an object identifier
func EncodeObjectIdentifier(oid ObjectIdentifier) []byte {
	value := oid.Encode()
	return []byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)}
}

// EncodeObjectIdentifierTag encodes an object identifier with application tag
func EncodeObjectIdentifierTag(oid ObjectIdentifier) []byte {
	data := EncodeObjectIdentifier(oid)
	tag := EncodeTag(uint8(TagObjectID), TagClassApplication, 4)
	return append(tag, data...)
}

// EncodeContextObjectIdentifier encodes an object identifier with context tag
func EncodeContextObjectIdentifier(tagNum uint8, oid ObjectIdentifier) []byte {
	data := EncodeObjectIdentifier(oid)
	return EncodeContextTag(tagNum, data)
}

// EncodeCharacterString encodes a character string (UTF-8)
func EncodeCharacterString(s string) []byte {
	// Character set 0 = UTF-8
	data := make([]byte, 1+len(s))
	data[0] = 0 // UTF-8 encoding
	copy(data[1:], s)
	return data
}

// EncodeCharacterStringTag encodes a character string with application tag
func EncodeCharacterStringTag(s string) []byte {
	data := EncodeCharacterString(s)
	tag := EncodeTag(uint8(TagCharacterString), TagClassApplication, len(data))
	return append(tag, data...)
}

// DecodeTagNumber decodes a tag from data
func DecodeTagNumber(data []byte) (tagNum uint8, class TagClass, length int, headerLen int, err error) {
	if len(data) < 1 {
		return 0, 0, 0, 0, ErrInvalidAPDU
	}

	tagNum = (data[0] >> 4) & 0x0F
	class = TagClass((data[0] >> 3) & 0x01)
	length = int(data[0] & 0x07)
	headerLen = 1

	// Extended tag number
	if tagNum == 0x0F {
		if len(data) < 2 {
			return 0, 0, 0, 0, ErrInvalidAPDU
		}
		tagNum = data[1]
		headerLen = 2
	}

	// Opening/closing tag
	if class == TagClassContext && (data[0]&0x07) == 0x06 {
		// Opening tag
		length = -1
		return
	}
	if class == TagClassContext && (data[0]&0x07) == 0x07 {
		// Closing tag
		length = -2
		return
	}

	// Extended length
	if length == 5 {
		if len(data) < headerLen+1 {
			return 0, 0, 0, 0, ErrInvalidAPDU
		}
		if data[headerLen] < 254 {
			length = int(data[headerLen])
			headerLen++
		} else if data[headerLen] == 254 {
			if len(data) < headerLen+3 {
				return 0, 0, 0, 0, ErrInvalidAPDU
			}
			length = int(binary.BigEndian.Uint16(data[headerLen+1:]))
			headerLen += 3
		} else {
			if len(data) < headerLen+5 {
				return 0, 0, 0, 0, ErrInvalidAPDU
			}
			length = int(binary.BigEndian.Uint32(data[headerLen+1:]))
			headerLen += 5
		}
	}

	return tagNum, class, length, headerLen, nil
}

// DecodeUnsigned decodes an unsigned integer from data
func DecodeUnsigned(data []byte) uint32 {
	switch len(data) {
	case 1:
		return uint32(data[0])
	case 2:
		return uint32(binary.BigEndian.Uint16(data))
	case 3:
		return uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
	case 4:
		return binary.BigEndian.Uint32(data)
	default:
		return 0
	}
}

// DecodeSigned decodes a signed integer from data
func DecodeSigned(data []byte) int32 {
	switch len(data) {
	case 1:
		return int32(int8(data[0]))
	case 2:
		return int32(int16(binary.BigEndian.Uint16(data)))
	case 3:
		v := uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
		if data[0]&0x80 != 0 {
			v |= 0xFF000000
		}
		return int32(v)
	case 4:
		return int32(binary.BigEndian.Uint32(data))
	default:
		return 0
	}
}

// DecodeReal decodes a float32 from data
func DecodeReal(data []byte) float32 {
	if len(data) != 4 {
		return 0
	}
	bits := binary.BigEndian.Uint32(data)
	return math.Float32frombits(bits)
}

// DecodeDouble decodes a float64 from data
func DecodeDouble(data []byte) float64 {
	if len(data) != 8 {
		return 0
	}
	bits := binary.BigEndian.Uint64(data)
	return math.Float64frombits(bits)
}

// DecodeCharacterString decodes a character string
func DecodeCharacterString(data []byte) string {
	if len(data) < 1 {
		return ""
	}
	// Skip character set byte
	return string(data[1:])
}

// DecodeObjectIdentifierFromBytes decodes an object identifier from bytes
func DecodeObjectIdentifierFromBytes(data []byte) ObjectIdentifier {
	if len(data) != 4 {
		return ObjectIdentifier{}
	}
	value := binary.BigEndian.Uint32(data)
	return DecodeObjectIdentifier(value)
}
