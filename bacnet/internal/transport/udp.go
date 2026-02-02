// Package transport provides the transport layer for BACnet communication
package transport

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// UDPTransport implements BACnet/IP transport over UDP
type UDPTransport struct {
	localAddr    string
	conn         *net.UDPConn
	mu           sync.RWMutex
	readTimeout  time.Duration
	writeTimeout time.Duration
	closed       bool
}

// NewUDPTransport creates a new UDP transport
func NewUDPTransport(localAddr string) *UDPTransport {
	return &UDPTransport{
		localAddr:    localAddr,
		readTimeout:  3 * time.Second,
		writeTimeout: 3 * time.Second,
	}
}

// SetReadTimeout sets the read timeout
func (t *UDPTransport) SetReadTimeout(d time.Duration) {
	t.mu.Lock()
	t.readTimeout = d
	t.mu.Unlock()
}

// SetWriteTimeout sets the write timeout
func (t *UDPTransport) SetWriteTimeout(d time.Duration) {
	t.mu.Lock()
	t.writeTimeout = d
	t.mu.Unlock()
}

// Open opens the UDP connection
func (t *UDPTransport) Open(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn != nil {
		return nil
	}

	var addr *net.UDPAddr
	var err error

	if t.localAddr != "" {
		addr, err = net.ResolveUDPAddr("udp4", t.localAddr)
		if err != nil {
			return fmt.Errorf("resolve local address: %w", err)
		}
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("listen UDP: %w", err)
	}

	t.conn = conn
	t.closed = false
	return nil
}

// Close closes the UDP connection
func (t *UDPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn == nil || t.closed {
		return nil
	}

	t.closed = true
	return t.conn.Close()
}

// LocalAddr returns the local address
func (t *UDPTransport) LocalAddr() net.Addr {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.conn == nil {
		return nil
	}
	return t.conn.LocalAddr()
}

// Send sends data to a specific address
func (t *UDPTransport) Send(ctx context.Context, addr *net.UDPAddr, data []byte) error {
	t.mu.RLock()
	conn := t.conn
	writeTimeout := t.writeTimeout
	t.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("transport not open")
	}

	// Set deadline from context or default timeout
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(writeTimeout)
	}
	if err := conn.SetWriteDeadline(deadline); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	n, err := conn.WriteToUDP(data, addr)
	if err != nil {
		return fmt.Errorf("write UDP: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("partial write: %d of %d bytes", n, len(data))
	}

	return nil
}

// Broadcast sends data to the broadcast address
func (t *UDPTransport) Broadcast(ctx context.Context, port int, data []byte) error {
	addr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: port,
	}
	return t.Send(ctx, addr, data)
}

// Receive receives data from the transport
func (t *UDPTransport) Receive(ctx context.Context) ([]byte, *net.UDPAddr, error) {
	t.mu.RLock()
	conn := t.conn
	readTimeout := t.readTimeout
	t.mu.RUnlock()

	if conn == nil {
		return nil, nil, fmt.Errorf("transport not open")
	}

	// Set deadline from context or default timeout
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(readTimeout)
	}
	if err := conn.SetReadDeadline(deadline); err != nil {
		return nil, nil, fmt.Errorf("set read deadline: %w", err)
	}

	buf := make([]byte, 1500) // MTU size
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, nil, err
	}

	return buf[:n], addr, nil
}

// ReceiveWithTimeout receives data with a specific timeout
func (t *UDPTransport) ReceiveWithTimeout(timeout time.Duration) ([]byte, *net.UDPAddr, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return t.Receive(ctx)
}

// IsClosed returns true if the transport is closed
func (t *UDPTransport) IsClosed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.closed
}
