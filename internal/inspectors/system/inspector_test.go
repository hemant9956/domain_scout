package system

import (
	"context"
	"reflect"
	"testing"
)

type resultExecutor struct {
	output string
}

func (e resultExecutor) Run(context.Context, string) (string, error) {
	return e.output, nil
}

func TestCollectParsesStructuredSystemInformation(t *testing.T) {
	executor := resultExecutor{output: `HOSTNAME=Tirth
IPV4=127.0.0.1,172.23.95.136
IPV6=::1,fe80::215:5dff:febe:f5c0
MAC=00:15:5d:be:f5:c0
SERVER_OS=Ubuntu 26.04.1 LTS
OS_VERSION=26.04
ENVIRONMENT=Host
ARCHITECTURE=amd64
CORES=6
THREADS=6
CPU_MODEL=Intel(R) Core(TM) i5-9400F CPU @ 2.90GHz
KERNEL=6.18.33.2-microsoft-standard-WSL2
UPTIME=23100
CURRENT_USER=parth
TIMEZONE=UTC
DISK=1006.85,953.29,53.56
MEMORY=3.80,3.26,3.00,0.54
`}

	info, warnings, err := New().Collect(context.Background(), executor)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if info.Hostname != "Tirth" || info.Cores != 6 || info.Threads != 6 {
		t.Errorf("unexpected identity or CPU data: %#v", info)
	}
	if !reflect.DeepEqual(info.IPv4, []string{"127.0.0.1", "172.23.95.136"}) {
		t.Errorf("IPv4 = %#v", info.IPv4)
	}
	if !reflect.DeepEqual(info.IPv6, []string{"::1", "fe80::215:5dff:febe:f5c0"}) {
		t.Errorf("IPv6 = %#v", info.IPv6)
	}
	if info.Disk.TotalGB != 1006.85 || info.Memory.AvailableGB != 3.26 {
		t.Errorf("unexpected resource data: disk=%#v memory=%#v", info.Disk, info.Memory)
	}
}

func TestCollectKeepsUnavailableListsEmptyAndWarns(t *testing.T) {
	info, warnings, err := New().Collect(context.Background(), resultExecutor{})
	if err != nil {
		t.Fatal(err)
	}
	if info.IPv4 == nil || info.IPv6 == nil {
		t.Fatalf("missing addresses must be empty slices, got ipv4=%#v ipv6=%#v", info.IPv4, info.IPv6)
	}
	if len(warnings) == 0 {
		t.Fatal("expected a warning for unavailable system metrics")
	}
}
