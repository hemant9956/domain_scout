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
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	return err == nil, nil
}
func (i *Inspector) Running(ctx context.Context, ex ssh.Executor) (bool, error) {
	if _, err := ex.Run(ctx, "systemctl is-active --quiet caddy"); err == nil {
		return true, nil
	}
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if _, err := ex.Run(ctx, "ps -eo comm= | grep -Fxq caddy"); err == nil {
		return true, nil
	}
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	return false, nil
}
func (i *Inspector) Discover(ctx context.Context, ex ssh.Executor) ([]inspectors.Domain, error) {
	out, err := ex.Run(ctx, "cat /etc/caddy/Caddyfile")
	if err != nil {
		out, err = ex.Run(ctx, "sudo -n cat /etc/caddy/Caddyfile")
	}
	if err != nil {
		return nil, err
	}
	return Parse(out), nil
}
