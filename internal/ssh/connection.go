package ssh

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type ConnectionConfig struct {
	Host           string
	Port           int
	Username       string
	Auth           []gossh.AuthMethod
	KnownHostsPath string
	Timeout        time.Duration
}

func Connect(ctx context.Context, cfg ConnectionConfig) (*Client, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	hostKeyCallback, err := knownhosts.New(cfg.KnownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts: %w", err)
	}

	address := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	dialer := net.Dialer{Timeout: cfg.Timeout}
	raw, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("dial SSH: %w", err)
	}
	_ = raw.SetDeadline(time.Now().Add(cfg.Timeout))
	conn, chans, reqs, err := gossh.NewClientConn(raw, address, &gossh.ClientConfig{
		User: cfg.Username, Auth: cfg.Auth, HostKeyCallback: hostKeyCallback, Timeout: cfg.Timeout,
	})
	if err != nil {
		_ = raw.Close()
		return nil, fmt.Errorf("SSH handshake: %w", err)
	}
	_ = raw.SetDeadline(time.Time{})
	return &Client{client: gossh.NewClient(conn, chans, reqs)}, nil
}
