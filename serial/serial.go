package serial

import (
	"time"

	"go.bug.st/serial"
)

const (
	defaultBaudRate  = 9600
	defaultReadSize  = 32
	byteWriteDelayMs = 100
)

// Port wraps the serial port for SMS UPS communication
type Port struct {
	port       serial.Port
	readTimeout time.Duration
}

// Open tries each port in the list until one succeeds (uses 9600 baud)
func Open(ports []string) (*Port, error) {
	return OpenWithBaudAndTimeout(ports, defaultBaudRate, 2*time.Second)
}

// OpenWithBaud opens with custom baud rate (for Manager III: use 2400)
func OpenWithBaud(ports []string, baudRate int) (*Port, error) {
	return OpenWithBaudAndTimeout(ports, baudRate, 2*time.Second)
}

// OpenWithTimeout opens serial port with custom read timeout (for diagnostics)
func OpenWithTimeout(ports []string, readTimeout time.Duration) (*Port, error) {
	return OpenWithBaudAndTimeout(ports, defaultBaudRate, readTimeout)
}

// OpenWithBaudAndTimeout opens with custom baud rate and read timeout (for diagnostics)
func OpenWithBaudAndTimeout(ports []string, baudRate int, readTimeout time.Duration) (*Port, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	if readTimeout <= 0 {
		readTimeout = 2 * time.Second
	}

	var lastErr error
	for _, name := range ports {
		p, err := serial.Open(name, mode)
		if err != nil {
			lastErr = err
			continue
		}

		if err := p.SetReadTimeout(readTimeout); err != nil {
			p.Close()
			lastErr = err
			continue
		}

		return &Port{port: p, readTimeout: readTimeout}, nil
	}

	return nil, lastErr
}

// SendCommand writes cmd bytes (with 100ms delay between each, as in Python)
// and reads up to 32 bytes of response
func (p *Port) SendCommand(cmd []byte) ([]byte, error) {
	// Flush any stale data from previous reads
	_ = p.port.ResetInputBuffer()

	for i, b := range cmd {
		_, err := p.port.Write([]byte{b})
		if err != nil {
			return nil, err
		}
		if i < len(cmd)-1 {
			time.Sleep(byteWriteDelayMs * time.Millisecond)
		}
	}

	// UPS needs time to process the command before responding
	time.Sleep(300 * time.Millisecond)

	buf := make([]byte, defaultReadSize)
	var total []byte
	for i := 0; i < 5; i++ {
		n, err := p.port.Read(buf)
		if err != nil {
			return total, err
		}
		if n > 0 {
			total = append(total, buf[:n]...)
			if len(total) >= 18 {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return total, nil
}

// Wake pulses DTR to wake some UPS models (Cypress USB adapter)
func (p *Port) Wake() {
	if p.port == nil {
		return
	}
	p.port.SetDTR(false)
	time.Sleep(100 * time.Millisecond)
	p.port.SetDTR(true)
	time.Sleep(200 * time.Millisecond)
}

// Close closes the serial port
func (p *Port) Close() error {
	if p.port != nil {
		return p.port.Close()
	}
	return nil
}
