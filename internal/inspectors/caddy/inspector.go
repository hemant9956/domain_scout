package caddy

import (
	"context"
	"domain-scout/internal/inspectors"
	"domain-scout/internal/ssh"
)

type Inspector struct{}

func New() *Inspector             { return &Inspector{} }
func (i *Inspector) Name() string { return "caddy" }
func (i *Inspector) Detect(ctx context.Context, ex ssh.Executor) (bool, error) {
	_, err := ex.Run(ctx, "command -v caddy")
	return err == nil, nil
}
func (i *Inspector) Discover(ctx context.Context, ex ssh.Executor) ([]inspectors.Domain, error) {
	out, err := ex.Run(ctx, "sudo -n caddy adapt --config /etc/caddy/Caddyfile --adapter caddyfile --pretty 2>/dev/null")
	if err != nil {
		out, err = ex.Run(ctx, "cat /etc/caddy/Caddyfile")
	}
	if err != nil {
		return nil, err
	}
	return Parse(out), nil
}
