package bacnet

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/edgeo/drivers/bacnet/bacnet/internal/transport"
)

// ConnectionState represents the client connection state
type ConnectionState int32

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	default:
		return "unknown"
	}
}

// Client is a BACnet/IP client
type Client struct {
	opts      *clientOptions
	transport *transport.UDPTransport

	state    atomic.Int32
	invokeID atomic.Uint32

	// Pending requests
	pendingMu  sync.RWMutex
	pending    map[uint8]chan *APDU

	// Discovered devices
	devicesMu sync.RWMutex
	devices   map[uint32]*DeviceInfo

	// COV subscriptions
	covMu     sync.RWMutex
	covSubs   map[uint32]COVHandler

	// Metrics
	metrics *Metrics

	// Logger
	logger *slog.Logger

	// Receiver goroutine
	receiverCtx    context.Context
	receiverCancel context.CancelFunc
	receiverDone   chan struct{}
}

// COVHandler is called when a COV notification is received
type COVHandler func(deviceID uint32, objectID ObjectIdentifier, values []PropertyValue)

// NewClient creates a new BACnet client
func NewClient(opts ...Option) (*Client, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	c := &Client{
		opts:     options,
		pending:  make(map[uint8]chan *APDU),
		devices:  make(map[uint32]*DeviceInfo),
		covSubs:  make(map[uint32]COVHandler),
		metrics:  NewMetrics(),
		logger:   options.logger,
	}

	// Create transport
	c.transport = transport.NewUDPTransport(options.localAddress)
	c.transport.SetReadTimeout(options.timeout)
	c.transport.SetWriteTimeout(options.timeout)

	return c, nil
}

// Connect opens the BACnet client connection
func (c *Client) Connect(ctx context.Context) error {
	if !c.state.CompareAndSwap(int32(StateDisconnected), int32(StateConnecting)) {
		return ErrAlreadyConnected
	}

	c.metrics.ConnectAttempts.Inc()

	if err := c.transport.Open(ctx); err != nil {
		c.state.Store(int32(StateDisconnected))
		c.metrics.ConnectFailures.Inc()
		return fmt.Errorf("open transport: %w", err)
	}

	// Start receiver goroutine
	c.receiverCtx, c.receiverCancel = context.WithCancel(context.Background())
	c.receiverDone = make(chan struct{})
	go c.receiver()

	c.state.Store(int32(StateConnected))
	c.metrics.ConnectSuccesses.Inc()

	c.logger.Info("connected",
		slog.String("local_addr", c.transport.LocalAddr().String()),
	)

	// Register as foreign device if BBMD is configured
	if c.opts.bbmdAddress != "" {
		if err := c.registerForeignDevice(ctx); err != nil {
			c.logger.Warn("failed to register as foreign device",
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

// Close closes the BACnet client connection
func (c *Client) Close() error {
	if c.state.Load() == int32(StateDisconnected) {
		return nil
	}

	c.state.Store(int32(StateDisconnected))
	c.metrics.Disconnects.Inc()

	// Stop receiver
	if c.receiverCancel != nil {
		c.receiverCancel()
		<-c.receiverDone
	}

	// Close pending requests
	c.pendingMu.Lock()
	for _, ch := range c.pending {
		close(ch)
	}
	c.pending = make(map[uint8]chan *APDU)
	c.pendingMu.Unlock()

	if err := c.transport.Close(); err != nil {
		return fmt.Errorf("close transport: %w", err)
	}

	c.logger.Info("disconnected")
	return nil
}

// State returns the current connection state
func (c *Client) State() ConnectionState {
	return ConnectionState(c.state.Load())
}

// Metrics returns the client metrics
func (c *Client) Metrics() *Metrics {
	return c.metrics
}

// nextInvokeID returns the next invoke ID
func (c *Client) nextInvokeID() uint8 {
	return uint8(c.invokeID.Add(1) & 0xFF)
}

// receiver handles incoming packets
func (c *Client) receiver() {
	defer close(c.receiverDone)

	for {
		select {
		case <-c.receiverCtx.Done():
			return
		default:
		}

		data, addr, err := c.transport.ReceiveWithTimeout(100 * time.Millisecond)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if c.transport.IsClosed() {
				return
			}
			c.logger.Debug("receive error", slog.String("error", err.Error()))
			continue
		}

		c.metrics.BytesReceived.Add(int64(len(data)))
		c.metrics.RecordActivity()

		go c.handlePacket(data, addr)
	}
}

// handlePacket processes an incoming packet
func (c *Client) handlePacket(data []byte, addr *net.UDPAddr) {
	// Decode BVLC header
	bvlc, err := DecodeBVLC(data)
	if err != nil {
		c.logger.Debug("invalid BVLC", slog.String("error", err.Error()))
		return
	}

	// Get NPDU data
	npduData := data[4:]
	if bvlc.Function == BVLCForwardedNPDU {
		// Skip forwarded address (6 bytes)
		if len(npduData) < 6 {
			return
		}
		npduData = npduData[6:]
	}

	// Decode NPDU
	npdu, offset, err := DecodeNPDU(npduData)
	if err != nil {
		c.logger.Debug("invalid NPDU", slog.String("error", err.Error()))
		return
	}

	// Skip network layer messages
	if npdu.Control&NPDUControlNetworkLayerMessage != 0 {
		return
	}

	// Decode APDU
	apduData := npduData[offset:]
	apdu, err := DecodeAPDU(apduData)
	if err != nil {
		c.logger.Debug("invalid APDU", slog.String("error", err.Error()))
		return
	}

	c.metrics.ResponsesReceived.Inc()

	// Handle based on PDU type
	switch apdu.Type {
	case PDUTypeUnconfirmedRequest:
		c.handleUnconfirmedRequest(apdu, addr, npdu)

	case PDUTypeSimpleAck, PDUTypeComplexAck:
		c.handleResponse(apdu)

	case PDUTypeError:
		c.metrics.ErrorsReceived.Inc()
		c.handleResponse(apdu)

	case PDUTypeReject:
		c.metrics.RejectsReceived.Inc()
		c.handleResponse(apdu)

	case PDUTypeAbort:
		c.metrics.AbortsReceived.Inc()
		c.handleResponse(apdu)
	}
}

// handleUnconfirmedRequest handles unconfirmed service requests
func (c *Client) handleUnconfirmedRequest(apdu *APDU, addr *net.UDPAddr, npdu *NPDU) {
	switch UnconfirmedServiceChoice(apdu.Service) {
	case ServiceIAm:
		c.handleIAm(apdu.Data, addr, npdu)

	case ServiceUnconfirmedCOVNotification:
		c.handleCOVNotification(apdu.Data)
	}
}

// handleIAm handles I-Am responses
func (c *Client) handleIAm(data []byte, addr *net.UDPAddr, npdu *NPDU) {
	c.metrics.IAmReceived.Inc()

	if len(data) < 4 {
		return
	}

	// Decode device object identifier
	tagNum, _, length, headerLen, err := DecodeTagNumber(data)
	if err != nil || tagNum != uint8(TagObjectID) || length != 4 {
		return
	}

	oidValue := binary.BigEndian.Uint32(data[headerLen:])
	oid := DecodeObjectIdentifier(oidValue)

	if oid.Type != ObjectTypeDevice {
		return
	}

	offset := headerLen + 4

	// Decode max APDU length
	if len(data) < offset+1 {
		return
	}
	tagNum, _, length, headerLen, err = DecodeTagNumber(data[offset:])
	if err != nil {
		return
	}
	maxAPDU := uint16(DecodeUnsigned(data[offset+headerLen : offset+headerLen+length]))
	offset += headerLen + length

	// Decode segmentation supported
	if len(data) < offset+1 {
		return
	}
	tagNum, _, length, headerLen, err = DecodeTagNumber(data[offset:])
	if err != nil {
		return
	}
	segmentation := Segmentation(DecodeUnsigned(data[offset+headerLen : offset+headerLen+length]))
	offset += headerLen + length

	// Decode vendor ID
	if len(data) < offset+1 {
		return
	}
	tagNum, _, length, headerLen, err = DecodeTagNumber(data[offset:])
	if err != nil {
		return
	}
	vendorID := uint16(DecodeUnsigned(data[offset+headerLen : offset+headerLen+length]))

	// Build device address
	var deviceAddr Address
	if npdu.Control&NPDUControlSourceSpecifier != 0 {
		deviceAddr = Address{
			Net:  npdu.SrcNet,
			Addr: npdu.SrcAddr,
		}
	} else {
		deviceAddr = Address{
			Net:  0,
			Addr: addr.IP.To4(),
		}
	}

	device := &DeviceInfo{
		ObjectID:      oid,
		Address:       deviceAddr,
		MaxAPDULength: maxAPDU,
		Segmentation:  segmentation,
		VendorID:      vendorID,
	}

	c.devicesMu.Lock()
	_, exists := c.devices[oid.Instance]
	c.devices[oid.Instance] = device
	c.devicesMu.Unlock()

	if !exists {
		c.metrics.DevicesDiscovered.Inc()
	}

	c.logger.Debug("device discovered",
		slog.Uint64("device_id", uint64(oid.Instance)),
		slog.String("address", addr.String()),
		slog.Uint64("vendor_id", uint64(vendorID)),
	)
}

// handleCOVNotification handles COV notification
func (c *Client) handleCOVNotification(data []byte) {
	c.metrics.COVNotifications.Inc()
	// TODO: Decode and dispatch to registered handlers
}

// handleResponse handles a response to a pending request
func (c *Client) handleResponse(apdu *APDU) {
	c.pendingMu.RLock()
	ch, ok := c.pending[apdu.InvokeID]
	c.pendingMu.RUnlock()

	if ok {
		select {
		case ch <- apdu:
		default:
		}
	}
}

// sendRequest sends a confirmed request and waits for response
func (c *Client) sendRequest(ctx context.Context, addr *net.UDPAddr, service ConfirmedServiceChoice, data []byte) (*APDU, error) {
	if c.State() != StateConnected {
		return nil, ErrNotConnected
	}

	invokeID := c.nextInvokeID()

	// Create response channel
	respCh := make(chan *APDU, 1)
	c.pendingMu.Lock()
	c.pending[invokeID] = respCh
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, invokeID)
		c.pendingMu.Unlock()
	}()

	// Encode APDU
	apdu := EncodeConfirmedRequest(invokeID, service, data, 0, 5)

	// Encode NPDU
	npdu := EncodeNPDU(true, NPDUControlPriorityNormal)

	// Encode BVLC
	bvlc := EncodeBVLC(BVLCOriginalUnicastNPDU, len(npdu)+len(apdu))

	// Build packet
	packet := make([]byte, 0, len(bvlc)+len(npdu)+len(apdu))
	packet = append(packet, bvlc...)
	packet = append(packet, npdu...)
	packet = append(packet, apdu...)

	// Send request
	start := time.Now()
	c.metrics.RequestsSent.Inc()
	c.metrics.ActiveRequests.Inc()
	defer c.metrics.ActiveRequests.Dec()

	if err := c.transport.Send(ctx, addr, packet); err != nil {
		c.metrics.RequestsFailed.Inc()
		return nil, fmt.Errorf("send request: %w", err)
	}

	c.metrics.BytesSent.Add(int64(len(packet)))

	// Wait for response
	select {
	case <-ctx.Done():
		c.metrics.RequestsTimedOut.Inc()
		return nil, ErrTimeout

	case resp, ok := <-respCh:
		c.metrics.RequestLatency.Record(time.Since(start))

		if !ok {
			return nil, ErrConnectionClosed
		}

		switch resp.Type {
		case PDUTypeSimpleAck, PDUTypeComplexAck:
			c.metrics.RequestsSucceeded.Inc()
			return resp, nil

		case PDUTypeError:
			c.metrics.RequestsFailed.Inc()
			return nil, c.decodeError(resp.Data)

		case PDUTypeReject:
			c.metrics.RequestsFailed.Inc()
			return nil, &RejectError{
				InvokeID: resp.InvokeID,
				Reason:   RejectReason(resp.Service),
			}

		case PDUTypeAbort:
			c.metrics.RequestsFailed.Inc()
			return nil, &AbortError{
				InvokeID: resp.InvokeID,
				Reason:   AbortReason(resp.Service),
			}

		default:
			return nil, fmt.Errorf("%w: unexpected PDU type %02x", ErrInvalidResponse, resp.Type)
		}
	}
}

// decodeError decodes a BACnet error response
func (c *Client) decodeError(data []byte) error {
	if len(data) < 2 {
		return ErrInvalidResponse
	}

	// Decode error class
	_, _, length, headerLen, err := DecodeTagNumber(data)
	if err != nil {
		return ErrInvalidResponse
	}
	errorClass := ErrorClass(DecodeUnsigned(data[headerLen : headerLen+length]))

	offset := headerLen + length

	// Decode error code
	_, _, length, headerLen, err = DecodeTagNumber(data[offset:])
	if err != nil {
		return ErrInvalidResponse
	}
	errorCode := ErrorCode(DecodeUnsigned(data[offset+headerLen : offset+headerLen+length]))

	return NewBACnetError(errorClass, errorCode)
}

// sendUnconfirmedRequest sends an unconfirmed request
func (c *Client) sendUnconfirmedRequest(ctx context.Context, addr *net.UDPAddr, broadcast bool, service UnconfirmedServiceChoice, data []byte) error {
	if c.State() != StateConnected {
		return ErrNotConnected
	}

	// Encode APDU
	apdu := EncodeUnconfirmedRequest(service, data)

	// Encode NPDU
	npdu := EncodeNPDU(false, NPDUControlPriorityNormal)

	// Encode BVLC
	var bvlcFunc BVLCFunction
	if broadcast {
		bvlcFunc = BVLCOriginalBroadcastNPDU
	} else {
		bvlcFunc = BVLCOriginalUnicastNPDU
	}
	bvlc := EncodeBVLC(bvlcFunc, len(npdu)+len(apdu))

	// Build packet
	packet := make([]byte, 0, len(bvlc)+len(npdu)+len(apdu))
	packet = append(packet, bvlc...)
	packet = append(packet, npdu...)
	packet = append(packet, apdu...)

	c.metrics.RequestsSent.Inc()

	var err error
	if broadcast {
		err = c.transport.Broadcast(ctx, DefaultPort, packet)
	} else {
		err = c.transport.Send(ctx, addr, packet)
	}

	if err != nil {
		c.metrics.RequestsFailed.Inc()
		return fmt.Errorf("send unconfirmed request: %w", err)
	}

	c.metrics.BytesSent.Add(int64(len(packet)))
	c.metrics.RequestsSucceeded.Inc()

	return nil
}

// registerForeignDevice registers as a foreign device with the BBMD
func (c *Client) registerForeignDevice(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", c.opts.bbmdAddress, c.opts.bbmdPort))
	if err != nil {
		return fmt.Errorf("resolve BBMD address: %w", err)
	}

	// TTL in seconds
	ttl := uint16(c.opts.foreignDeviceTTL.Seconds())

	// Build register foreign device request
	data := make([]byte, 6)
	data[0] = byte(BVLCTypeBACnetIP)
	data[1] = byte(BVLCRegisterForeignDevice)
	binary.BigEndian.PutUint16(data[2:], 6) // Length
	binary.BigEndian.PutUint16(data[4:], ttl)

	if err := c.transport.Send(ctx, addr, data); err != nil {
		return fmt.Errorf("send registration: %w", err)
	}

	c.logger.Info("registered as foreign device",
		slog.String("bbmd", addr.String()),
		slog.Duration("ttl", c.opts.foreignDeviceTTL),
	)

	return nil
}

// WhoIs sends a Who-Is request to discover devices
func (c *Client) WhoIs(ctx context.Context, opts ...DiscoverOption) ([]*DeviceInfo, error) {
	options := defaultDiscoverOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Build Who-Is request
	var data []byte
	if options.LowLimit != nil && options.HighLimit != nil {
		data = append(data, EncodeContextUnsigned(0, *options.LowLimit)...)
		data = append(data, EncodeContextUnsigned(1, *options.HighLimit)...)
	}

	// Send as broadcast
	if err := c.sendUnconfirmedRequest(ctx, nil, true, ServiceWhoIs, data); err != nil {
		return nil, err
	}

	c.metrics.WhoIsSent.Inc()

	// Wait for responses
	time.Sleep(options.Timeout)

	// Collect discovered devices
	c.devicesMu.RLock()
	devices := make([]*DeviceInfo, 0, len(c.devices))
	for _, dev := range c.devices {
		devices = append(devices, dev)
	}
	c.devicesMu.RUnlock()

	return devices, nil
}

// GetDevice returns information about a discovered device
func (c *Client) GetDevice(deviceID uint32) (*DeviceInfo, bool) {
	c.devicesMu.RLock()
	defer c.devicesMu.RUnlock()
	dev, ok := c.devices[deviceID]
	return dev, ok
}

// resolveDevice resolves a device ID to its address
func (c *Client) resolveDevice(ctx context.Context, deviceID uint32) (*net.UDPAddr, error) {
	c.devicesMu.RLock()
	dev, ok := c.devices[deviceID]
	c.devicesMu.RUnlock()

	if !ok {
		// Try to discover the device
		_, err := c.WhoIs(ctx, WithDeviceRange(deviceID, deviceID), WithDiscoveryTimeout(2*time.Second))
		if err != nil {
			return nil, err
		}

		c.devicesMu.RLock()
		dev, ok = c.devices[deviceID]
		c.devicesMu.RUnlock()

		if !ok {
			return nil, ErrDeviceNotFound
		}
	}

	// Convert device address to UDP address
	if len(dev.Address.Addr) == 4 {
		return &net.UDPAddr{
			IP:   net.IP(dev.Address.Addr),
			Port: DefaultPort,
		}, nil
	} else if len(dev.Address.Addr) == 6 {
		// IP + port format
		return &net.UDPAddr{
			IP:   net.IP(dev.Address.Addr[:4]),
			Port: int(binary.BigEndian.Uint16(dev.Address.Addr[4:])),
		}, nil
	}

	return nil, fmt.Errorf("invalid device address format")
}

// ReadProperty reads a property from a BACnet object
func (c *Client) ReadProperty(ctx context.Context, deviceID uint32, objectID ObjectIdentifier, propertyID PropertyIdentifier, opts ...ReadOption) (interface{}, error) {
	options := &ReadOptions{}
	for _, opt := range opts {
		opt(options)
	}

	addr, err := c.resolveDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	// Build ReadProperty request
	data := make([]byte, 0, 16)
	data = append(data, EncodeContextObjectIdentifier(0, objectID)...)
	data = append(data, EncodeContextEnumerated(1, uint32(propertyID))...)
	if options.ArrayIndex != nil {
		data = append(data, EncodeContextUnsigned(2, *options.ArrayIndex)...)
	}

	resp, err := c.sendRequest(ctx, addr, ServiceReadProperty, data)
	if err != nil {
		return nil, err
	}

	// Decode response
	return c.decodeReadPropertyResponse(resp.Data)
}

// decodeReadPropertyResponse decodes a ReadProperty response
func (c *Client) decodeReadPropertyResponse(data []byte) (interface{}, error) {
	if len(data) < 8 {
		return nil, ErrInvalidResponse
	}

	offset := 0

	// Skip object identifier [0]
	tagNum, class, length, headerLen, err := DecodeTagNumber(data[offset:])
	if err != nil || tagNum != 0 || class != TagClassContext {
		return nil, ErrInvalidResponse
	}
	offset += headerLen + length

	// Skip property identifier [1]
	tagNum, class, length, headerLen, err = DecodeTagNumber(data[offset:])
	if err != nil || tagNum != 1 || class != TagClassContext {
		return nil, ErrInvalidResponse
	}
	offset += headerLen + length

	// Check for optional array index [2]
	if len(data) > offset {
		tagNum, class, _, headerLen, err = DecodeTagNumber(data[offset:])
		if err == nil && tagNum == 2 && class == TagClassContext {
			offset += headerLen + length
		}
	}

	// Check for opening tag [3]
	if len(data) <= offset {
		return nil, ErrInvalidResponse
	}
	tagNum, class, length, _, err = DecodeTagNumber(data[offset:])
	if err != nil || tagNum != 3 || class != TagClassContext || length != -1 {
		return nil, ErrInvalidResponse
	}
	offset++

	// Decode property value
	return c.decodePropertyValue(data[offset:])
}

// decodePropertyValue decodes a property value
func (c *Client) decodePropertyValue(data []byte) (interface{}, error) {
	if len(data) < 1 {
		return nil, ErrInvalidResponse
	}

	tagNum, class, length, headerLen, err := DecodeTagNumber(data)
	if err != nil {
		return nil, err
	}

	// Check for closing tag
	if length == -2 {
		return nil, nil
	}

	if class == TagClassApplication {
		valueData := data[headerLen : headerLen+length]

		switch ApplicationTag(tagNum) {
		case TagNull:
			return nil, nil
		case TagBoolean:
			return length == 1, nil
		case TagUnsignedInt:
			return DecodeUnsigned(valueData), nil
		case TagSignedInt:
			return DecodeSigned(valueData), nil
		case TagReal:
			return DecodeReal(valueData), nil
		case TagDouble:
			return DecodeDouble(valueData), nil
		case TagOctetString:
			return valueData, nil
		case TagCharacterString:
			return DecodeCharacterString(valueData), nil
		case TagEnumerated:
			return DecodeUnsigned(valueData), nil
		case TagObjectID:
			oidValue := binary.BigEndian.Uint32(valueData)
			return DecodeObjectIdentifier(oidValue), nil
		default:
			return valueData, nil
		}
	}

	return data[headerLen : headerLen+length], nil
}

// WriteProperty writes a property to a BACnet object
func (c *Client) WriteProperty(ctx context.Context, deviceID uint32, objectID ObjectIdentifier, propertyID PropertyIdentifier, value interface{}, opts ...WriteOption) error {
	options := &WriteOptions{}
	for _, opt := range opts {
		opt(options)
	}

	addr, err := c.resolveDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	// Build WriteProperty request
	data := make([]byte, 0, 32)
	data = append(data, EncodeContextObjectIdentifier(0, objectID)...)
	data = append(data, EncodeContextEnumerated(1, uint32(propertyID))...)

	if options.ArrayIndex != nil {
		data = append(data, EncodeContextUnsigned(2, *options.ArrayIndex)...)
	}

	// Property value [3]
	data = append(data, EncodeOpeningTag(3)...)
	encodedValue, err := c.encodePropertyValue(value)
	if err != nil {
		return fmt.Errorf("encode value: %w", err)
	}
	data = append(data, encodedValue...)
	data = append(data, EncodeClosingTag(3)...)

	// Priority [4]
	if options.Priority != nil {
		data = append(data, EncodeContextUnsigned(4, uint32(*options.Priority))...)
	}

	_, err = c.sendRequest(ctx, addr, ServiceWriteProperty, data)
	return err
}

// encodePropertyValue encodes a property value for writing
func (c *Client) encodePropertyValue(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		return []byte{0x00}, nil
	case bool:
		return EncodeBooleanTag(v), nil
	case int:
		if v >= 0 {
			return EncodeUnsignedTag(uint32(v)), nil
		}
		data := EncodeSigned(int32(v))
		tag := EncodeTag(uint8(TagSignedInt), TagClassApplication, len(data))
		return append(tag, data...), nil
	case int32:
		if v >= 0 {
			return EncodeUnsignedTag(uint32(v)), nil
		}
		data := EncodeSigned(v)
		tag := EncodeTag(uint8(TagSignedInt), TagClassApplication, len(data))
		return append(tag, data...), nil
	case uint32:
		return EncodeUnsignedTag(v), nil
	case float32:
		return EncodeRealTag(v), nil
	case float64:
		data := EncodeDouble(v)
		tag := EncodeTag(uint8(TagDouble), TagClassApplication, len(data))
		return append(tag, data...), nil
	case string:
		return EncodeCharacterStringTag(v), nil
	case ObjectIdentifier:
		return EncodeObjectIdentifierTag(v), nil
	default:
		return nil, fmt.Errorf("unsupported value type: %T", value)
	}
}

// ReadPropertyMultiple reads multiple properties from one or more objects
func (c *Client) ReadPropertyMultiple(ctx context.Context, deviceID uint32, requests []ReadPropertyRequest) ([]PropertyValue, error) {
	addr, err := c.resolveDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	// Build ReadPropertyMultiple request
	data := make([]byte, 0, 64)

	// Group requests by object
	objectRequests := make(map[ObjectIdentifier][]ReadPropertyRequest)
	for _, req := range requests {
		objectRequests[req.ObjectID] = append(objectRequests[req.ObjectID], req)
	}

	for oid, reqs := range objectRequests {
		data = append(data, EncodeContextObjectIdentifier(0, oid)...)
		data = append(data, EncodeOpeningTag(1)...)
		for _, req := range reqs {
			data = append(data, EncodeContextEnumerated(0, uint32(req.PropertyID))...)
			if req.ArrayIndex != nil {
				data = append(data, EncodeContextUnsigned(1, *req.ArrayIndex)...)
			}
		}
		data = append(data, EncodeClosingTag(1)...)
	}

	resp, err := c.sendRequest(ctx, addr, ServiceReadPropertyMultiple, data)
	if err != nil {
		return nil, err
	}

	// Decode response
	return c.decodeReadPropertyMultipleResponse(resp.Data)
}

// decodeReadPropertyMultipleResponse decodes a ReadPropertyMultiple response
func (c *Client) decodeReadPropertyMultipleResponse(data []byte) ([]PropertyValue, error) {
	var results []PropertyValue
	offset := 0

	for offset < len(data) {
		// Object identifier [0]
		tagNum, class, length, headerLen, err := DecodeTagNumber(data[offset:])
		if err != nil {
			break
		}
		if tagNum != 0 || class != TagClassContext {
			break
		}

		oidValue := binary.BigEndian.Uint32(data[offset+headerLen:])
		oid := DecodeObjectIdentifier(oidValue)
		offset += headerLen + length

		// List of results [1]
		tagNum, class, length, _, err = DecodeTagNumber(data[offset:])
		if err != nil || tagNum != 1 || class != TagClassContext || length != -1 {
			break
		}
		offset++

		// Parse property results
		for offset < len(data) {
			tagNum, class, length, _, err = DecodeTagNumber(data[offset:])
			if err != nil {
				break
			}

			// Closing tag
			if length == -2 && tagNum == 1 {
				offset++
				break
			}

			// Property identifier [2]
			if tagNum != 2 || class != TagClassContext {
				offset++
				continue
			}
			offset += headerLen
			propID := PropertyIdentifier(DecodeUnsigned(data[offset : offset+length]))
			offset += length

			// Optional array index [3]
			var arrayIndex *uint32
			tagNum, class, length, headerLen, err = DecodeTagNumber(data[offset:])
			if err == nil && tagNum == 3 && class == TagClassContext {
				idx := DecodeUnsigned(data[offset+headerLen : offset+headerLen+length])
				arrayIndex = &idx
				offset += headerLen + length
			}

			// Property value [4] or property access error [5]
			tagNum, class, length, _, err = DecodeTagNumber(data[offset:])
			if err != nil {
				break
			}

			if tagNum == 4 && class == TagClassContext && length == -1 {
				// Property value
				offset++
				value, _ := c.decodePropertyValue(data[offset:])

				// Skip to closing tag
				for offset < len(data) {
					_, _, l, h, _ := DecodeTagNumber(data[offset:])
					offset += h
					if l == -2 {
						break
					}
					if l > 0 {
						offset += l
					}
				}

				results = append(results, PropertyValue{
					ObjectID:   oid,
					PropertyID: propID,
					ArrayIndex: arrayIndex,
					Value:      value,
				})
			} else if tagNum == 5 && class == TagClassContext && length == -1 {
				// Property access error - skip
				offset++
				for offset < len(data) {
					_, _, l, h, _ := DecodeTagNumber(data[offset:])
					offset += h
					if l == -2 {
						break
					}
					if l > 0 {
						offset += l
					}
				}
			}
		}
	}

	return results, nil
}

// SubscribeCOV subscribes to COV (Change of Value) notifications
func (c *Client) SubscribeCOV(ctx context.Context, deviceID uint32, objectID ObjectIdentifier, handler COVHandler, opts ...SubscribeOption) (uint32, error) {
	options := &SubscribeOptions{
		Confirmed: false,
	}
	for _, opt := range opts {
		opt(options)
	}

	addr, err := c.resolveDevice(ctx, deviceID)
	if err != nil {
		return 0, err
	}

	// Generate subscription ID
	subID := uint32(c.nextInvokeID())

	// Build SubscribeCOV request
	data := make([]byte, 0, 32)
	data = append(data, EncodeContextUnsigned(0, subID)...)
	data = append(data, EncodeContextObjectIdentifier(1, objectID)...)

	if options.Confirmed {
		data = append(data, EncodeContextBoolean(2, true)...)
	}

	if options.Lifetime != nil {
		data = append(data, EncodeContextUnsigned(3, *options.Lifetime)...)
	}

	_, err = c.sendRequest(ctx, addr, ServiceSubscribeCOV, data)
	if err != nil {
		return 0, err
	}

	// Register handler
	c.covMu.Lock()
	c.covSubs[subID] = handler
	c.covMu.Unlock()

	c.metrics.COVSubscriptions.Inc()

	return subID, nil
}

// UnsubscribeCOV unsubscribes from COV notifications
func (c *Client) UnsubscribeCOV(ctx context.Context, deviceID uint32, objectID ObjectIdentifier, subID uint32) error {
	addr, err := c.resolveDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	// Build SubscribeCOV request with cancel
	data := make([]byte, 0, 16)
	data = append(data, EncodeContextUnsigned(0, subID)...)
	data = append(data, EncodeContextObjectIdentifier(1, objectID)...)
	// No confirmed or lifetime = unsubscribe

	_, err = c.sendRequest(ctx, addr, ServiceSubscribeCOV, data)
	if err != nil {
		return err
	}

	// Remove handler
	c.covMu.Lock()
	delete(c.covSubs, subID)
	c.covMu.Unlock()

	return nil
}

// GetObjectList retrieves the list of objects from a device
func (c *Client) GetObjectList(ctx context.Context, deviceID uint32) ([]ObjectIdentifier, error) {
	// First, read the object-list length
	lengthVal, err := c.ReadProperty(ctx, deviceID,
		NewObjectIdentifier(ObjectTypeDevice, deviceID),
		PropertyObjectList,
		WithArrayIndex(0),
	)
	if err != nil {
		return nil, err
	}

	length, ok := lengthVal.(uint32)
	if !ok {
		return nil, fmt.Errorf("unexpected object-list length type: %T", lengthVal)
	}

	// Read each object identifier
	objects := make([]ObjectIdentifier, 0, length)
	for i := uint32(1); i <= length; i++ {
		val, err := c.ReadProperty(ctx, deviceID,
			NewObjectIdentifier(ObjectTypeDevice, deviceID),
			PropertyObjectList,
			WithArrayIndex(i),
		)
		if err != nil {
			continue
		}

		if oid, ok := val.(ObjectIdentifier); ok {
			objects = append(objects, oid)
		}
	}

	return objects, nil
}
