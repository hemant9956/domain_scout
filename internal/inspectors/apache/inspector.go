package apache

import (
	"context"
	"domain-scout/internal/inspectors"
	"domain-scout/internal/ssh"
)

type Inspector struct{}

func New() *Inspector             { return &Inspector{} }
func (i *Inspector) Name() string { return "apache" }
func (i *Inspector) Detect(ctx context.Context, ex ssh.Executor) (bool, error) {
	_, e1 := ex.Run(ctx, "command -v apachectl")
	if e1 == nil {
		return true, nil
	}
	_, e2 := ex.Run(ctx, "command -v httpd")
	return e2 == nil, nil
}
func (i *Inspector) Discover(ctx context.Context, ex ssh.Executor) ([]inspectors.Domain, error) {
	out, err := ex.Run(ctx, "sudo -n apachectl -S 2>&1")
	if err != nil {
		out, err = ex.Run(ctx, "sudo -n httpd -S 2>&1")
	}
	if err != nil {
		return nil, err
	}
	return Parse(out), nil
}
