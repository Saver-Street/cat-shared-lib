package validation

import (
	"net"
	"strconv"
)

// PortNumber validates that value is a valid TCP/UDP port number (1–65535).
func PortNumber(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "port number is required"}
	}
	port, err := strconv.Atoi(value)
	if err != nil {
		return &ValidationError{Field: field, Message: "port must be a number"}
	}
	if port < 1 || port > 65535 {
		return &ValidationError{Field: field, Message: "port must be between 1 and 65535"}
	}
	return nil
}

// HostPort validates that value is a valid host:port combination.
// The host may be a hostname, IPv4, or bracketed IPv6 address.
func HostPort(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "host:port is required"}
	}
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return &ValidationError{Field: field, Message: "invalid host:port format"}
	}
	if host == "" {
		return &ValidationError{Field: field, Message: "host is required"}
	}
	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return &ValidationError{Field: field, Message: "port must be between 1 and 65535"}
	}
	return nil
}
