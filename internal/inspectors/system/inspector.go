package system

import (
	"context"
	"strconv"
	"strings"

	"domain-scout/internal/inspectors"
	"domain-scout/internal/server"
	"domain-scout/internal/ssh"
)

type Inspector struct{}

func New() *Inspector                                                   { return &Inspector{} }
func (i *Inspector) Name() string                                       { return "system" }
func (i *Inspector) Detect(context.Context, ssh.Executor) (bool, error) { return true, nil }
func (i *Inspector) Discover(ctx context.Context, ex ssh.Executor) ([]inspectors.Domain, error) {
	return []inspectors.Domain{}, nil
}

func (i *Inspector) Collect(ctx context.Context, ex ssh.Executor) (server.SystemInfo, []string, error) {
	info := server.SystemInfo{IPv4: []string{}, IPv6: []string{}}
	output, err := ex.Run(ctx, collectScript)
	if err != nil {
		return info, nil, err
	}

	values := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if ok {
			values[key] = value
		}
	}
	info.Hostname = values["HOSTNAME"]
	info.IPv4 = splitAddresses(values["IPV4"])
	info.IPv6 = splitAddresses(values["IPV6"])
	info.MAC = values["MAC"]
	info.ServerOS = values["SERVER_OS"]
	info.OSVersion = values["OS_VERSION"]
	info.Environment = values["ENVIRONMENT"]
	info.Architecture = values["ARCHITECTURE"]
	info.CPUModel = values["CPU_MODEL"]
	info.Kernel = values["KERNEL"]
	info.CurrentUser = values["CURRENT_USER"]
	info.Timezone = values["TIMEZONE"]
	info.Cores = parseInt(values["CORES"])
	info.Threads = parseInt(values["THREADS"])
	info.Uptime = parseUint(values["UPTIME"])
	info.Disk = parseDisk(values["DISK"])
	info.Memory = parseMemory(values["MEMORY"])

	var unavailable []string
	for _, field := range []struct {
		name  string
		value string
	}{
		{"hostname", info.Hostname}, {"server_os", info.ServerOS}, {"os_version", info.OSVersion},
		{"architecture", info.Architecture}, {"cpu_model", info.CPUModel}, {"kernel", info.Kernel},
		{"current_user", info.CurrentUser}, {"timezone", info.Timezone}, {"mac", info.MAC},
	} {
		if field.value == "" {
			unavailable = append(unavailable, field.name)
		}
	}
	if info.Cores == 0 {
		unavailable = append(unavailable, "cores")
	}
	if info.Threads == 0 {
		unavailable = append(unavailable, "threads")
	}
	if info.Uptime == 0 {
		unavailable = append(unavailable, "uptime")
	}
	if len(info.IPv4) == 0 {
		unavailable = append(unavailable, "ipv4")
	}
	if len(info.IPv6) == 0 {
		unavailable = append(unavailable, "ipv6")
	}
	if values["DISK"] == "" {
		unavailable = append(unavailable, "disk")
	}
	if values["MEMORY"] == "" {
		unavailable = append(unavailable, "memory")
	}
	var warnings []string
	if len(unavailable) > 0 {
		warnings = append(warnings, "system information unavailable for: "+strings.Join(unavailable, ", "))
	}
	return info, warnings, nil
}

func splitAddresses(value string) []string {
	addresses := []string{}
	seen := map[string]bool{}
	for _, address := range strings.Split(value, ",") {
		address = strings.TrimSpace(address)
		if address != "" && !seen[address] {
			addresses = append(addresses, address)
			seen[address] = true
		}
	}
	return addresses
}

func parseInt(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func parseUint(value string) uint64 {
	parsed, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return parsed
}

func parseDisk(value string) server.DiskInfo {
	var disk server.DiskInfo
	parts := strings.Split(value, ",")
	if len(parts) != 3 {
		return disk
	}
	disk.TotalGB, _ = strconv.ParseFloat(parts[0], 64)
	disk.FreeGB, _ = strconv.ParseFloat(parts[1], 64)
	disk.UsedGB, _ = strconv.ParseFloat(parts[2], 64)
	return disk
}

func parseMemory(value string) server.MemoryInfo {
	var memory server.MemoryInfo
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return memory
	}
	memory.TotalGB, _ = strconv.ParseFloat(parts[0], 64)
	memory.AvailableGB, _ = strconv.ParseFloat(parts[1], 64)
	memory.FreeGB, _ = strconv.ParseFloat(parts[2], 64)
	memory.UsedGB, _ = strconv.ParseFloat(parts[3], 64)
	return memory
}

const collectScript = `hostname=$(hostname 2>/dev/null)
server_os=$(sed -n 's/^PRETTY_NAME=//p' /etc/os-release 2>/dev/null | head -n 1 | tr -d '"')
os_version=$(sed -n 's/^VERSION_ID=//p' /etc/os-release 2>/dev/null | head -n 1 | tr -d '"')
ipv4=$(ip -o -4 addr show 2>/dev/null | awk '{sub(/\/.*/, "", $4); printf "%s%s", sep, $4; sep=","}')
ipv6=$(ip -o -6 addr show 2>/dev/null | awk '{sub(/\/.*/, "", $4); printf "%s%s", sep, $4; sep=","}')
mac=$(ip -o link show 2>/dev/null | awk '/link\/ether/ {for (i=1; i<=NF; i++) if ($i == "link/ether") {print $(i+1); exit}}')
architecture=$(uname -m 2>/dev/null)
case "$architecture" in
  x86_64) architecture=amd64 ;;
  aarch64) architecture=arm64 ;;
  armv7*) architecture=arm ;;
esac
cores=$(lscpu -p=CORE,SOCKET 2>/dev/null | awk -F, '!/^#/ && NF > 1 {seen[$1 FS $2]=1} END {n=0; for (key in seen) n++; print n}')
threads=$(getconf _NPROCESSORS_ONLN 2>/dev/null)
if [ -z "$threads" ] || [ "$threads" -eq 0 ] 2>/dev/null; then threads=$(nproc 2>/dev/null); fi
if [ -z "$cores" ] || [ "$cores" -eq 0 ]; then cores=$threads; fi
cpu_model=$(awk -F: '/^model name[[:space:]]*:/ {sub(/^[[:space:]]*/, "", $2); print $2; exit}' /proc/cpuinfo 2>/dev/null)
kernel=$(uname -r 2>/dev/null)
uptime=$(awk '{printf "%d", $1}' /proc/uptime 2>/dev/null)
current_user=$(id -un 2>/dev/null)
timezone=$(date +%Z 2>/dev/null)
if [ -f /.dockerenv ] || grep -Eqi 'docker|containerd|kubepods|lxc' /proc/1/cgroup /proc/1/mountinfo 2>/dev/null; then environment=Container; else environment=Host; fi
disk=$(df -B1P / 2>/dev/null | awk 'NR == 2 {printf "%.2f,%.2f,%.2f", $2/1073741824, $4/1073741824, $3/1073741824}')
memory=$(awk '/^MemTotal:/ {total=$2} /^MemAvailable:/ {available=$2} /^MemFree:/ {free=$2} END {if (total > 0) printf "%.2f,%.2f,%.2f,%.2f", total/1048576, available/1048576, free/1048576, (total-available)/1048576}' /proc/meminfo 2>/dev/null)
printf 'HOSTNAME=%s\nIPV4=%s\nIPV6=%s\nMAC=%s\nSERVER_OS=%s\nOS_VERSION=%s\nENVIRONMENT=%s\nARCHITECTURE=%s\nCORES=%s\nTHREADS=%s\nCPU_MODEL=%s\nKERNEL=%s\nUPTIME=%s\nCURRENT_USER=%s\nTIMEZONE=%s\nDISK=%s\nMEMORY=%s\n' "$hostname" "$ipv4" "$ipv6" "$mac" "$server_os" "$os_version" "$environment" "$architecture" "$cores" "$threads" "$cpu_model" "$kernel" "$uptime" "$current_user" "$timezone" "$disk" "$memory"`
