package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"domain-scout/internal/inspectors"
	"domain-scout/internal/ssh"
)

func TestResultSerializesSharedScanResponseAtTopLevel(t *testing.T) {
	result := Result{
		ScanResponse: ScanResponse{
			Servers: []ServerInfo{{Server: "nginx", Running: true, Domains: []DomainInfo{}}},
			System:  SystemInfo{IPv4: []string{}, IPv6: []string{}},
		},
		Status: "completed", ScannedAt: time.Unix(0, 0).UTC(),
	}
	output, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(output)
	for _, field := range []string{`"servers":[`, `"system":{`, `"status":"completed"`} {
		if !strings.Contains(encoded, field) {
			t.Errorf("JSON response missing top-level field %s: %s", field, encoded)
		}
	}
	if strings.Contains(encoded, `"ScanResponse"`) {
		t.Errorf("embedded response was not flattened: %s", encoded)
	}
}

type discoveryExecutor struct{}

func (discoveryExecutor) Run(_ context.Context, command string) (string, error) {
	switch command {
	case "hostname":
		return "node\n", nil
	case `cat /etc/os-release 2>/dev/null | sed -n 's/^PRETTY_NAME=//p' | head -n 1 | tr -d '"'`:
		return "Linux\n", nil
	case "uname -r":
		return "kernel\n", nil
	case `ip -o route get 1.1.1.1 2>/dev/null | awk '{print $7; exit}'`:
		return "192.0.2.1\n", nil
	default:
		return "", errors.New("unexpected command")
	}
}

type stoppedInspector struct{}

func (stoppedInspector) Name() string { return "apache" }
func (stoppedInspector) Detect(context.Context, ssh.Executor) (bool, error) {
	return true, nil
}
func (stoppedInspector) Running(context.Context, ssh.Executor) (bool, error) {
	return false, nil
}
func (stoppedInspector) Discover(context.Context, ssh.Executor) ([]inspectors.Domain, error) {
	return []inspectors.Domain{{
		Name: "site", Aliases: []string{"wwwsite"}, ConfigPath: "/etc/apache/site.conf",
		DocumentRoot: "/var/www/site", Ports: []int{80},
	}}, nil
}

func TestDiscoverReportsStoppedServiceAndStructuredDomains(t *testing.T) {
	result, err := NewService([]inspectors.Inspector{stoppedInspector{}}).Discover(context.Background(), discoveryExecutor{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Servers) != 1 || result.Servers[0].Running {
		t.Fatalf("expected one stopped server, got %#v", result.Servers)
	}
	if len(result.Servers[0].Domains) != 1 {
		t.Fatalf("expected one structured domain, got %#v", result.Servers[0].Domains)
	}
	domain := result.Servers[0].Domains[0]
	if domain.Domain != "site" || domain.ConfigFile != "/etc/apache/site.conf" ||
		domain.DocumentRoot != "/var/www/site" || len(domain.Aliases) != 1 || domain.Aliases[0] != "wwwsite" ||
		len(domain.Ports) != 1 || domain.Ports[0] != 80 {
		t.Errorf("unexpected domain response: %#v", domain)
	}
}
