package inspectors

import (
	"context"
	"domain-scout/internal/ssh"
)

type Domain struct {
	Name         string   `json:"name"`
	Source       string   `json:"source"`
	ConfigPath   string   `json:"config_path,omitempty"`
	DocumentRoot string   `json:"document_root,omitempty"`
	ProxyTarget  string   `json:"proxy_target,omitempty"`
	SSL          bool     `json:"ssl"`
	Aliases      []string `json:"aliases,omitempty"`
	Ports        []int    `json:"ports,omitempty"`
}

type Inspector interface {
	Name() string
	Detect(context.Context, ssh.Executor) (bool, error)
	Discover(context.Context, ssh.Executor) ([]Domain, error)
}
