# DomainScout

DomainScout has two execution modes that share the same discovery engine and JSON result model:

1. **API mode** — connect to another Linux server over SSH using the HTTP API.
2. **Agent mode** — run a standalone binary directly on the Linux server being inspected. No SSH connection or Go installation is required on that server.

Discovery is configuration-based: it inspects Nginx, Apache, Caddy and Docker, then optionally verifies discovered names through DNS. A configured hostname is not proof that a website is publicly reachable.

## Requirements

- Go 1.22+ to build from source.
- Agent runtime: Linux with `/bin/sh` and the relevant service configuration/tools installed.
- Root or appropriate read permissions are recommended. Some inspectors attempt `sudo -n` for privileged configuration output.
- DNS verification requires network/DNS access. The scan still reports discovered configuration when DNS lookup fails.

## Build the standalone agent

Build on your development machine for the target server architecture:

```bash
# Linux x86_64 / AMD64
make build-linux-amd64

# Linux ARM64
make build-linux-arm64

# Build both Linux architectures
./scripts/build.sh
```

The binaries are written to `dist/`. The target server does not need Go installed.

Copy the correct binary to the target server, then run:

```bash
chmod +x domain-scout-linux-amd64
sudo ./domain-scout-linux-amd64 scan --json
```

Other commands:

```bash
# Compact JSON
sudo ./domain-scout-linux-amd64 scan --compact

# Save the JSON response
sudo ./domain-scout-linux-amd64 scan --json > domains.json

# Set a scan timeout
sudo ./domain-scout-linux-amd64 scan --timeout 3m

# Print version
./domain-scout-linux-amd64 version
```

`scan --json` writes the discovery result to stdout. Diagnostics and usage information go to stderr. On a fatal scan error, stdout contains an API-style JSON error object and the process exits with status 1. Invalid command-line usage exits with status 2.

### Agent JSON response

The agent serializes the same `discovery.Result` model returned by `POST /api/v1/discoveries`. In addition to the existing `status`, `scanned_at`, `server`, `detected_services`, `total_domains`, `domains`, and optional `warnings`, the response includes:

- `servers`: installed web/container services, their actual running state, and structured domains with aliases, config files, document roots, and listening ports.
- `system`: hostname, IP addresses, MAC, OS, CPU, uptime, user, timezone, disk, and memory information.

The added fields are top-level and additive, so existing API response fields remain available. CPU `cores` counts physical cores when `lscpu` is available; `threads` counts online logical CPUs. Disk values describe the root filesystem. Disk and memory values labeled `*_gb` use GiB (bytes divided by 1024^3); unavailable metrics remain zero/empty and are described in `warnings`. Missing IP or domain lists are returned as empty arrays.

## Run the HTTP API

```bash
go mod tidy
go run ./cmd/api
```

The API listens on `:8080` by default.

### Endpoints

- `GET /api/v1/health`
- `POST /api/v1/servers/test-connection`
- `POST /api/v1/discoveries`

Example discovery request:

```json
{
  "host": "203.0.113.10",
  "port": 22,
  "username": "deploy",
  "auth_type": "private_key",
  "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\\n...\\n-----END OPENSSH PRIVATE KEY-----",
  "known_hosts": "/app/ssh/known_hosts"
}
```

For password authentication, send `auth_type: "password"` and `password` instead.

The `known_hosts` path is on the machine running the API, not on the remote server. Populate it using a trusted out-of-band host-key verification process. Do not blindly trust `ssh-keyscan` output from an untrusted network.

## Tests and quality checks

```bash
gofmt -w cmd internal pkg
go test ./...
go vet ./...
```

## Current limitations

- The agent currently targets Linux. It uses `/bin/sh` to run the same inspection commands as the SSH mode.
- Apache discovery uses `apachectl -S` / `httpd -S` output and is not a complete parser for every virtual-host directive.
- Caddy parsing is basic and does not fully resolve imported snippets or every JSON configuration form.
- Nginx `nginx -T` may require elevated permissions.
- Docker label discovery supports common hostname/rule labels but not every reverse-proxy convention.
- DNS matching compares returned addresses with the server's detected primary interface IP. NAT, load balancers, CDNs, and multiple public IPs can affect this comparison.
- The first version is stateless and does not store credentials, servers, or discovery history.

## Security

- Run the agent with the least privilege that still permits reading the relevant service configurations. Root may be needed for complete inspection on some hosts.
- The agent executes a fixed set of read-only inspection commands through `/bin/sh`; do not modify it to accept arbitrary shell commands from untrusted input.
- Do not expose the API publicly without authentication, rate limiting, audit logging, and network restrictions.
- SSH host-key verification is mandatory in API mode.
- Never log SSH passwords, private keys, or passphrases.
