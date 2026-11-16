package docker

import (
	"context"
	"domain-scout/internal/inspectors"
	"domain-scout/internal/ssh"
	"encoding/json"
	"strings"
)

type Inspector struct{}

func New() *Inspector             { return &Inspector{} }
func (i *Inspector) Name() string { return "docker" }
func (i *Inspector) Detect(ctx context.Context, ex ssh.Executor) (bool, error) {
	_, err := ex.Run(ctx, "docker info --format '{{.ServerVersion}}'")
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	return err == nil, nil
}
func (i *Inspector) Running(context.Context, ssh.Executor) (bool, error) {
	return true, nil
}
func (i *Inspector) Discover(ctx context.Context, ex ssh.Executor) ([]inspectors.Domain, error) {
	out, err := ex.Run(ctx, "docker ps --format '{{json .}}'")
	if err != nil {
		return nil, err
	}
	var result []inspectors.Domain
	for _, line := range strings.Split(out, "\n") {
		var row struct {
			Names  string `json:"Names"`
			Labels string `json:"Labels"`
		}
		if json.Unmarshal([]byte(line), &row) != nil {
			continue
		}
		for _, label := range strings.Split(row.Labels, ",") {
			parts := strings.SplitN(label, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := parts[0]
			if strings.Contains(key, ".rule") || strings.Contains(key, ".host") {
				value := strings.Trim(parts[1], "`\" ")
				for _, host := range strings.FieldsFunc(value, func(r rune) bool { return r == ' ' || r == '`' || r == '(' || r == ')' || r == ',' || r == '"' }) {
					if strings.Contains(host, ".") {
						result = append(result, inspectors.Domain{Name: host, Source: "docker", ProxyTarget: row.Names})
					}
				}
			}
		}
	}
	return result, nil
}
