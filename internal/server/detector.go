package server

import (
	"context"
	"strings"

	"domain-scout/internal/ssh"
)

type Detector struct{}

func NewDetector() *Detector { return &Detector{} }

func (d *Detector) Detect(ctx context.Context, ex ssh.Executor) (Info, error) {
	var out Info
	commands := []struct {
		cmd string
		dst *string
	}{
		{"hostname", &out.Hostname},
		{"cat /etc/os-release 2>/dev/null | sed -n 's/^PRETTY_NAME=//p' | head -n 1 | tr -d '\"'", &out.OperatingSystem},
		{"uname -r", &out.Kernel},
		{"ip -o route get 1.1.1.1 2>/dev/null | awk '{print $7; exit}'", &out.PrimaryIP},
	}
	for _, item := range commands {
		value, err := ex.Run(ctx, item.cmd)
		if err != nil {
			if ctx.Err() != nil {
				return out, ctx.Err()
			}
			continue
		}
		*item.dst = strings.TrimSpace(value)
	}
	return out, nil
}
