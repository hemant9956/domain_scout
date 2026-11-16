package validator

import (
	"fmt"
	"net"
	"strings"
)

func ValidateSSH(host string, port int, username, authType, knownHosts string) error {
	if strings.TrimSpace(host) == "" || net.ParseIP(host) == nil {
		// Hostnames are accepted too, but whitespace and URL syntax are not.
		if strings.TrimSpace(host) == "" || strings.ContainsAny(host, " /\\:@") {
			return fmt.Errorf("host must be an IP address or hostname")
		}
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("username is required")
	}
	if authType != "password" && authType != "private_key" {
		return fmt.Errorf("auth_type must be password or private_key")
	}
	if strings.TrimSpace(knownHosts) == "" {
		return fmt.Errorf("known_hosts path is required for SSH host verification")
	}
	return nil
}
