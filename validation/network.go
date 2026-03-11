package validation

import (
	"fmt"
	"strconv"
)

// Port validates that value is a valid TCP/UDP port number (1-65535).
func Port(field, value string) error {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("%s must be a valid port number (1-65535)", field)
	}
	return nil
}

// PortInt validates that n is a valid TCP/UDP port number (1-65535).
func PortInt(field string, n int) error {
	if n < 1 || n > 65535 {
		return fmt.Errorf("%s must be a valid port number (1-65535)", field)
	}
	return nil
}

// HostPort validates that value is a valid host:port combination.
func HostPort(field, value string) error {
	if len(value) == 0 {
		return fmt.Errorf("%s must be a valid host:port", field)
	}
	// Find the last colon to split host and port
	lastColon := -1
	for i := len(value) - 1; i >= 0; i-- {
		if value[i] == ':' {
			lastColon = i
			break
		}
	}
	if lastColon < 1 || lastColon >= len(value)-1 {
		return fmt.Errorf("%s must be a valid host:port", field)
	}
	host := value[:lastColon]
	port := value[lastColon+1:]

	if err := Hostname(field, host); err != nil {
		return fmt.Errorf("%s must be a valid host:port", field)
	}
	if err := Port(field, port); err != nil {
		return fmt.Errorf("%s must be a valid host:port", field)
	}
	return nil
}
