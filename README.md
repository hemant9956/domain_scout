# DomainScout

DomainScout connects to a Linux server over SSH, inspects supported web-server and Docker configuration, extracts configured hostnames, and checks whether DNS resolves to the server's primary outbound IP.

## Requirements

- Go 1.22+
- SSH access to the target host
- A trusted `known_hosts` file available to the API process
- Remote permissions to read the relevant service configuration

## Run

```bash
go mod tidy
go run ./cmd/api
```

The API listens on `:8080` by default.

## Endpoints

- `GET /api/v1/health`
- `POST /api/v1/servers/test-connection`
- `POST /api/v1/discoveries`

## Discovery request

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

The `known_hosts` path is a path on the machine running DomainScout, not on the remote server. Populate it using a trusted out-of-band host-key verification process. Do not blindly trust `ssh-keyscan` output from an untrusted network.

## Important limitations

- Discovery is configuration-based. A hostname in a web-server config is not proof that the site is publicly reachable.
- DNS matching compares results with the remote server's primary outbound interface IP. Servers behind NAT, load balancers, CDNs, or multiple public IPs may need a configurable expected-IP list.
- Apache discovery currently uses `apachectl -S` / `httpd -S` output. It is a starting point; additional config parsing should be added for aliases, document roots, and SSL.
- Caddy parsing is intentionally basic and should be extended for imported snippets and JSON configuration.
- Nginx `nginx -T` may require elevated permissions. The implementation tries `sudo -n` and then a non-sudo command.
- Docker label discovery supports common hostname/rule labels but is not a complete parser for every proxy.
- This first version is stateless and does not store credentials, servers, or discovery history.

## Security

- SSH host-key verification is mandatory.
- Do not expose this API publicly without authentication, rate limiting, audit logging, and network restrictions.
- Never log SSH passwords, private keys, or passphrases.
- Prefer a dedicated least-privilege SSH user and narrowly scoped sudo permissions.
- Treat SSH command output as untrusted input.
