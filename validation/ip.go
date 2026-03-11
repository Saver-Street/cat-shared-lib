package validation

import (
	"fmt"
	"net"
	"net/netip"
)

// IPAddress validates that value is a valid IPv4 or IPv6 address.
// Unlike IP which takes a single string, IPAddress follows the
// field+value pattern used by other validators in this package.
func IPAddress(field, value string) error {
	if net.ParseIP(value) == nil {
		return fmt.Errorf("%s must be a valid IP address", field)
	}
	return nil
}

// IPv6 validates that value is a valid IPv6 address (not an IPv4 address).
func IPv6(field, value string) error {
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() != nil {
		return fmt.Errorf("%s must be a valid IPv6 address", field)
	}
	return nil
}

// PrivateIP validates that value is a valid IP address within a private
// (RFC 1918/RFC 4193) range.
func PrivateIP(field, value string) error {
	ip := net.ParseIP(value)
	if ip == nil {
		return fmt.Errorf("%s must be a valid private IP address", field)
	}
	if !ip.IsPrivate() {
		return fmt.Errorf("%s must be a valid private IP address", field)
	}
	return nil
}

// IPInRange validates that value is an IP address within the given CIDR range.
func IPInRange(field, value, cidr string) error {
	ip, err := netip.ParseAddr(value)
	if err != nil {
		return fmt.Errorf("%s must be a valid IP address", field)
	}
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return fmt.Errorf("%s: invalid CIDR range %s", field, cidr)
	}
	if !prefix.Contains(ip) {
		return fmt.Errorf("%s must be within range %s", field, cidr)
	}
	return nil
}
