package security

import (
	"fmt"
	"net"
	"time"
)

// ClamAVClient is a client for ClamAV antivirus scanner
type ClamAVClient struct {
	enabled bool
	address string
	timeout time.Duration
}

// NewClamAVClient creates a new ClamAV client
func NewClamAVClient(enabled bool, address string) *ClamAVClient {
	return &ClamAVClient{
		enabled: enabled,
		address: address,
		timeout: 30 * time.Second,
	}
}

// Scan scans data using ClamAV
func (c *ClamAVClient) Scan(data []byte) (bool, string, error) {
	if !c.enabled {
		return true, "", nil
	}

	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return false, "", fmt.Errorf("failed to connect to ClamAV: %w", err)
	}
	defer conn.Close()

	// Set deadline
	conn.SetDeadline(time.Now().Add(c.timeout))

	// Send INSTREAM command
	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return false, "", fmt.Errorf("failed to send command: %w", err)
	}

	// Send data size (4 bytes, network order)
	size := uint32(len(data))
	sizeBytes := []byte{
		byte(size >> 24),
		byte(size >> 16),
		byte(size >> 8),
		byte(size),
	}
	if _, err := conn.Write(sizeBytes); err != nil {
		return false, "", fmt.Errorf("failed to send size: %w", err)
	}

	// Send data
	if _, err := conn.Write(data); err != nil {
		return false, "", fmt.Errorf("failed to send data: %w", err)
	}

	// Send zero-length chunk to signal end
	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return false, "", fmt.Errorf("failed to send end marker: %w", err)
	}

	// Read response
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		return false, "", fmt.Errorf("failed to read response: %w", err)
	}

	result := string(response[:n])

	// Check result
	if result == "stream: OK\x00" {
		return true, result, nil
	}

	return false, result, nil
}

// Ping checks if ClamAV is available
func (c *ClamAVClient) Ping() error {
	if !c.enabled {
		return nil
	}

	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to ClamAV: %w", err)
	}
	defer conn.Close()

	// Send PING command
	if _, err := conn.Write([]byte("zPING\x00")); err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	// Read response
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if string(response[:n]) != "PONG\x00" {
		return fmt.Errorf("unexpected response: %s", string(response[:n]))
	}

	return nil
}
